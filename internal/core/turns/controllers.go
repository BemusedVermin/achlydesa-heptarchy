package turns

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// PlayerController is the reserved controller name for entities the player plays: the loop
	// stops and waits for commands.
	PlayerController = "player"
	// IdleController is the built-in controller that just ends its turn.
	IdleController = "idle"
)

// Controller plays a turn for anything the player doesn't control.
type Controller interface {
	// TakeTurn may step and act; the loop ends the turn if the controller didn't.
	TakeTurn(w *sim.World, actor ecs.Entity) error
}

type idle struct{}

func (idle) TakeTurn(*sim.World, ecs.Entity) error { return nil }

// RegisterController makes c available to templates as name. The name "player" is reserved, and
// names are unique.
func RegisterController(w *sim.World, name string, c Controller) error {
	switch {
	case name == "" || c == nil:
		return errors.New("turns: a controller needs a name and an implementation")
	case name == PlayerController:
		return fmt.Errorf("turns: controller name %q is reserved", name)
	}
	st := stateOf(w)
	if _, dup := st.controllers[name]; dup {
		return fmt.Errorf("turns: controller %q already registered", name)
	}
	st.controllers[name] = c
	return nil
}

// SetSpeed replaces how an entity's speed is found. The default is the configured default_speed;
// M2-02 sums the legs.
func SetSpeed(w *sim.World, fn func(w *sim.World, e ecs.Entity) int) { stateOf(w).speed = fn }

// SetBudget replaces what an entity may do in one turn. The default is one step and no actions;
// M2-02 counts working arms.
func SetBudget(w *sim.World, fn func(w *sim.World, e ecs.Entity) (steps, actions int)) {
	stateOf(w).budget = fn
}
