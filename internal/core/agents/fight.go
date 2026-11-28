package agents

import (
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const (
	// EngagePicker is the target picker that hands the body to the fighting controller.
	EngagePicker = "engage"
	// FleePicker is the target picker that steps away from the hostiles the agent believes near.
	FleePicker = "flee"
	// FightController is the name combat registers its hunting controller under.
	FightController = "hunt"
	rethinkSystem   = "agents.rethink"
)

func installFight(w *sim.World) error {
	if err := RegisterPicker(w, EngagePicker, engagePicker{}); err != nil {
		return err
	}
	if err := RegisterPicker(w, FleePicker, fleePicker{}); err != nil {
		return err
	}
	return w.Systems.Register(sim.System{Name: rethinkSystem, Cadence: sim.PerEntry, Run: rethink})
}

// hostileBelief is a body the agent believes is hostile and near, and the cell it believes it in.
type hostileBelief struct {
	subject api.UID
	at      api.Cell
}

// believedHostiles lists the bodies the body believes hostile and near, in UID order: seen within
// agents.hostile_ticks, within sight radius of where it saw them, and not seen dead.
func believedHostiles(w *sim.World, body ecs.Entity) ([]hostileBelief, error) {
	cfg := w.Content.Agents()
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if cfg.AtClaim == "" || cfg.HostileTicks <= 0 || !w.Alive(body) || !positions.Has(body) {
		return nil, nil
	}
	from := positions.Get(body).At
	var out []hostileBelief
	seen := map[api.UID]bool{}
	for _, fact := range factsOf(w, body, cfg.AtClaim) {
		if seen[fact.Subject] {
			continue
		}
		seen[fact.Subject] = true
		other, ok := w.ByUID(fact.Subject)
		if !ok || other == body || !w.Alive(other) || anatomy.OutOfFight(w, other) || !anatomy.Hostile(w, body, other) {
			continue
		}
		best, _ := knowledge.Best(w, body, fact.Subject, cfg.AtClaim)
		if w.Now-best.Tick > int64(cfg.HostileTicks) {
			continue
		}
		if cfg.DeadClaim != "" {
			if _, dead := knowledge.Best(w, body, fact.Subject, cfg.DeadClaim); dead {
				continue
			}
		}
		at, err := knowledge.ParseCell(best.Value)
		if err != nil {
			return nil, err
		}
		if hexmap.CellDistance(from, at) <= cfg.SightRadius {
			out = append(out, hostileBelief{subject: fact.Subject, at: at})
		}
	}
	slices.SortFunc(out, func(a, b hostileBelief) int { return int(a.subject) - int(b.subject) })
	return out, nil
}

// engagePicker hands the body to the hunt controller. It starts no action itself, so its step fails
// and the plan is dropped; the hunter takes over from the next decision.
type engagePicker struct{}

func (engagePicker) Pick(w *sim.World, body ecs.Entity, _ Step) (actions.Intent, bool, error) {
	if err := sense(w, body); err != nil {
		return actions.Intent{}, false, err
	}
	foes, err := believedHostiles(w, body)
	if err != nil || len(foes) == 0 {
		return actions.Intent{}, false, err
	}
	return actions.Intent{}, false, timeline.DelegateTo(w, body, FightController)
}

// fleePicker steps to the open neighbouring cell farthest from the nearest believed hostile. A body
// with no cell that gets it farther fights instead.
type fleePicker struct{}

func (fleePicker) Pick(w *sim.World, body ecs.Entity, _ Step) (actions.Intent, bool, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if w.Map == nil || !positions.Has(body) {
		return actions.Intent{}, false, nil
	}
	if err := sense(w, body); err != nil {
		return actions.Intent{}, false, err
	}
	foes, err := believedHostiles(w, body)
	if err != nil || len(foes) == 0 {
		return actions.Intent{}, false, err
	}
	from := positions.Get(body).At
	gap := func(cell api.Cell) int {
		nearest := -1
		for _, foe := range foes {
			if d := hexmap.CellDistance(cell, foe.at); nearest < 0 || d < nearest {
				nearest = d
			}
		}
		return nearest
	}
	held := cellsHeldByOthers(w, body)
	rules := actions.StepRules(w)
	best, bestDir, found := gap(from), api.Dir(0), false
	for dir := range api.Dir(api.DirCount) {
		step, can, err := w.Map.Step(from, dir, rules)
		if err != nil {
			return actions.Intent{}, false, fmt.Errorf("agents: flee: %w", err)
		}
		if !can || held[step.To.Cell] {
			continue
		}
		if d := gap(step.To.Cell); d > best {
			best, bestDir, found = d, dir, true
		}
	}
	if !found {
		return actions.Intent{}, false, timeline.DelegateTo(w, body, FightController)
	}
	return actions.Intent{Dir: &bestDir}, true, nil
}

// Finished says the flight is over once no hostile is believed near.
func (fleePicker) Finished(w *sim.World, body ecs.Entity, _ Step) bool {
	foes, err := believedHostiles(w, body)
	return err == nil && len(foes) == 0
}

// rethink lets a delegated mind choose its goal again every agents.think_ticks, and takes the body
// back from its delegate when the goal changes, so a fight can end in flight and a flight in a fight.
func rethink(w *sim.World) error {
	dead := ecs.NewMap[comp.Dead](w.ECS)
	far := ecs.NewMap[comp.Far](w.ECS)
	minds := ecs.NewMap[comp.Mind](w.ECS)
	var bodies []ecs.Entity
	query := ecs.NewFilter2[comp.Mind, comp.Delegate](w.ECS).Query()
	for query.Next() {
		bodies = append(bodies, query.Entity())
	}
	for _, body := range w.SortByUID(bodies) {
		mind := minds.Get(body)
		if dead.Has(body) || far.Has(body) || w.Now-mind.Thought < int64(w.Content.Agents().ThinkTicks) {
			continue
		}
		previous := mind.Goal
		if err := think(w, body); err != nil {
			return err
		}
		if mind.Goal != previous {
			drop(w, body)
			timeline.Resume(w, body)
		}
	}
	return nil
}
