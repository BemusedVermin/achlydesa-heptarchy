package combat_test

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// propertyChecks is low because a fight is slow; -rapid.checks or RAPID_CHECKS overrides it.
	propertyChecks = 20
	propertyTicks  = 300
	// propertySteps bounds a fight in case the clock stops advancing.
	propertySteps = 100
	// replayEvery: one fight in this many is replayed, since a replay costs as much as the fight.
	replayEvery = 5
)

// withChecks sets rapid's check count for this test alone, unless the run asked for one.
func withChecks(t *testing.T, checks int) {
	t.Helper()
	explicit := os.Getenv("RAPID_CHECKS") != ""
	flag.Visit(func(set *flag.Flag) { explicit = explicit || set.Name == "rapid.checks" })
	if explicit {
		return
	}
	setting := flag.Lookup("rapid.checks").Value
	before := setting.String()
	if err := setting.Set(strconv.Itoa(checks)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := setting.Set(before); err != nil {
			t.Error(err)
		}
	})
}

// propertyContent is the hall as shipped (a hunting crawler and the whole party) with a piece of
// Serilmen in the inventory, so every form is on the menu.
func propertyContent(t *testing.T) *content.Registry {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	party := bundle.Templates["template.party.main"]
	party.Inventory = append(slices.Clone(party.Inventory), content.StackDef{Item: "item.serilmen_arm", Count: 1})
	bundle.Templates["template.party.main"] = party
	hall := bundle.Places["place.test.hall"]
	rows := slices.Clone(hall.Layers[0].Rows)
	for i, row := range rows {
		if strings.Contains(row, "@") {
			rows[i] = strings.Replace(strings.Replace(row, "@", ".", 1), ".......c", "..@....c", 1)
		}
	}
	hall.Layers = []content.LayerDef{{Depth: hall.Layers[0].Depth, Rows: rows}}
	bundle.Places["place.test.hall"] = hall
	return contenttest.Registry(t, bundle)
}

// fight is one random fight: its game, and what its events have shown so far.
type fight struct {
	g                         *game.Game
	w                         *sim.World
	events                    []api.Event
	started, landed, canceled int
	problems                  []string
}

func (f *fight) fail(format string, args ...any) {
	f.problems = append(f.problems, fmt.Sprintf(format, args...))
}

func newFight(t *testing.T, reg *content.Registry, seed uint64) (*fight, game.Config) {
	t.Helper()
	f := &fight{}
	cfg := game.Config{Seed: seed, Content: reg, Install: func(w *sim.World, _ *clock.Clock) error {
		w.Bus.SubscribeAll("test.property", func(w *sim.World, e api.Event) { f.observe(w, e) })
		return nil
	}}
	f.g, f.w, _ = gametest.Start(t, cfg)
	return f, cfg
}

// observe checks what each event says at the moment it is published.
func (f *fight) observe(w *sim.World, e api.Event) {
	f.events = append(f.events, e)
	switch e := e.(type) {
	case api.ActionStarted:
		f.started++
		if body, ok := w.ByUID(e.Source); ok && ecs.NewMap[comp.Dead](w.ECS).Has(body) {
			f.fail("dead body %d started %s", e.Source, e.Action)
		}
	case api.ActionLanded:
		f.landed++
	case api.ActionCanceled:
		f.canceled++
	case api.PartDamaged:
		f.checkInReach(w, e)
	case api.Rejected:
		f.fail("an offered option was rejected: %+v", e)
	}
}

// checkInReach says a blow that a body dealt landed on a body within a step of it.
func (f *fight) checkInReach(w *sim.World, e api.PartDamaged) {
	if e.Source == 0 || e.Source == e.Body {
		return
	}
	source, ok1 := w.ByUID(e.Source)
	target, ok2 := w.ByUID(e.Body)
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !ok1 || !ok2 || !positions.Has(source) || !positions.Has(target) {
		return
	}
	from, to := positions.Get(source).At, positions.Get(target).At
	reach := hexmap.CellDistance(from, to)
	if reach > 1 {
		f.fail("body %d hit body %d from %d cells away", e.Source, e.Body, reach)
	}
}

// check looks at the world between two commands.
func (f *fight) check() {
	w := f.w
	dead := ecs.NewMap[comp.Dead](w.ECS)
	partOf := ecs.NewMap[comp.PartOf](w.ECS)
	positions := ecs.NewMap[comp.Pos](w.ECS)

	cells := map[comp.Pos]api.UID{}
	query := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for query.Next() {
		body := query.Entity()
		if dead.Has(body) {
			continue
		}
		pos := *positions.Get(body)
		if other, clash := cells[pos]; clash {
			f.fail("bodies %d and %d share %+v", other, w.UIDOf(body), pos)
		}
		cells[pos] = w.UIDOf(body)
	}
	owner := map[ecs.Entity]api.UID{}
	all := ecs.NewFilter1[comp.Identity](w.ECS).Query()
	var everyone []ecs.Entity
	for all.Next() {
		everyone = append(everyone, all.Entity())
	}
	for _, body := range everyone {
		for _, part := range anatomy.Parts(w, body) {
			if first, twice := owner[part]; twice {
				f.fail("part %d is on bodies %d and %d", w.UIDOf(part), first, w.UIDOf(body))
			}
			owner[part] = w.UIDOf(body)
		}
	}

	used := map[api.UID]api.UID{}
	pendings := 0
	pending := ecs.NewFilter2[comp.Pending, comp.Landing](w.ECS).Query()
	for pending.Next() {
		p, landing := pending.Get()
		pendings++
		actor, ok := w.ByUID(p.Actor)
		if !ok {
			f.fail("pending action of unknown actor %d", p.Actor)
			continue
		}
		for _, channel := range landing.Channels {
			if first, busy := used[channel]; busy {
				f.fail("channel %d is in two actions in flight (%d and %d)", channel, first, p.Actor)
			}
			used[channel] = p.Actor
			part, ok := w.ByUID(channel)
			if !ok {
				f.fail("channel %d of %d is gone", channel, p.Actor)
				continue
			}
			if part != actor && (!partOf.Has(part) || partOf.GetRelation(part) != actor) {
				f.fail("channel %d acts for %d but is not on it", channel, p.Actor)
			}
		}
		if dead.Has(actor) {
			f.fail("a dead body %d has an action in flight", p.Actor)
		}
	}
	if f.started != f.landed+f.canceled+pendings {
		f.fail("%d actions started, %d landed, %d canceled, %d in flight", f.started, f.landed, f.canceled, pendings)
	}
}

// decides says who the view waits on, and checks they are alive.
func (f *fight) decides(view api.View) {
	if view.Timeline == nil || view.Timeline.Current == 0 {
		return
	}
	if body, ok := f.w.ByUID(view.Timeline.Current); ok && ecs.NewMap[comp.Dead](f.w.ECS).Has(body) {
		f.fail("the dead body %d is deciding", view.Timeline.Current)
	}
}

// TestRandomFightsKeepTheInvariants plays random fights, choosing an offered option each time for
// up to 300 ticks, and holds every property of combat after each command. Only options the view
// offers are applied, and none may be rejected. About one fight in replayEvery is also replayed
// from the same seed and must give the same hash.
func TestRandomFightsKeepTheInvariants(t *testing.T) {
	if testing.Short() {
		t.Skip("random fights")
	}
	reg := propertyContent(t)
	withChecks(t, propertyChecks)
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Uint64().Draw(rt, "seed")
		picks := rapid.SliceOfN(rapid.IntRange(0, 1<<20), propertySteps/4, propertySteps).Draw(rt, "picks")
		replay := rapid.IntRange(0, replayEvery-1).Draw(rt, "replay") == 0
		f, cfg := newFight(t, reg, seed)
		var trail []string
		for _, pick := range picks {
			if f.w.Now >= propertyTicks {
				break
			}
			view, err := f.g.View()
			if err != nil {
				rt.Fatal(err)
			}
			f.decides(view)
			var enabled []api.Option
			for _, option := range view.Options {
				if option.Enabled {
					enabled = append(enabled, option)
				}
			}
			if len(enabled) == 0 {
				break
			}
			chosen := enabled[pick%len(enabled)]
			trail = append(trail, chosen.ID)
			if _, err := f.g.Apply(chosen.Command); err != nil {
				rt.Fatalf("option %q: %v", chosen.ID, err)
			}
			f.check()
			if len(f.problems) > 0 {
				rt.Fatalf("after %q at tick %d: %v (trail %v)", chosen.ID, f.w.Now, f.problems, trail)
			}
		}
		if !replay {
			return
		}
		want, err := f.g.Hash()
		if err != nil {
			rt.Fatal(err)
		}
		replayed, err := game.Replay(cfg, f.g.History())
		if err != nil {
			rt.Fatal(err)
		}
		got, err := replayed.Hash()
		if err != nil {
			rt.Fatal(err)
		}
		if got != want {
			rt.Fatal("replaying the same commands from the same seed gave another hash")
		}
	})
}
