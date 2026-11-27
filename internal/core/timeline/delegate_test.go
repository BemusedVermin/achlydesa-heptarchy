package timeline_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// deputy is a delegate that logs the ticks it decides at and steps, and hands the body back after
// resumeAfter decisions. With step false it hands off without acting.
type deputy struct {
	ticks       *[]int64
	resumeAfter int
	step        bool
}

func (d deputy) Decide(w *sim.World, body ecs.Entity) error {
	*d.ticks = append(*d.ticks, w.Now)
	if len(*d.ticks) >= d.resumeAfter {
		timeline.Resume(w, body)
	}
	if !d.step {
		return nil
	}
	ticks, err := timeline.StepTicks(w, body)
	if err != nil {
		return err
	}
	return timeline.Occupy(w, timeline.Channels(w, body)[0], ticks)
}

// delegator hands the body to target without acting.
type delegator struct{ target string }

func (d delegator) Decide(w *sim.World, body ecs.Entity) error {
	return timeline.DelegateTo(w, body, d.target)
}

func TestADelegateDecidesUntilItResumes(t *testing.T) {
	fx := newFixture(t, 1)
	var deputyTicks []int64
	if err := timeline.RegisterController(
		fx.w,
		"deputy",
		deputy{ticks: &deputyTicks, resumeAfter: 2, step: true},
	); err != nil {
		t.Fatal(err)
	}
	body := fx.spawn(t, "record", 2)
	if err := timeline.DelegateTo(fx.w, body, "deputy"); err != nil {
		t.Fatal(err)
	}
	if got := timeline.ControllerOf(fx.w, body); got != "deputy" {
		t.Fatalf("controller after DelegateTo = %q, want deputy", got)
	}
	step := fx.stepTicks(t, 2)
	fx.stopAfter = 3 * step
	fx.run(t)
	if want := []int64{0, step}; !slices.Equal(deputyTicks, want) {
		t.Errorf("deputy decided at %v, want %v", deputyTicks, want)
	}
	uid := fx.w.UIDOf(body)
	if want := []decision{{2 * step, uid}, {3 * step, uid}}; !slices.Equal(fx.taken, want) {
		t.Errorf("own controller decided %v, want %v", fx.taken, want)
	}
	if got := timeline.ControllerOf(fx.w, body); got != "record" {
		t.Errorf("controller after Resume = %q, want record", got)
	}
}

func TestAHandOffWithoutActingLetsTheNextControllerDecideAtOnce(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		delegate string
		want     int
	}{
		{"a delegate that resumes at once", "record", "yield", 1},
		{"a controller that delegates at once", "pass", "record", 1},
		{"two controllers handing back and forth", "pass", "yield", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newFixture(t, 1)
			var yielded []int64
			if err := timeline.RegisterController(fx.w, "yield", deputy{ticks: &yielded, resumeAfter: 1}); err != nil {
				t.Fatal(err)
			}
			if err := timeline.RegisterController(fx.w, "pass", delegator{target: tt.delegate}); err != nil {
				t.Fatal(err)
			}
			body := fx.spawn(t, tt.base, 2)
			if tt.base == "record" {
				if err := timeline.DelegateTo(fx.w, body, tt.delegate); err != nil {
					t.Fatal(err)
				}
			}
			player := fx.spawn(t, timeline.PlayerController, 2)
			if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player || fx.w.Now != 0 {
				t.Fatalf("Run = %v, %v at tick %d; want the player at tick 0", current, err, fx.w.Now)
			}
			if len(fx.taken) != tt.want {
				t.Errorf("record decided %d times at tick 0, want %d", len(fx.taken), tt.want)
			}
		})
	}
}

func TestDelegateToRefusesWhatCannotDecide(t *testing.T) {
	fx := newFixture(t, 1)
	body := fx.spawn(t, "record", 2)
	for _, name := range []string{"nobody", timeline.PlayerController} {
		if err := timeline.DelegateTo(fx.w, body, name); err == nil {
			t.Errorf("DelegateTo(%q) succeeded, want an error", name)
		}
	}
	if got := timeline.ControllerOf(fx.w, body); got != "record" {
		t.Errorf("controller = %q after refused delegations, want record", got)
	}
}

func TestADelegateSurvivesASaveAndLoad(t *testing.T) {
	g, w, cfg := gametest.New(t, 1, nil)
	var bodies []ecs.Entity
	query := ecs.NewFilter1[comp.Actor](w.ECS).Query()
	for query.Next() {
		bodies = append(bodies, query.Entity())
	}
	bodies = w.SortByUID(bodies)
	if len(bodies) == 0 {
		t.Fatal("the game has no actors")
	}
	uid := w.UIDOf(bodies[0])
	if err := timeline.DelegateTo(w, bodies[0], timeline.IdleController); err != nil {
		t.Fatal(err)
	}
	_, loaded, _ := gametest.Reload(t, cfg, g)
	body, ok := loaded.ByUID(uid)
	if !ok {
		t.Fatalf("body %d is gone after the load", uid)
	}
	if got := timeline.ControllerOf(loaded, body); got != timeline.IdleController {
		t.Errorf("controller after the load = %q, want %q", got, timeline.IdleController)
	}
}
