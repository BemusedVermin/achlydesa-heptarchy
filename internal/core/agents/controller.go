package agents

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// controllerName is the controller a template names for a body that acts on its behaviors.
const controllerName = "agents"

// controller chooses among a body's behaviors at each decision and runs one step of the best.
type controller struct{}

// Decide scores the behaviors the body's template lists, records the winner in [comp.Doing] and
// calls its primitive's Step once. A body with no behavior worth doing waits.
func (controller) Decide(w *sim.World, body ecs.Entity) error {
	if !ecs.NewMap[comp.Pos](w.ECS).Has(body) {
		return nil
	}
	template, ok := w.Content.Template(w.TemplateOf(body))
	if !ok || len(template.Behaviors) == 0 {
		return nil
	}
	chooser := newChooser(w, body)
	best, ok, err := chooser.choose(template.Behaviors)
	if err != nil {
		return fmt.Errorf("agents: body %d: %w", w.UIDOf(body), err)
	}
	doing := ecs.NewMap[comp.Doing](w.ECS)
	if !ok {
		if doing.Has(body) {
			doing.Remove(body)
		}
		return nil
	}
	primitive, ok := stateOf(w).primitives[best.def.Do]
	if !ok {
		return fmt.Errorf("agents: behavior %q: unknown behavior primitive %q", best.id, best.def.Do)
	}
	next := comp.Doing{Behavior: best.id, Cell: best.target.Cell, Since: w.Now}
	if !best.target.Body.IsZero() {
		next.Target = w.UIDOf(best.target.Body)
	}
	if doing.Has(body) {
		if doing.Get(body).Behavior == best.id {
			next.Since = doing.Get(body).Since
		}
		*doing.Get(body) = next
	} else {
		doing.Add(body, &next)
	}
	done, err := primitive.Step(w, body, best.def, best.target)
	if err != nil {
		return fmt.Errorf("agents: body %d: behavior %q: %w", w.UIDOf(body), best.id, err)
	}
	// A finished behavior has no claim to the keep bonus at the next decision.
	if done && w.Alive(body) && doing.Has(body) {
		doing.Remove(body)
	}
	return nil
}
