package harness

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// EmergenceResult says what happened in one seeded run of the living deep.
type EmergenceResult struct {
	Seed         uint64
	Steps        int
	LastDay      int
	Deaths       int
	Migrations   int
	TownsEmptied int
}

// PlayDays plays one seeded game until the clock passes days or maxSteps commands have been sent.
func PlayDays(reg *content.Registry, seed uint64, policy Policy, days, maxSteps int) (EmergenceResult, error) {
	g, err := game.New(game.Config{Seed: seed, Content: reg})
	if err != nil {
		return EmergenceResult{}, err
	}
	if binder, ok := policy.(Binder); ok {
		binder.Bind(g)
	}
	track := newTracker()
	result := EmergenceResult{Seed: seed}
	for result.Steps < maxSteps {
		events, ok, err := Step(g, policy)
		if err != nil {
			return EmergenceResult{}, fmt.Errorf("step %d: %w", result.Steps+1, err)
		}
		if !ok {
			break
		}
		result.Steps++
		view, err := g.View()
		if err != nil {
			return EmergenceResult{}, err
		}
		track.observe(events, view)
		if view.Clock == nil {
			continue
		}
		result.LastDay = view.Clock.Day
		if view.Clock.Day > days {
			break
		}
	}
	result.Deaths, result.Migrations, result.TownsEmptied = track.deaths+track.offScreen, track.migrations, track.towns
	return result, nil
}

// RestPolicy lets days pass: it rests when the view offers a rest, and otherwise waits for the next
// phase.
type RestPolicy struct{ reg *content.Registry }

// NewRestPolicy returns a RestPolicy that finds the rest action in reg.
func NewRestPolicy(reg *content.Registry) *RestPolicy { return &RestPolicy{reg: reg} }

// Choose implements [Policy].
func (p *RestPolicy) Choose(view api.View) (api.Option, bool) {
	var wait api.Option
	var haveWait bool
	for _, option := range view.Options {
		if !option.Enabled {
			continue
		}
		switch cmd := option.Command.(type) {
		case api.Do:
			if def, ok := p.reg.Action(cmd.Action); ok && def.Key == "rest" {
				return option, true
			}
		case api.Wait:
			if option.ID == api.OptionWaitPhase {
				wait, haveWait = option, true
			}
		}
	}
	return wait, haveWait
}
