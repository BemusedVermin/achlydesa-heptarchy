package combat

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// stillController is the name of the controller of a body that does nothing for a while.
const stillController = "still"

// stiller keeps a body still: it takes every one of the body's channels for the configured still ticks,
// then decides again. The built-in idle controller waits for the next entry on the timeline
// instead, which with two such bodies alternating means time moves on a tick at a time. A real
// creature mind (M2-08) replaces this.
type stiller struct{}

// Decide occupies every free channel of the body for the still ticks.
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
