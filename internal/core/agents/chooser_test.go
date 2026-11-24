package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

func (f *primField) members() []ecs.Entity {
	f.t.Helper()
	party, ok := anatomy.ControlledParty(f.w)
	if !ok {
		f.t.Fatal("no controlled party")
	}
	return anatomy.Members(f.w, party)
}

func (f *primField) place(body ecs.Entity, at api.Cell) {
	ecs.NewMap[comp.Pos](f.w.ECS).Get(body).At = at
}

func (f *primField) vitalPart(body ecs.Entity) ecs.Entity {
	f.t.Helper()
	parts := ecs.NewMap[comp.Part](f.w.ECS)
	for _, part := range anatomy.Parts(f.w, body) {
		if parts.Get(part).Vital {
			return part
		}
	}
	f.t.Fatal("body has no vital part")
	return ecs.Entity{}
}

func TestTheAgentsChooserGoesForTheDying(t *testing.T) {
	for _, dying := range []int{0, 1} {
		for seed := uint64(1); seed <= 3; seed++ {
			f := newPrimField(t, seed, nil, func(bundle *content.Bundle) {
				bundle.Rules["agents_score"] = "target.dying ? 10 : 0"
			})
			members := f.members()
			f.place(members[0], cellAt(13, 10))
			f.place(members[1], cellAt(15, 10))
			ecs.NewMap[comp.Part](f.w.ECS).Get(f.vitalPart(members[dying])).Integrity = 1
			predatorUID := f.w.UIDOf(f.spawn("template.test.predator", cellAt(14, 10)))
			var swing *api.ActionStarted
			f.until(200, func() bool {
				for _, e := range f.events {
					if s, ok := e.(api.ActionStarted); ok && s.Source == predatorUID && s.Target != 0 {
						swing = &s
						return true
					}
				}
				return false
			})
			if swing == nil {
				t.Fatalf("dying %d, seed %d: the predator never swung", dying, seed)
			}
			if want := f.w.UIDOf(members[dying]); swing.Target != want {
				t.Errorf("dying %d, seed %d: swung at %d, want the dying member %d", dying, seed, swing.Target, want)
			}
		}
	}
}

func TestADodgerStepsOutOfRangeOfAWindingUpStrike(t *testing.T) {
	f := newPrimField(t, 1, nil, nil)
	attacker := f.members()[0]
	f.place(attacker, cellAt(13, 10))
	dodger := f.spawn("template.test.dodger", cellAt(14, 10))
	err := actions.Start(f.w, actions.Intent{
		Actor: attacker, Action: "action.strike", Target: dodger, TargetPart: f.vitalPart(dodger),
		Choices: map[string]string{"type": "type.cut"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(timeline.AimedAt(f.w, f.w.UIDOf(dodger))) != 1 {
		t.Fatal("the strike is not in flight")
	}
	controller, ok := timeline.ControllerNamed(f.w, "agents")
	if !ok {
		t.Fatal("no agents controller")
	}
	if err := controller.Decide(f.w, dodger); err != nil {
		t.Fatal(err)
	}
	positions := ecs.NewMap[comp.Pos](f.w.ECS)
	if d := hexmap.Distance(positions.Get(attacker).At.Hex, positions.Get(dodger).At.Hex); d <= 1 {
		t.Fatalf("the dodger is %d hexes from the attacker, want out of the strike's reach", d)
	}
	missed := f.until(40, func() bool {
		for _, e := range f.events {
			if m, ok := e.(api.ActionMissed); ok && m.Source == f.w.UIDOf(attacker) {
				return true
			}
		}
		return false
	})
	if !missed {
		t.Error("the strike did not miss")
	}
	for _, e := range f.events {
		if _, hurt := e.(api.PartDamaged); hurt {
			t.Errorf("a part was damaged despite the dodge: %+v", e)
		}
	}
}
