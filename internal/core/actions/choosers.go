package actions

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// DefaultChooser is the chooser a body uses when its template's actor names none.
const DefaultChooser = "score"

// Chooser picks one of the intents a body could start. Choosers are registered by name, and a
// template's actor.chooser says which one its bodies use.
type Chooser interface {
	// Choose returns the intent to start, or false when it wants none.
	Choose(w *sim.World, actor ecs.Entity, intents []Intent) (Intent, bool)
}

// RegisterChooser makes a chooser available by name. Names are unique.
func RegisterChooser(w *sim.World, name string, c Chooser) error {
	if name == "" || c == nil {
		return errors.New("actions: a chooser needs a name and an implementation")
	}
	st := stateOf(w)
	if _, dup := st.choosers[name]; dup {
		return fmt.Errorf("actions: chooser %q already registered", name)
	}
	st.choosers[name] = c
	return nil
}

// ChooserFor is the chooser the body's template names in actor.chooser, else [DefaultChooser]. It
// reports false when that chooser is not registered.
func ChooserFor(w *sim.World, body ecs.Entity) (Chooser, bool) {
	c, ok := stateOf(w).choosers[ChooserName(w, body)]
	return c, ok
}

// checkChooser refuses to spawn a body whose template names a chooser nobody registered.
func checkChooser(w *sim.World, _ ecs.Entity, t content.Template) error {
	if t.Actor == nil || t.Actor.Chooser == "" {
		return nil
	}
	if _, ok := stateOf(w).choosers[t.Actor.Chooser]; !ok {
		return fmt.Errorf("unknown chooser %q", t.Actor.Chooser)
	}
	return nil
}

// ChooserName is the name of the chooser the body's template names in actor.chooser, else [DefaultChooser].
func ChooserName(w *sim.World, body ecs.Entity) string {
	t, ok := w.Content.Template(w.TemplateOf(body))
	if !ok || t.Actor == nil || t.Actor.Chooser == "" {
		return DefaultChooser
	}
	return t.Actor.Chooser
}
