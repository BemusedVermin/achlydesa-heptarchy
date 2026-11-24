package clock

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Install registers the clock's events and its view builder. reg renders phase
// names for the view when the world has no content of its own.
func Install(w *sim.World, c *Clock, reg *content.Registry) error {
	if err := api.RegisterEvent[api.PhaseStarted](w.Codec); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.DayStarted](w.Codec); err != nil {
		return err
	}
	return w.Views.Register(sim.ViewBuilder{
		Name: "clock",
		Build: func(w *sim.World, v *api.View) error {
			current := w.Content // the live registry, which content hot reload swaps
			if current == nil {
				current = reg
			}
			index := c.Phase(w.Now)
			phase := c.calendar.Phases[index]
			v.Clock = &api.ClockView{
				Day:         c.Day(w.Now),
				Phase:       current.Render(api.Text{ID: phase.Name}),
				TickInPhase: c.TickInPhase(w.Now),
				PhaseTicks:  c.PhaseTicks(index),
			}
			return nil
		},
	})
}
