package game

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/combat"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// installWipe registers the wake command and the builder of the wipe screen.
func installWipe(w *sim.World, clk *clock.Clock) error {
	err := sim.Handle(w, func(w *sim.World, _ api.Wake) (int64, error) {
		if !combat.PartyWiped(w) || !explore.CanWake(w) {
			w.Bus.Publish(api.Reject(api.WakeKind, "text.game.cannot_wake"))
			return 0, nil
		}
		if err := clk.Advance(w, max(w.Content.Party().WakeTicks, 0)); err != nil {
			return 0, err
		}
		return 0, explore.Wake(w)
	})
	if err != nil {
		return err
	}
	return w.Views.Register(sim.ViewBuilder{
		Name: "game.wipe", After: []string{"timeline", "actions", "explore.control"}, Build: buildWipe,
	})
}

// buildWipe adds the wipe screen while the party is wiped, and leaves the wake option as the only
// option, or none when nobody can wake.
func buildWipe(w *sim.World, v *api.View) error {
	if !combat.PartyWiped(w) {
		return nil
	}
	canWake := explore.CanWake(w)
	line := "text.game.wipe.lost"
	v.Options = nil
	if canWake {
		line = "text.game.wipe.wake"
		v.Options = []api.Option{{
			ID:      api.OptionWake,
			Label:   w.Content.Render(api.Text{ID: "text.game.wipe.option_wake"}),
			Command: api.Wake{},
			Enabled: true,
		}}
	}
	v.Wipe = &api.WipeView{
		Title:   w.Content.Render(api.Text{ID: "text.game.wipe.title"}),
		Lines:   []string{w.Content.Render(api.Text{ID: api.ContentID(line)})},
		CanWake: canWake,
	}
	return nil
}
