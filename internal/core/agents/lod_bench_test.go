package agents_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	crowdNear      = 100
	crowdFar       = 900
	crowdWells     = 16
	crowdNearWells = 4
)

var crowdParty = api.Cell{Hex: api.Hex{Q: 200, R: 200}}

// crowdPlanner sends each thirsty agent to the well it was told of and has it drink there; an agent with
// nothing to do wanders.
type crowdPlanner struct{ wells map[api.UID]agents.Target }

func (p crowdPlanner) Plan(snapshot agents.Snapshot, goal api.ContentID) (agents.Plan, bool) {
	drink := agents.Step{Operator: farDrinkOp}
	well, knows := p.wells[snapshot.Agent]
	switch {
	case goal != drinkGoal || !knows:
		return agents.Plan{Steps: []agents.Step{{Operator: wanderOp}}}, true
	case slices.Contains(snapshot.Beliefs, agents.Fact{Kind: content.FactNear, Tag: waterMark}):
		return agents.Plan{Steps: []agents.Step{drink}}, true
	}
	return agents.Plan{Steps: []agents.Step{{Operator: goToOp, Target: &well}, drink}}, true
}

// crowdCell spreads the i-th cell over a square of side hexes around center, the same every run. A far cell
// is moved out of reach of the party.
func crowdCell(center api.Cell, i, side int, far bool) api.Cell {
	offset := (i * 7919) % (side * side)
	cell := api.Cell{Hex: api.Hex{Q: center.Hex.Q - side/2 + offset%side, R: center.Hex.R - side/2 + offset/side}}
	if far && hexmap.Distance(cell.Hex, center.Hex) < 4*hexmap.ChunkSize {
		cell.Hex.Q += 8 * hexmap.ChunkSize
	}
	return cell
}

// BenchmarkPhaseWithFarAgents plays one phase at a time of 1,000 minds, one in ten near the party and the
// rest far from it, on an open floor with wells about.
func BenchmarkPhaseWithFarAgents(b *testing.B) {
	wells := map[api.UID]agents.Target{}
	g, _, _ := gametest.New(b, 1, farContent, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			setGround(w, floorGround())
			return agents.RegisterPlanner(w, farPlanner, crowdPlanner{wells: wells})
		}
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error {
			placeParty(w, crowdParty)
			var marks []agents.Target
			for i := range crowdWells {
				at := crowdCell(crowdParty, i*61, 360, true)
				if i < crowdNearWells {
					at = crowdCell(crowdParty, 3+i*97, 2*hexmap.ChunkSize, false)
				}
				well, err := w.Spawn("template.source.well", &comp.Pos{At: at})
				if err != nil {
					return err
				}
				marks = append(marks, agents.Target{Subject: w.UIDOf(well), At: at})
			}
			for i := range crowdNear + crowdFar {
				near := i < crowdNear
				at := crowdCell(crowdParty, i, 360, true)
				mark := marks[i%crowdWells]
				if near {
					at, mark = crowdCell(crowdParty, i, 2*hexmap.ChunkSize, false), marks[i%crowdNearWells]
				}
				body, err := w.Spawn(farAgent, &comp.Pos{At: at})
				if err != nil {
					return err
				}
				if i%2 == 0 {
					teach(w, body, mark.Subject, mark.At)
					wells[w.UIDOf(body)] = mark
					ecs.NewMap[comp.Drives](w.ECS).Get(body).Levels[water] = 10
				}
			}
			return nil
		}
	})
	for b.Loop() {
		if _, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitPhase}); err != nil {
			b.Fatal(err)
		}
	}
}
