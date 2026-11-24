package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

type partyRig struct {
	t *testing.T
	g *game.Game
	w *sim.World
}

func newPartyRig(t *testing.T) *partyRig {
	t.Helper()
	g, w, _ := gametest.New(t, 1, nil)
	return &partyRig{t: t, g: g, w: w}
}

func (r *partyRig) member(template api.ContentID) ecs.Entity {
	r.t.Helper()
	query := ecs.NewFilter1[comp.Identity](r.w.ECS).Query()
	for query.Next() {
		if query.Get().Template == template {
			entity := query.Entity()
			query.Close()
			return entity
		}
	}
	r.t.Fatalf("no body of %s", template)
	return ecs.Entity{}
}

func (r *partyRig) level(body ecs.Entity, need api.ContentID) float64 {
	r.t.Helper()
	for _, held := range ecs.NewMap[comp.Needs](r.w.ECS).Get(body).Levels {
		if held.Need == need {
			return held.Level
		}
	}
	r.t.Fatalf("no need %s", need)
	return 0
}

func (r *partyRig) days(n int) {
	r.t.Helper()
	for range n * 4 {
		if _, err := r.g.Apply(api.Wait{Source: r.g.Current(), Until: api.WaitPhase}); err != nil {
			r.t.Fatal(err)
		}
	}
}

func (r *partyRig) slowed(body ecs.Entity) bool {
	statuses := ecs.NewMap[comp.Statuses](r.w.ECS)
	if !statuses.Has(body) {
		return false
	}
	for _, entry := range statuses.Get(body).Active {
		if entry.Status == "status.slowed" && entry.Until > r.w.Now {
			return true
		}
	}
	return false
}

func (r *partyRig) torso(body ecs.Entity) int {
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, part := range anatomy.Parts(r.w, body) {
		if parts.Get(part).Slot == "torso" {
			return parts.Get(part).Integrity
		}
	}
	r.t.Fatal("no torso")
	return 0
}

func (r *partyRig) strikeWindup(actor, target ecs.Entity) int64 {
	r.t.Helper()
	var torso ecs.Entity
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, part := range anatomy.Parts(r.w, target) {
		if parts.Get(part).Slot == "torso" {
			torso = part
		}
	}
	windup := actions.Windup(r.w, actions.Intent{
		Actor: actor, Action: "action.strike", Target: target, TargetPart: torso,
		Choices: map[string]string{"type": "type.cut"},
	})
	if windup == 0 {
		r.t.Fatal("the strike could not be worked out")
	}
	return windup
}

func TestAMemberWhoGoesWithoutIsSlowedThenWeakened(t *testing.T) {
	r := newPartyRig(t)
	reaper, salken := r.member("template.person.reaper"), r.member("template.person.salken")
	before := r.strikeWindup(reaper, salken)
	startTorso := r.torso(reaper)

	r.days(1)
	if r.slowed(reaper) || r.torso(reaper) != startTorso {
		t.Fatalf("after one day: slowed=%v torso=%d, want neither", r.slowed(reaper), r.torso(reaper))
	}
	r.days(2)
	if !r.slowed(reaper) {
		t.Errorf("after three days the Reaper is not slowed")
	}
	if got := r.strikeWindup(reaper, salken); got <= before {
		t.Errorf("slowed wind-up = %d, want more than %d", got, before)
	}
	if r.torso(reaper) != startTorso {
		t.Errorf("torso = %d after three days, want %d", r.torso(reaper), startTorso)
	}
	r.days(1)
	if r.torso(reaper) >= startTorso {
		t.Errorf("torso = %d after four days, want less than %d", r.torso(reaper), startTorso)
	}
}

func TestDrinkingSpendsLakeWaterAndLowersThirst(t *testing.T) {
	r := newPartyRig(t)
	reaper := r.member("template.person.reaper")
	r.days(1)
	thirst := r.level(reaper, "need.water")
	water := inventory.Count(r.w, inventory.HolderOf(r.w, reaper), "item.lake_water", 0)

	view, err := r.g.View()
	if err != nil {
		t.Fatal(err)
	}
	for _, option := range view.Options {
		if do, ok := option.Command.(api.Do); ok && do.Action == "action.drink" {
			if _, err := r.g.Apply(option.Command); err != nil {
				t.Fatal(err)
			}
			if got := r.level(reaper, "need.water"); got >= thirst {
				t.Errorf("thirst = %v, want below %v", got, thirst)
			}
			if got := inventory.Count(r.w, inventory.HolderOf(r.w, reaper), "item.lake_water", 0); got != water-1 {
				t.Errorf("lake water = %d, want %d", got, water-1)
			}
			return
		}
	}
	t.Fatal("no drink option")
}

func TestEatingAndDrinkingAreRefusedWithAHostileInSight(t *testing.T) {
	r := newPartyRig(t)
	reaper := r.member("template.person.reaper")
	party, ok := anatomy.PartyOf(r.w, reaper)
	if !ok {
		t.Fatal("the Reaper is in no party")
	}
	ecs.NewMap[comp.Party](r.w.ECS).Get(party).Alert = true

	view, err := r.g.View()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[api.ContentID]bool{}
	for _, option := range view.Options {
		do, isDo := option.Command.(api.Do)
		if !isDo || (do.Action != "action.eat" && do.Action != "action.drink") {
			continue
		}
		seen[do.Action] = true
		want := r.w.Content.Render(api.Text{ID: "text.actions.on_alert"})
		if option.Enabled || option.Reason != want {
			t.Errorf(
				"%s on alert: enabled=%v reason=%q, want disabled with %q",
				option.ID,
				option.Enabled,
				option.Reason,
				want,
			)
		}
	}
	if !seen["action.eat"] || !seen["action.drink"] {
		t.Errorf("options seen %v, want eat and drink both listed", seen)
	}
}
