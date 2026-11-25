package agents

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// lodStream is the random stream promotion draws from.
const lodStream = "lod"

// demotedReason is the text ID given for actions cut short when their body is counted back.
const demotedReason api.ContentID = "text.agents.demoted"

// StartLevelOfDetail promotes the populations of the map the controlled party starts on. Call it
// once, after the regions are spawned.
func StartLevelOfDetail(w *sim.World) error {
	at, ok := partyMap(w)
	if !ok {
		return nil
	}
	return promote(w, at)
}

func onPartyCrossed(w *sim.World, e api.PartyCrossed) {
	party, ok := anatomy.ControlledParty(w)
	if !ok || w.UIDOf(party) != e.Party {
		return
	}
	if err := demote(w, e.From); err != nil {
		w.Bus.Fail(err)
		return
	}
	if err := promote(w, e.To); err != nil {
		w.Bus.Fail(err)
	}
}

// promote wakes the individuals asleep on the map and turns each population of its region into
// bodies, then sets the counts to zero.
func promote(w *sim.World, id api.MapID) error {
	region, ok := RegionOf(w, id)
	if !ok {
		return nil
	}
	resumeIndividuals(w, id)
	populations := ecs.NewMap[comp.Populations](w.ECS)
	if !populations.Has(region) {
		return nil
	}
	of := populations.Get(region).Of
	if !slices.ContainsFunc(of, func(p comp.Population) bool { return p.Count > 0 }) {
		return nil
	}
	placer, err := newPlacer(w, id)
	if err != nil {
		return err
	}
	stream := w.RNG.Get(lodStream)
	var promoted []api.FamilyCount
	for i := range of {
		population := &of[i]
		if population.Count == 0 {
			continue
		}
		def, ok := w.Content.Family(population.Family)
		if !ok {
			return fmt.Errorf("agents: promote: unknown family %q", population.Family)
		}
		for n := range population.Count {
			template := pickTemplate(def.Templates, stream.IntN(totalWeight(def.Templates)))
			at, err := placer.place(def, n, stream.IntN)
			if err != nil {
				return fmt.Errorf("agents: promote %q: %w", population.Family, err)
			}
			body, err := w.Spawn(template, &comp.Pos{Map: id, At: at})
			if err != nil {
				return fmt.Errorf("agents: promote %q: %w", population.Family, err)
			}
			sim.Put(w, body, comp.Family{ID: population.Family})
		}
		promoted = append(promoted, api.FamilyCount{Family: population.Family, Count: population.Count})
		population.Count = 0
	}
	w.Bus.Publish(api.Promoted{Map: id, Families: promoted})
	return nil
}

func totalWeight(templates []content.Weighted) int {
	total := 0
	for _, t := range templates {
		total += t.Weight
	}
	return total
}

func pickTemplate(templates []content.Weighted, roll int) api.ContentID {
	for _, t := range templates {
		if roll < t.Weight {
			return t.Template
		}
		roll -= t.Weight
	}
	return templates[len(templates)-1].Template
}

// resumeIndividuals puts the map's dormant bodies back on the timeline, in UID order.
func resumeIndividuals(w *sim.World, id api.MapID) {
	var asleep []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Dormant](w.ECS).Query()
	for query.Next() {
		if pos, _ := query.Get(); pos.Map == id {
			asleep = append(asleep, query.Entity())
		}
	}
	for _, body := range w.SortByUID(asleep) {
		timeline.Resume(w, body)
	}
}

// demote counts the map's family bodies back into its region and takes individuals off the
// timeline. Dead bodies stay where they are.
func demote(w *sim.World, id api.MapID) error {
	var bodies []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Family](w.ECS).Query()
	for query.Next() {
		if pos, _ := query.Get(); pos.Map == id {
			bodies = append(bodies, query.Entity())
		}
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	individuals := ecs.NewMap[comp.Individual](w.ECS)
	families := ecs.NewMap[comp.Family](w.ECS)
	counts := map[api.ContentID]int{}
	for _, body := range w.SortByUID(bodies) {
		switch {
		case dead.Has(body):
		case individuals.Has(body):
			timeline.Suspend(w, body, demotedReason)
		default:
			counts[families.Get(body).ID]++
			for _, pending := range timeline.PendingOf(w, w.UIDOf(body)) {
				timeline.Cancel(w, pending, demotedReason)
			}
			w.Despawn(body)
		}
	}
	if len(counts) == 0 {
		return nil
	}
	region, ok := RegionOf(w, id)
	if !ok {
		return fmt.Errorf("agents: demote: map %d has no region", id)
	}
	populations := ecs.NewMap[comp.Populations](w.ECS).Get(region)
	var demoted []api.FamilyCount
	for _, family := range sortedFamilies(counts) {
		addToPopulation(populations, family, counts[family])
		demoted = append(demoted, api.FamilyCount{Family: family, Count: counts[family]})
	}
	w.Bus.Publish(api.Demoted{Map: id, Families: demoted})
	return nil
}

func sortedFamilies(counts map[api.ContentID]int) []api.ContentID {
	out := make([]api.ContentID, 0, len(counts))
	for family := range counts {
		out = append(out, family)
	}
	slices.Sort(out)
	return out
}

// addToPopulation adds to the family's count, keeping Of sorted by family.
func addToPopulation(populations *comp.Populations, family api.ContentID, count int) {
	at, found := slices.BinarySearchFunc(populations.Of, family, func(p comp.Population, f api.ContentID) int {
		return cmp.Compare(p.Family, f)
	})
	if found {
		populations.Of[at].Count += count
		return
	}
	populations.Of = slices.Insert(populations.Of, at, comp.Population{Family: family, Count: count})
}

// placer picks the cells promoted bodies stand on.
type placer struct {
	w        *sim.World
	m        *hexmap.Map
	taken    map[api.Cell]bool
	seen     map[api.Cell]bool
	party    api.Cell
	hasParty bool
	radius   int
	pools    map[api.Cell][]api.Cell
}

func newPlacer(w *sim.World, id api.MapID) (*placer, error) {
	m := w.Maps.Get(id)
	if m == nil {
		return nil, fmt.Errorf("agents: promote: unknown map %d", id)
	}
	p := &placer{
		w: w, m: m, taken: map[api.Cell]bool{}, seen: map[api.Cell]bool{},
		radius: w.Content.Agents().PromoteRadius, pools: map[api.Cell][]api.Cell{},
	}
	query := ecs.NewFilter1[comp.Pos](w.ECS).Query()
	for query.Next() {
		if pos := query.Get(); pos.Map == id {
			p.taken[pos.At] = true
		}
	}
	party, ok := anatomy.ControlledParty(w)
	if !ok {
		return p, nil
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	cfg := w.Content.Party()
	for _, member := range anatomy.Members(w, party) {
		if !positions.Has(member) || positions.Get(member).Map != id {
			continue
		}
		at := positions.Get(member).At
		if !p.hasParty {
			p.party, p.hasParty = at, true
		}
		cells, err := m.Visible(at, cfg.SightRadius, cfg.SightDepth)
		if err != nil {
			return nil, fmt.Errorf("agents: promote: sight: %w", err)
		}
		for _, cell := range cells {
			p.seen[cell] = true
		}
	}
	return p, nil
}

// free lists the free standable cells within the promote radius of the anchor, nearest first, then
// by depth, R and Q.
func (p *placer) free(anchor api.Cell) ([]api.Cell, error) {
	pool, ok := p.pools[anchor]
	if !ok {
		for _, cell := range p.m.CellsIn(anchor.Hex, p.radius, anchor.Depth-1, anchor.Depth+1) {
			standable, err := p.m.Standable(cell)
			if err != nil {
				return nil, err
			}
			if _, gate := p.m.GateAt(cell); standable && !gate {
				pool = append(pool, cell)
			}
		}
		slices.SortFunc(pool, func(a, b api.Cell) int {
			return cmp.Or(
				cmp.Compare(hexmap.CellDistance(anchor, a), hexmap.CellDistance(anchor, b)),
				hexmap.CompareCells(a, b),
			)
		})
		p.pools[anchor] = pool
	}
	out := make([]api.Cell, 0, len(pool))
	for _, cell := range pool {
		if !p.taken[cell] {
			out = append(out, cell)
		}
	}
	return out, nil
}

// place picks a cell for the n-th member of the family: next to one of its marks in turn, or at
// random around the party when it has none on this map. A cell the party sees is chosen only when
// no other is free.
func (p *placer) place(def content.FamilyDef, n int, intN func(int) int) (api.Cell, error) {
	var anchors []api.Cell
	for _, name := range def.Marks {
		if cell, ok := p.m.Mark(name); ok {
			anchors = append(anchors, cell)
		}
	}
	anchor, byMark := p.party, len(anchors) > 0
	if byMark {
		anchor = anchors[n%len(anchors)]
	} else if !p.hasParty {
		return api.Cell{}, errors.New("no party on the map to place around")
	}
	free, err := p.free(anchor)
	if err != nil {
		return api.Cell{}, err
	}
	unseen := make([]api.Cell, 0, len(free))
	for _, cell := range free {
		if !p.seen[cell] {
			unseen = append(unseen, cell)
		}
	}
	if len(unseen) > 0 {
		free = unseen
	}
	if len(free) == 0 {
		return api.Cell{}, fmt.Errorf("no free cell within %d hexes of %v", p.radius, anchor)
	}
	at := free[0]
	if !byMark {
		at = free[intN(len(free))]
	}
	p.taken[at] = true
	return at, nil
}
