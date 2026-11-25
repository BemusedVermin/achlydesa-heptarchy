package harness

import (
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// Camping thresholds. They are placeholders for M7's tuning.
const (
	DefaultCampingShare = 0.9
	DefaultCampingDays  = 10
)

// CampingConfig says what degenerate camping is: a family with more than Share of its living
// members in one town for more than Days days in a row.
type CampingConfig struct {
	Share float64
	Days  int
}

// DefaultCamping returns the placeholder thresholds.
func DefaultCamping() CampingConfig {
	return CampingConfig{Share: DefaultCampingShare, Days: DefaultCampingDays}
}

// Camp is one family found camping in one town.
type Camp struct {
	Family api.ContentID
	Map    api.MapID
	// FromDay is the first day of the streak; ToDay the last day seen.
	FromDay, ToDay int
}

type campKey struct {
	family api.ContentID
	mapID  api.MapID
}

// camping follows each family's share of its living members in each town.
type camping struct {
	cfg     CampingConfig
	streaks map[campKey]int
	found   map[campKey]Camp
}

func newCamping(cfg CampingConfig) *camping {
	return &camping{cfg: cfg, streaks: map[campKey]int{}, found: map[campKey]Camp{}}
}

// observe takes the census on a day. A streak survives only while every sample of its days is over
// the share, so a day is judged by the last sample before the day turns.
func (c *camping) observe(day int, census []game.FamilyCount) {
	totals := map[api.ContentID]int{}
	for _, entry := range census {
		totals[entry.Family] += entry.Count
	}
	over := map[campKey]bool{}
	for _, entry := range census {
		if entry.Town && float64(entry.Count) > c.cfg.Share*float64(totals[entry.Family]) {
			over[campKey{entry.Family, entry.Map}] = true
		}
	}
	for key := range c.streaks {
		if !over[key] {
			delete(c.streaks, key)
		}
	}
	for key := range over {
		start, ok := c.streaks[key]
		if !ok {
			start = day
			c.streaks[key] = start
		}
		if day-start+1 > c.cfg.Days {
			camp := c.found[key]
			camp.Family, camp.Map, camp.FromDay, camp.ToDay = key.family, key.mapID, start, day
			c.found[key] = camp
		}
	}
}

func (c *camping) camps() []Camp {
	out := make([]Camp, 0, len(c.found))
	for _, camp := range c.found {
		out = append(out, camp)
	}
	slices.SortFunc(out, func(a, b Camp) int {
		if a.Map != b.Map {
			return int(a.Map) - int(b.Map)
		}
		return compareIDs(a.Family, b.Family)
	})
	return out
}

func compareIDs(a, b api.ContentID) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// EmergenceResult says what happened in one seeded run of the living deep.
type EmergenceResult struct {
	Seed         uint64
	Steps        int
	LastDay      int
	Deaths       int
	Migrations   int
	TownsEmptied int
	Camps        []Camp
}

// counts tallies a run's deaths, migrations and emptied towns from its events.
type counts struct {
	deaths, migrations, townsEmptied int
}

func (c *counts) observe(events []api.Event) {
	for _, event := range events {
		switch e := event.(type) {
		case api.BodyDied:
			c.deaths++
		case api.PopulationDied:
			c.deaths += e.Count
		case api.Migrated:
			c.migrations += e.Count
		case api.TownEmptied:
			c.townsEmptied++
		}
	}
}

// PlayDays plays one seeded game until the clock passes days or maxSteps commands have been sent,
// and checks it for camping.
func PlayDays(reg *content.Registry, seed uint64, policy Policy, days, maxSteps int, cfg CampingConfig) (EmergenceResult, error) {
	g, err := game.New(game.Config{Seed: seed, Content: reg})
	if err != nil {
		return EmergenceResult{}, err
	}
	if binder, ok := policy.(Binder); ok {
		binder.Bind(g)
	}
	var tally counts
	watch := newCamping(cfg)
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
		tally.observe(events)
		view, err := g.View()
		if err != nil {
			return EmergenceResult{}, err
		}
		if view.Clock == nil {
			continue
		}
		result.LastDay = view.Clock.Day
		watch.observe(view.Clock.Day, g.FamilyCensus())
		if view.Clock.Day > days {
			break
		}
	}
	result.Deaths, result.Migrations, result.TownsEmptied = tally.deaths, tally.migrations, tally.townsEmptied
	result.Camps = watch.camps()
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
