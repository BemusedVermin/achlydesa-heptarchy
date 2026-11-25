package agents

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// offscreenSystem is the per-phase system that moves and feeds the regions the party is not in.
const offscreenSystem = "agents.offscreen"

// regionState is one region as it stood at the start of the phase.
type regionState struct {
	entity     ecs.Entity
	id         api.MapID
	env        content.RegionEnv
	town       *content.TownDef
	neighbours []int
}

// transfer is a number of one family moving between regions.
type transfer struct {
	family api.ContentID
	to     int
	count  int
	// short is the item whose shortage sent the residents away; empty for drift.
	short api.ContentID
}

// death is a number of one family dying.
type death struct {
	family api.ContentID
	count  int
}

// regionPlan is everything one region does this phase, worked out from the start-of-phase state.
type regionPlan struct {
	from      int
	deaths    []death
	transfers []transfer
	spent     []comp.ItemStack
	refill    []comp.ItemStack
}

// tickOffscreen plans every region the party is not in from the state at the start of the phase,
// then applies the plans in map ID order, so the order the regions are visited in decides nothing.
func tickOffscreen(w *sim.World) error {
	states, err := snapshotRegions(w)
	if err != nil {
		return err
	}
	away, hasParty := partyMap(w)
	var awayPtr *api.MapID
	if hasParty {
		awayPtr = &away
	}
	stream := w.RNG.Get(lodStream)
	var plans []regionPlan
	for from, state := range states {
		if hasParty && state.id == away {
			continue
		}
		plan, err := planRegion(w, states, from, stream, awayPtr)
		if err != nil {
			return fmt.Errorf("agents: off-screen map %d: %w", state.id, err)
		}
		plans = append(plans, plan)
	}
	for _, plan := range plans {
		applyPlan(w, states, plan)
	}
	return nil
}

// snapshotRegions reads every region, in map ID order, with its neighbours as indices.
func snapshotRegions(w *sim.World) ([]regionState, error) {
	var entities []ecs.Entity
	regions := ecs.NewMap[comp.Region](w.ECS)
	query := ecs.NewFilter1[comp.Region](w.ECS).Query()
	for query.Next() {
		entities = append(entities, query.Entity())
	}
	slices.SortFunc(entities, func(a, b ecs.Entity) int { return cmp.Compare(regions.Get(a).Map, regions.Get(b).Map) })
	dead := map[api.MapID]int{}
	bodies := ecs.NewFilter2[comp.Dead, comp.Pos](w.ECS).Query()
	for bodies.Next() {
		_, pos := bodies.Get()
		dead[pos.Map]++
	}
	populations := ecs.NewMap[comp.Populations](w.ECS)
	inventories := ecs.NewMap[comp.Inventory](w.ECS)
	towns := ecs.NewMap[comp.Town](w.ECS)
	index := map[ecs.Entity]int{}
	states := make([]regionState, len(entities))
	for i, entity := range entities {
		region := regions.Get(entity)
		def, ok := w.Content.Map(region.Def)
		if !ok {
			return nil, fmt.Errorf("agents: region of map %d: unknown map %q", region.Map, region.Def)
		}
		record, _ := w.Content.Record(region.Def)
		state := regionState{
			entity: entity, id: region.Map,
			env: content.RegionEnv{
				Data: record, Counts: map[string]int{}, Stock: map[string]int{}, Dead: dead[region.Map],
			},
		}
		if populations.Has(entity) {
			for _, population := range populations.Get(entity).Of {
				state.env.Counts[string(population.Family)] = population.Count
			}
		}
		if inventories.Has(entity) {
			for _, stack := range inventories.Get(entity).Stacks {
				state.env.Stock[string(stack.Item)] += stack.Count
			}
		}
		if towns.Has(entity) && def.Region != nil {
			state.town = def.Region.Town
		}
		states[i] = state
		index[entity] = i
	}
	for i := range states {
		for _, next := range Neighbours(w, states[i].entity) {
			states[i].neighbours = append(states[i].neighbours, index[next])
		}
	}
	return states, nil
}

// planRegion works out the deaths, moves and stock changes of one region. A family's deaths and
// moves share its count, so none of them can take more than the region holds.
func planRegion(w *sim.World, states []regionState, from int, stream floatSource, away *api.MapID) (regionPlan, error) {
	state := states[from]
	plan := regionPlan{from: from}
	var open []int
	for _, next := range state.neighbours {
		if away == nil || states[next].id != *away {
			open = append(open, next)
		}
	}
	short := planStock(&plan, state)
	for _, family := range slices.Sorted(keysOf(state.env.Counts)) {
		count := state.env.Counts[family]
		if count <= 0 {
			continue
		}
		id := api.ContentID(family)
		remaining := count
		attrition := content.AttritionEnv{Region: state.env, Family: family, Count: count}
		dying, err := w.Content.FamilyAttrition(id, attrition)
		if err != nil {
			return plan, err
		}
		if n := min(roundWith(stream, dying), remaining); n > 0 {
			plan.deaths = append(plan.deaths, death{family: id, count: n})
			remaining -= n
		}
		fractions := make([]float64, len(open))
		total := 0.0
		for i, next := range open {
			env := content.DriftEnv{Region: state.env, Neighbour: states[next].env, Family: family}
			drift, err := w.Content.FamilyDrift(id, env)
			if err != nil {
				return plan, err
			}
			fractions[i] = max(drift, 0)
			total += fractions[i]
		}
		for i, next := range open {
			share := fractions[i]
			if total > 1 {
				share /= total
			}
			if n := min(roundWith(stream, share*float64(count)), remaining); n > 0 {
				plan.transfers = append(plan.transfers, transfer{family: id, to: next, count: n})
				remaining -= n
			}
		}
		if state.town == nil || state.town.Residents != id || short == "" {
			continue
		}
		to, ok := richestTown(states, open, short)
		if !ok {
			continue
		}
		if n := min(roundWith(stream, state.town.Exodus*float64(remaining)), remaining); n > 0 {
			plan.transfers = append(plan.transfers, transfer{family: id, to: to, count: n, short: short})
		}
	}
	return plan, nil
}

// planStock fills in what a town gains and spends, the refill counting toward what it can spend,
// and returns the first used item, in item ID order, that it could not cover.
func planStock(plan *regionPlan, state regionState) api.ContentID {
	if state.town == nil {
		return ""
	}
	residents := state.env.Counts[string(state.town.Residents)]
	use := slices.Clone(state.town.Use)
	slices.SortFunc(use, func(a, b content.StackDef) int { return cmp.Compare(a.Item, b.Item) })
	refilled := map[api.ContentID]int{}
	for _, stack := range state.town.Refill {
		refilled[stack.Item] += stack.Count
		plan.refill = append(plan.refill, comp.ItemStack{Item: stack.Item, Count: stack.Count})
	}
	var short api.ContentID
	for _, stack := range use {
		wanted := stack.Count * residents
		spent := min(wanted, state.env.Stock[string(stack.Item)]+refilled[stack.Item])
		if spent > 0 {
			plan.spent = append(plan.spent, comp.ItemStack{Item: stack.Item, Count: spent})
		}
		if spent < wanted && short == "" {
			short = stack.Item
		}
	}
	return short
}

// richestTown picks the open neighbour that is a town holding the most of the item, the lowest map
// ID among equals; it fails when none holds any.
func richestTown(states []regionState, open []int, item api.ContentID) (int, bool) {
	best, found := 0, false
	for _, next := range open {
		if states[next].town == nil || states[next].env.Stock[string(item)] <= 0 {
			continue
		}
		if !found || states[next].env.Stock[string(item)] > states[best].env.Stock[string(item)] {
			best, found = next, true
		}
	}
	return best, found
}

// applyPlan carries out one region's plan and publishes its events.
func applyPlan(w *sim.World, states []regionState, plan regionPlan) {
	state := states[plan.from]
	populations := ecs.NewMap[comp.Populations](w.ECS).Get(state.entity)
	for _, stack := range plan.refill {
		inventory.Add(w, state.entity, stack)
	}
	for _, stack := range plan.spent {
		inventory.Take(w, state.entity, stack.Item, 0, stack.Count)
	}
	for _, lost := range plan.deaths {
		addToPopulation(populations, lost.family, -lost.count)
		w.Bus.Publish(api.PopulationDied{Map: state.id, Family: lost.family, Count: lost.count})
	}
	for _, move := range plan.transfers {
		to := states[move.to]
		addToPopulation(populations, move.family, -move.count)
		addToPopulation(ecs.NewMap[comp.Populations](w.ECS).Get(to.entity), move.family, move.count)
		if move.short != "" {
			w.Bus.Publish(api.Migrated{
				From: state.id, To: to.id, Family: move.family, Count: move.count, Short: move.short,
			})
		}
	}
	if state.town == nil || state.town.Residents == "" {
		return
	}
	family := state.town.Residents
	if state.env.Counts[string(family)] > 0 && countOf(populations, family) == 0 {
		w.Bus.Publish(api.TownEmptied{Map: state.id, Family: family})
	}
}

func countOf(populations *comp.Populations, family api.ContentID) int {
	for _, p := range populations.Of {
		if p.Family == family {
			return p.Count
		}
	}
	return 0
}

// floatSource is a random stream that draws floats in [0, 1).
type floatSource interface{ Float64() float64 }

// roundWith rounds up with the probability of the fraction, so that the expected value is kept.
func roundWith(stream floatSource, x float64) int {
	if x <= 0 {
		return 0
	}
	whole, fraction := math.Modf(x)
	n := int(whole)
	if fraction > 0 && stream.Float64() < fraction {
		n++
	}
	return n
}

func keysOf[V any](m map[string]V) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		for key := range m {
			if !yield(key) {
				return
			}
		}
	}
}
