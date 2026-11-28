package agents

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// SeekGroundPicker is the target picker that steps toward the least decayed ground the agent
// remembers, if that is less decayed than where it stands, and otherwise wanders. Its step is done
// once the agent stands in undecayed ground.
const SeekGroundPicker = "seek_ground"

type seekGroundPicker struct{}

// decayUnderfoot is the decay level of the ground the body stands in, 0 with no geography or position.
func decayUnderfoot(w *sim.World, body ecs.Entity) (int64, error) {
	geography := w.Content.World().Geography
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if geography == "" || !positions.Has(body) {
		return 0, nil
	}
	level, err := w.Content.DecayLevel(geography, content.DecayEnv{
		Depth: positions.Get(body).At.Depth, Exports: max(w.ExportCount, 0),
	})
	if err != nil {
		return 0, fmt.Errorf("agents: ground: %w", err)
	}
	return level, nil
}

func (seekGroundPicker) Pick(w *sim.World, body ecs.Entity, step Step) (actions.Intent, bool, error) {
	here, err := decayUnderfoot(w, body)
	if err != nil {
		return actions.Intent{}, false, err
	}
	known, err := KnownGround(w, body)
	if err != nil {
		return actions.Intent{}, false, err
	}
	if len(known) > 0 && known[0].Level < here {
		step.Target = &Target{At: known[0].At}
		return goToPicker{}.Pick(w, body, step)
	}
	return wanderPicker{}.Pick(w, body, step)
}

func (seekGroundPicker) Finished(w *sim.World, body ecs.Entity, _ Step) bool {
	here, err := decayUnderfoot(w, body)
	return err == nil && here == 0
}
