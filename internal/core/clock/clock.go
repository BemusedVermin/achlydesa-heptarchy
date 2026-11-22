package clock

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Clock maps ticks onto phases and days. It holds no time itself: the tick is
// sim.World.Now.
type Clock struct {
	calendar content.Calendar
}

// New returns a clock for cal, which must be valid (content.Load checks it).
func New(cal content.Calendar) *Clock {
	cal.Phases = append([]content.Phase(nil), cal.Phases...)
	return &Clock{calendar: cal}
}

func (c *Clock) ticksPerDay() int64 {
	return c.calendar.TicksPerPhase * int64(len(c.calendar.Phases))
}

// Day is the day number at tick now, counting from 1.
func (c *Clock) Day(now int64) int { return int(now/c.ticksPerDay()) + 1 }

// Phase is the index into the calendar's phases at tick now.
func (c *Clock) Phase(now int64) int {
	return int(now / c.calendar.TicksPerPhase % int64(len(c.calendar.Phases)))
}

// Advance moves time forward by ticks. For each phase boundary crossed, in
// order, it sets w.Now to the boundary tick, publishes PhaseStarted and runs
// PerPhase systems; when that boundary also starts a day, it then publishes
// DayStarted and runs PerDay systems. Finally it sets w.Now to the end tick.
// PerTurn systems are run by turns (M1-15), not here.
func (c *Clock) Advance(w *sim.World, ticks int64) error {
	if ticks < 0 {
		return fmt.Errorf("clock: cannot advance by %d ticks", ticks)
	}
	per := c.calendar.TicksPerPhase
	end := w.Now + ticks
	for boundary := (w.Now/per + 1) * per; boundary <= end; boundary += per {
		w.Now = boundary
		phase := c.calendar.Phases[c.Phase(boundary)]
		w.Bus.Publish(api.PhaseStarted{Day: c.Day(boundary), Phase: phase.ID})
		if _, err := w.Bus.Drain(w); err != nil {
			return fmt.Errorf("clock: phase %q: %w", phase.ID, err)
		}
		if err := w.Systems.Run(w, sim.PerPhase); err != nil {
			return fmt.Errorf("clock: per-phase systems: %w", err)
		}
		if boundary%c.ticksPerDay() != 0 {
			continue
		}
		w.Bus.Publish(api.DayStarted{Day: c.Day(boundary)})
		if _, err := w.Bus.Drain(w); err != nil {
			return fmt.Errorf("clock: day %d: %w", c.Day(boundary), err)
		}
		if err := w.Systems.Run(w, sim.PerDay); err != nil {
			return fmt.Errorf("clock: per-day systems: %w", err)
		}
	}
	w.Now = end
	return nil
}
