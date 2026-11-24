package combat

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// state is what combat keeps on the world, as an Ark resource.
type state struct {
	choosers map[string]Chooser
}

func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{choosers: map[string]Chooser{}}
	res.Add(st)
	return st
}
