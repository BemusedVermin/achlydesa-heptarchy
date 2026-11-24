package timeline

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// PlayerController is the reserved controller name for bodies the player plays: the loop
	// stops and waits for commands.
	PlayerController = "player"
	// IdleController is the built-in controller that does nothing, so the loop makes it wait.
	IdleController = "idle"
)

// Controller decides for anything the player doesn't control.
type Controller interface {
	// Decide may occupy channels or wait. If it does neither, the loop makes the body wait until
	// the next entry on the timeline.
	Decide(w *sim.World, body ecs.Entity) error
}

type idle struct{}

func (idle) Decide(*sim.World, ecs.Entity) error { return nil }

// RegisterController makes c available to templates as name. The name "player" is reserved, and
// names are unique.
func RegisterController(w *sim.World, name string, c Controller) error {
	switch {
	case name == "" || c == nil:
		return errors.New("timeline: a controller needs a name and an implementation")
	case name == PlayerController:
		return fmt.Errorf("timeline: controller name %q is reserved", name)
	}
	st := stateOf(w)
	if _, dup := st.controllers[name]; dup {
		return fmt.Errorf("timeline: controller %q already registered", name)
	}
	st.controllers[name] = c
	return nil
}

// SetChannels replaces how a body's channels are found. The default is one channel, the body
// itself, tagged "body"; M2-02 adds one per attached part.
func SetChannels(w *sim.World, fn func(w *sim.World, body ecs.Entity) []Channel) {
	stateOf(w).channels = fn
}

// SetUsable replaces the test for whether a channel can be used at all. The default is that every
// channel is usable; the actions package installs [SetUsableChannels] instead, which marks a channel
// usable when some action could start with it.
func SetUsable(w *sim.World, fn func(w *sim.World, body ecs.Entity, c Channel) bool) {
	st := stateOf(w)
	st.usable, st.usableAll = fn, nil
}

// SetUsableChannels replaces the test for whether channels can be used at all with one that answers
// for all of a body's channels in one pass. It returns one flag per channel, in order, and must agree
// with what a per-channel test would say. A later [SetUsable] replaces it.
func SetUsableChannels(w *sim.World, fn func(w *sim.World, body ecs.Entity, channels []Channel) []bool) {
	stateOf(w).usableAll = fn
}

// SetSpeed replaces how a body's speed is found. The default is default_speed; errors from fn propagate.
func SetSpeed(w *sim.World, fn func(w *sim.World, body ecs.Entity) (int, error)) {
	stateOf(w).speed = fn
}

// SetVisible replaces which bodies the player may see on the timeline. The default is party members.
func SetVisible(w *sim.World, fn func(w *sim.World, body ecs.Entity) bool) { stateOf(w).visible = fn }

// SetHalt sets a check Run makes before each step: when it reports true, Run returns at once with
// the zero entity and no error, because no decision is left to wait for. Combat sets it to the
// party being wiped, so a landing that downs the last member ends the call. The default never halts.
func SetHalt(w *sim.World, fn func(w *sim.World) bool) { stateOf(w).halt = fn }
