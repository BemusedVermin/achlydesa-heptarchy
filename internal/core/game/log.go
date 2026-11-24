package game

import (
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// LogLines is how many message log lines the view keeps.
const LogLines = 200

// installLog subscribes to every event, to collect the events of the current Apply and to fill the
// message log, and registers the view builders that publish the log and the tick.
func (g *Game) installLog() error {
	g.world.Bus.SubscribeAll("game.log", func(w *sim.World, e api.Event) {
		g.events = append(g.events, e)
		if loggable, ok := e.(api.Loggable); ok {
			g.appendLog(w.Content.Render(loggable.LogText()))
		}
	})
	err := g.world.Views.Register(sim.ViewBuilder{Name: "game.log", Build: func(_ *sim.World, v *api.View) error {
		v.Log = slices.Clone(g.log)
		return nil
	}})
	if err != nil {
		return err
	}
	return g.world.Views.Register(sim.ViewBuilder{Name: "game.tick", Build: func(w *sim.World, v *api.View) error {
		v.Tick = w.Now
		return nil
	}})
}

// appendLog adds a line at the current tick, dropping the oldest beyond LogLines.
func (g *Game) appendLog(text string) {
	g.log = append(g.log, api.LogLine{Tick: g.world.Now, Text: text})
	if len(g.log) > LogLines {
		g.log = g.log[len(g.log)-LogLines:]
	}
}
