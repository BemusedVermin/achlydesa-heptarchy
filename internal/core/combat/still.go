package combat

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const stillController = "still"

// stiller occupies every free channel of a body for the still ticks, then decides again. Unlike the
// idle controller, two of them don't advance time a tick at a time.
type stiller struct{}

func (stiller) Decide(w *sim.World, body ecs.Entity) error {
	ticks := max(w.Content.Combat().StillTicks, 1)
	for _, channel := range timeline.Channels(w, body) {
		if !timeline.Free(w, channel) {
			continue
		}
		if err := timeline.Occupy(w, channel, ticks); err != nil {
			return err
		}
	}
	return nil
}
