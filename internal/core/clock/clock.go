package clock

import (
	"fmt"
	"sort"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Clock maps ticks onto phases and days. It holds no time itself: the tick is
// sim.World.Now.
type Clock struct {
	calendar content.Calendar
	// Starts[i] is the tick within a day at which phase i begins
	starts   []int64
	dayTicks int64
}

// New returns a clock for cal, which must be valid (content.Load checks it).
func New(cal content.Calendar) *Clock {
	clk := &Clock{}
	clk.Reset(cal)
	return clk
}

// Reset replaces the calendar in place, so everything holding this clock sees the new one. cal
// must be valid. Ticks keep their meaning (the tick is still sim.World.Now), but a new day length
// moves where days and phases fall.
func (c *Clock) Reset(cal content.Calendar) {
	cal.Phases = append([]content.Phase(nil), cal.Phases...)
	c.calendar = cal
	c.starts = make([]int64, len(cal.Phases))
	c.dayTicks = 0
	for i, phase := range cal.Phases {
		c.starts[i] = c.dayTicks
		c.dayTicks += phase.Ticks
	}
}

// Day is the day number at tick now, counting from 1.
func (c *Clock) Day(now int64) int { return int(now/c.dayTicks) + 1 }

// Phase is the index into the calendar's phases at tick now.
func (c *Clock) Phase(now int64) int {
	inDay := now % c.dayTicks
	return sort.Search(len(c.starts), func(i int) bool { return c.starts[i] > inDay }) - 1
}

// TickInPhase is how many ticks into its phase tick now is.
func (c *Clock) TickInPhase(now int64) int64 {
	return now%c.dayTicks - c.starts[c.Phase(now)]
}

// PhaseTicks is how long the phase at index phase lasts.
func (c *Clock) PhaseTicks(phase int) int64 { return c.calendar.Phases[phase].Ticks }

// Advance moves time forward by ticks. For each phase boundary crossed, in
// order, it sets w.Now to the boundary tick, publishes PhaseStarted and runs
// PerPhase systems; when that boundary also starts a day, it then publishes
// DayStarted and runs PerDay systems. Finally it sets w.Now to the end tick.
func (c *Clock) Advance(w *sim.World, ticks int64) error {
	if ticks < 0 {
		return fmt.Errorf("clock: cannot advance by %d ticks", ticks)
	}
	end := w.Now + ticks
	// The next boundary is the start of the phase after the current one.
	phase := c.Phase(w.Now)
	boundary := w.Now - c.TickInPhase(w.Now) + c.PhaseTicks(phase)
	for ; boundary <= end; boundary += c.PhaseTicks(phase) {
		phase = (phase + 1) % len(c.calendar.Phases)
		w.Now = boundary
		started := c.calendar.Phases[phase]
		w.Bus.Publish(api.PhaseStarted{Day: c.Day(boundary), Phase: started.ID})
		if _, err := w.Bus.Drain(w); err != nil {
			return fmt.Errorf("clock: phase %q: %w", started.ID, err)
		}
		if err := w.Systems.Run(w, sim.PerPhase); err != nil {
			return fmt.Errorf("clock: per-phase systems: %w", err)
		}
		if phase != 0 {
			continue
		}
		w.Bus.Publish(api.DayStarted{Day: c.Day(boundary)})
		if _, err := w.Bus.Drain(w); err != nil {
			return fmt.Errorf("clock: day %d: %w", c.Day(boundary), err)
		}
		if err := c.exportIfDue(w, c.Day(boundary)); err != nil {
			return err
		}
		if err := w.Systems.Run(w, sim.PerDay); err != nil {
			return fmt.Errorf("clock: per-day systems: %w", err)
		}
	}
	w.Now = end
	return nil
}

// exportIfDue begins an export when day has reached the scheduled export day, and draws the day of
// the next one from the "clock.exports" stream.
func (c *Clock) exportIfDue(w *sim.World, day int) error {
	schedule := c.calendar.Exports
	if w.NextExportDay == 0 {
		w.NextExportDay = schedule.FirstDay
	}
	if day < w.NextExportDay {
		return nil
	}
	w.ExportCount++
	gap := schedule.MinGap + w.RNG.Get("clock.exports").IntN(schedule.MaxGap-schedule.MinGap+1)
	w.NextExportDay = day + gap
	w.Bus.Publish(api.ExportBegan{Day: day, Count: w.ExportCount})
	if _, err := w.Bus.Drain(w); err != nil {
		return fmt.Errorf("clock: export %d: %w", w.ExportCount, err)
	}
	return nil
}
