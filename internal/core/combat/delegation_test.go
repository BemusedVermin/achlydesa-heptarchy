package combat_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// lastChooser takes the last intent it is offered and remembers each pick.
type lastChooser struct{ picks *[]actions.Intent }

func (c lastChooser) Choose(_ *sim.World, _ ecs.Entity, intents []actions.Intent) (actions.Intent, bool) {
	if len(intents) == 0 {
		return actions.Intent{}, false
	}
	*c.picks = append(*c.picks, intents[len(intents)-1])
	return intents[len(intents)-1], true
}

// deputy hands every decision to the hunt controller.
type deputy struct{}

func (deputy) Decide(w *sim.World, body ecs.Entity) error {
	hunt, ok := timeline.ControllerNamed(w, "hunt")
	if !ok {
		return errors.New("no hunt controller")
	}
	return hunt.Decide(w, body)
}

// standIn is the hunt field with the hall's creature made idle and a copy of it, whose actor is
// actor, in its cell. The copy is spawned after install, so it may name a controller or chooser
// that install registers.
func standIn(t *testing.T, actor content.ActorDef, install func(*sim.World) error) (*field, ecs.Entity) {
	t.Helper()
	f := huntFieldInstalling(t, func(b *content.Bundle) {
		movingClay := b.Templates["template.creature.moving_clay"]
		copied := movingClay
		copied.Name, copied.Actor = nil, &actor
		b.Templates["template.test.stand_in"] = copied
		b.Text["text.template.test.stand_in"] = "Stand-in"
		movingClay.Actor = &content.ActorDef{Controller: timeline.IdleController}
		b.Templates["template.creature.moving_clay"] = movingClay
	}, install)
	pos := *ecs.NewMap[comp.Pos](f.w.ECS).Get(f.creature())
	body, err := f.w.Spawn("template.test.stand_in", &pos)
	if err != nil {
		t.Fatal(err)
	}
	f.moveParty(east(pos.At, -5))
	return f, body
}

func TestTheHunterPicksThroughTheChooserItsTemplateNames(t *testing.T) {
	var picks []actions.Intent
	f, body := standIn(t, content.ActorDef{Controller: "hunt", Chooser: "test.last"}, func(w *sim.World) error {
		return actions.RegisterChooser(w, "test.last", lastChooser{picks: &picks})
	})
	uid := f.w.UIDOf(body)
	if !f.until(300, func() bool { return f.movingClayStarted(uid) }) {
		t.Fatal("the stand-in never attacked")
	}
	if len(picks) == 0 {
		t.Fatal("the test.last chooser was never asked")
	}
	swing, _ := f.started(uid)
	if first := picks[0]; swing.Action != first.Action || swing.Target != f.w.UIDOf(first.Target) {
		t.Errorf("the stand-in started %s on %d, want the chooser's %s on %d",
			swing.Action, swing.Target, first.Action, f.w.UIDOf(first.Target))
	}
}

func TestAControllerThatDelegatesToHuntAttacksAsAHunterDoes(t *testing.T) {
	f, body := standIn(t, content.ActorDef{Controller: "test.deputy"}, func(w *sim.World) error {
		return timeline.RegisterController(w, "test.deputy", deputy{})
	})
	uid := f.w.UIDOf(body)
	if !f.until(300, func() bool { return f.movingClayStarted(uid) }) {
		t.Fatal("the stand-in never attacked")
	}
	if got := f.moves(uid); got < 3 {
		t.Errorf("the stand-in took %d steps before its first swing, want several", got)
	}
	swing, _ := f.started(uid)
	if swing.Action != "action.strike" {
		t.Errorf("first action = %s, want a strike", swing.Action)
	}
	if !slices.Contains(f.memberUIDs(), swing.Target) {
		t.Errorf("the swing is aimed at %d, not a member of the party", swing.Target)
	}
}
