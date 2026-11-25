package clock

import (
	"errors"
	"strconv"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Install registers the clock's events and its view builder.
func Install(w *sim.World, c *Clock) error {
	if w.Content == nil {
		return errors.New("clock: install: world has no content")
	}
	if err := api.RegisterEvent[api.PhaseStarted](w.Codec); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.DayStarted](w.Codec); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.ExportBegan](w.Codec); err != nil {
		return err
	}
	return w.Views.Register(sim.ViewBuilder{
		Name: "clock",
		Build: func(w *sim.World, v *api.View) error {
			index := c.Phase(w.Now)
			phase := c.calendar.Phases[index]
			view := &api.ClockView{
				Day:         c.Day(w.Now),
				Phase:       w.Content.Render(api.Text{ID: phase.Name}),
				TickInPhase: c.TickInPhase(w.Now),
				PhaseTicks:  c.PhaseTicks(index),
			}
			view.Label = w.Content.Render(api.Text{ID: "text.clock.label", Args: map[string]string{
				"day": strconv.Itoa(view.Day), "phase": view.Phase,
			}})
			v.Clock = view
			return nil
		},
	})
}
