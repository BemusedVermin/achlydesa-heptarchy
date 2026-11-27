package agents

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Fact is one thing an agent believes, in the form a planner reasons over. N is the count of a has
// fact and 0 for any other. It names no UID, so states stay small and equal across agents.
type Fact struct {
	Kind content.FactKind
	Tag  api.ContentID
	N    int
}

// State is what an agent believes, sorted by kind then tag, each kind and tag at most once, and no
// has fact with a count below 1. Apply returns a new State and never changes the one it is given.
type State []Fact

// Key is a string equal for equal states and different otherwise, for keying caches.
func (s State) Key() string {
	var b strings.Builder
	for _, fact := range s {
		fmt.Fprintf(&b, "%s(%s,%d);", fact.Kind, fact.Tag, fact.N)
	}
	return b.String()
}

func compareFacts(a, b Fact) int {
	return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Tag, b.Tag))
}

func (s State) find(kind content.FactKind, tag api.ContentID) (int, bool) {
	return slices.BinarySearchFunc(s, Fact{Kind: kind, Tag: tag}, compareFacts)
}

// Holds says whether every precondition of the operator holds in the state.
func Holds(op content.Operator, s State) bool { return holdsAll(op.Pre, s) }

func holdsAll(conds []content.Condition, s State) bool {
	for _, cond := range conds {
		at, found := s.find(cond.Fact, cond.Tag)
		holds := found
		if cond.Fact == content.FactHas {
			holds = found && s[at].N >= cond.N
		}
		if holds == cond.Not {
			return false
		}
	}
	return true
}

// Apply is the state after the operator's effects, in the order the operator lists them. It does not
// check the preconditions: see Holds.
func Apply(op content.Operator, s State) State {
	out := slices.Clone(s)
	for _, eff := range op.Eff {
		at, found := out.find(eff.Fact, eff.Tag)
		switch {
		case eff.Fact == content.FactHas:
			count := eff.N
			if found {
				count += out[at].N
			}
			switch {
			case count < 1 && found:
				out = slices.Delete(out, at, at+1)
			case count < 1:
			case found:
				out[at].N = count
			default:
				out = slices.Insert(out, at, Fact{Kind: eff.Fact, Tag: eff.Tag, N: count})
			}
		case eff.Not && found:
			out = slices.Delete(out, at, at+1)
		case !eff.Not && !found:
			out = slices.Insert(out, at, Fact{Kind: eff.Fact, Tag: eff.Tag})
		}
	}
	return out
}

// StateOf is what the body believes: the item tags it carries, each counting every carried item with the tag, the landmarks it remembers and which of
// them it believes are within agents.near_radius, and the drives agents.met counts as met. Landmarks
// come from its memory, so a well it has never seen or been told of is not in it.
func StateOf(w *sim.World, body ecs.Entity) (State, error) {
	if !w.Alive(body) {
		return nil, nil
	}
	var s State
	if inventories := ecs.NewMap[comp.Inventory](w.ECS); inventories.Has(body) {
		counts := map[api.ContentID]int{}
		for _, stack := range inventories.Get(body).Stacks {
			item, ok := w.Content.Item(stack.Item)
			if !ok {
				return nil, fmt.Errorf("agents: body %d carries unknown item %q", w.UIDOf(body), stack.Item)
			}
			for _, tag := range item.Tags {
				counts[api.ContentID(tag)] += stack.Count
			}
		}
		for _, tag := range slices.Sorted(maps.Keys(counts)) {
			if counts[tag] > 0 {
				s = append(s, Fact{Kind: content.FactHas, Tag: tag, N: counts[tag]})
			}
		}
	}
	if positions := ecs.NewMap[comp.Pos](w.ECS); positions.Has(body) {
		from := positions.Get(body).At
		for _, tag := range w.Content.RecordIDs(content.CollectionAffordances) {
			landmarks, err := Known(w, body, tag)
			if err != nil {
				return nil, err
			}
			if len(landmarks) == 0 {
				continue
			}
			s = append(s, Fact{Kind: content.FactKnown, Tag: tag})
			if hexmap.CellDistance(from, landmarks[0].At) <= w.Content.Agents().NearRadius {
				s = append(s, Fact{Kind: content.FactNear, Tag: tag})
			}
		}
	}
	if carriers := ecs.NewMap[comp.Drives](w.ECS); carriers.Has(body) {
		drives := carriers.Get(body)
		for _, id := range slices.Sorted(maps.Keys(drives.Levels)) {
			met, err := w.Content.DriveMet(content.DriveEnv{Level: drives.Levels[id], Weight: drives.Weights[id]})
			if err != nil {
				return nil, fmt.Errorf("agents: body %d, drive %q: %w", w.UIDOf(body), id, err)
			}
			if met {
				s = append(s, Fact{Kind: content.FactDriveMet, Tag: id})
			}
		}
	}
	slices.SortFunc(s, compareFacts)
	return s, nil
}
