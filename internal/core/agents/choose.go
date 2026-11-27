package agents

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// beliefEnv is the state in the form a goal's when formula reads.
func beliefEnv(s State) content.BeliefEnv {
	env := content.BeliefEnv{
		Has: map[string]int{}, Near: map[string]bool{}, Known: map[string]bool{}, Met: map[string]bool{},
	}
	for _, fact := range s {
		tag := string(fact.Tag)
		switch fact.Kind {
		case content.FactHas:
			env.Has[tag] = fact.N
		case content.FactNear:
			env.Near[tag] = true
		case content.FactKnown:
			env.Known[tag] = true
		case content.FactDriveMet:
			env.Met[tag] = true
		}
	}
	return env
}

// Choose picks the goal the body works toward: the applicable goal with the highest weight, the
// body's current goal counting agents.keep_bonus more. A goal whose drive the body does not carry
// never applies. Ties go to the lowest goal ID. It returns "" when no goal applies.
func Choose(w *sim.World, body ecs.Entity) (api.ContentID, error) {
	beliefs, err := StateOf(w, body)
	if err != nil {
		return "", err
	}
	env := beliefEnv(beliefs)
	current := ecs.NewMap[comp.Mind](w.ECS).Get(body).Goal
	var carried *comp.Drives
	if carriers := ecs.NewMap[comp.Drives](w.ECS); carriers.Has(body) {
		carried = carriers.Get(body)
	}
	var best api.ContentID
	bestScore := 0.0
	for _, id := range w.Content.GoalIDs() {
		goal, _ := w.Content.Goal(id)
		if id == w.Content.Agents().ChatGoal {
			if _, can := chatPartner(w, body); !can {
				continue
			}
		}
		var goalEnv content.GoalEnv
		if goal.Drive != "" {
			if carried == nil {
				continue
			}
			level, has := carried.Levels[goal.Drive]
			if !has {
				continue
			}
			goalEnv = content.GoalEnv{Level: level, Weight: carried.Weights[goal.Drive]}
			if goalEnv.Urge, err = w.Content.DriveUrge(
				goal.Drive,
				content.DriveEnv{Level: level, Weight: goalEnv.Weight},
			); err != nil {
				return "", fmt.Errorf("agents: body %d: %w", w.UIDOf(body), err)
			}
		}
		applies, err := w.Content.GoalApplies(id, env)
		if err != nil {
			return "", fmt.Errorf("agents: body %d: %w", w.UIDOf(body), err)
		}
		if !applies {
			continue
		}
		score, err := w.Content.GoalWeight(id, goalEnv)
		if err != nil {
			return "", fmt.Errorf("agents: body %d: %w", w.UIDOf(body), err)
		}
		if id == current {
			score += w.Content.Agents().KeepBonus
		}
		if best == "" || score > bestScore {
			best, bestScore = id, score
		}
	}
	return best, nil
}

// think chooses the body's goal when it has none or agents.think_ticks have passed since it last did.
func think(w *sim.World, body ecs.Entity) error {
	mind := ecs.NewMap[comp.Mind](w.ECS).Get(body)
	if mind.Goal != "" && w.Now-mind.Thought < int64(w.Content.Agents().ThinkTicks) {
		return nil
	}
	goal, err := Choose(w, body)
	if err != nil {
		return err
	}
	mind.Goal, mind.Thought = goal, w.Now
	return nil
}
