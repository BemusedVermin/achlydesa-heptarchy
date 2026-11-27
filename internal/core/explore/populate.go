package explore

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// populateWorld spawns every individual the generated ground starts with, wherever it stands and whether or
// not its chunk is loaded. The spawns come in chunk key then cell order, so UIDs never depend on how chunks
// came to load. It runs once, at world creation: a dead creature is never replaced.
func populateWorld(w *sim.World) error {
	st := stateOf(w)
	if st.generated == nil {
		return nil
	}
	spawns, err := st.generated.PopulateAll(0)
	if err != nil {
		return fmt.Errorf("explore: populate: %w", err)
	}
	minds := ecs.NewMap[comp.Mind](w.ECS)
	for _, spawn := range spawns {
		body, err := w.Spawn(spawn.Template, &comp.Pos{At: spawn.Cell})
		if err != nil {
			return fmt.Errorf("explore: populate: spawn %q at %v: %w", spawn.Template, spawn.Cell, err)
		}
		if spawn.Planner == "" {
			continue
		}
		if !minds.Has(body) {
			return fmt.Errorf("explore: populate: %q at %v has no mind to give planner %q",
				spawn.Template, spawn.Cell, spawn.Planner)
		}
		minds.Get(body).Planner = spawn.Planner
	}
	return nil
}
