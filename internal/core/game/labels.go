package game

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// installLabels registers the view builder of the panel headings.
func installLabels(w *sim.World) error {
	return w.Views.Register(sim.ViewBuilder{Name: "game.labels", Build: func(w *sim.World, v *api.View) error {
		v.Labels = api.LabelsView{
			Actions:   w.Content.Render(api.Text{ID: "text.game.actions"}),
			Look:      w.Content.Render(api.Text{ID: "text.game.look"}),
			LookEmpty: w.Content.Render(api.Text{ID: "text.game.look_empty"}),
			Stock:     w.Content.Render(api.Text{ID: "text.game.stock"}),
		}
		return nil
	}})
}
