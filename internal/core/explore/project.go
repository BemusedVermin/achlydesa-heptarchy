package explore

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// focusOf returns the party the player commands, its members, and the member the view is centered
// on: the current actor if it is a member, otherwise the leader.
func focusOf(w *sim.World) (party ecs.Entity, members []ecs.Entity, focus ecs.Entity, ok bool) {
	party, ok = anatomy.ControlledParty(w)
	if !ok {
		return ecs.Entity{}, nil, ecs.Entity{}, false
	}
	members = membersOf(w, party)
	if len(members) == 0 {
		return ecs.Entity{}, nil, ecs.Entity{}, false
	}
	focus = members[0]
	if leader, found := w.ByUID(
		ecs.NewMap[comp.Party](w.ECS).Get(party).Leader,
	); found &&
		slices.Contains(members, leader) {
		focus = leader
	}
	if current, found := timeline.Current(w); found && slices.Contains(members, current) {
		focus = current
	}
	return party, members, focus, true
}

func buildParty(w *sim.World, v *api.View) error {
	party, members, focus, ok := focusOf(w)
	if !ok {
		return nil
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	appearances := ecs.NewMap[comp.Appearance](w.ECS)
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	leader := state.Leader
	current, _ := timeline.Current(w)
	view := &api.PartyView{UID: w.UIDOf(party), Control: state.Control, Alert: state.Alert}
	if positions.Has(focus) {
		pos := positions.Get(focus)
		view.Map, view.Hex = pos.Map, pos.At.Hex
	}
	for _, member := range members {
		item := api.MemberView{
			UID:     w.UIDOf(member),
			Name:    w.Content.Render(templateName(w, member)),
			Leader:  w.UIDOf(member) == leader,
			Current: member == current,
		}
		if appearances.Has(member) {
			look := appearances.Get(member)
			item.Glyph, item.Color = look.Glyph, look.Color
		}
		view.Members = append(view.Members, item)
	}
	if state.Alert {
		view.Status = append(view.Status, w.Content.Render(api.Text{ID: "text.explore.status.alert"}))
	}
	if playerDecides(w, members) {
		view.Status = append(
			view.Status,
			w.Content.Render(api.Text{ID: api.ContentID("text.explore.status." + state.Control)}),
		)
	}
	v.Party = view
	return nil
}

// column is the cells of one hex worth considering, by depth.
type column map[api.Hex][]int

func (c column) add(cell api.Cell) { c[cell.Hex] = append(c[cell.Hex], cell.Depth) }

// drawn is a candidate cell and how to draw it.
type drawn struct {
	depth int
	glyph string
	color string
	wall  bool
}

// buildMap fills the plan view: for each hex, at most one cell, the candidate nearest the focus
// member in depth. Floor the party can walk to from the focus, by the party's step rules and within
// the plan range, over cells it has seen, wins its hex outright; elsewhere the nearest candidate in
// depth wins. The view holds only how far above or below the party each cell lies.
func buildMap(w *sim.World, v *api.View) error {
	party, _, focus, ok := focusOf(w)
	if !ok {
		return nil
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !positions.Has(focus) {
		return nil
	}
	here := *positions.Get(focus)
	m := w.Maps.Get(here.Map)
	if m == nil {
		return nil
	}
	sight, err := partySight(w, party)
	if err != nil {
		return err
	}
	visible := sight[here.Map]
	visibleColumns, seenColumns := column{}, column{}
	known := make(map[api.Cell]bool, len(visible))
	for cell := range visible {
		visibleColumns.add(cell)
		known[cell] = true
	}
	for _, cell := range seenIn(w, party, here.Map) {
		seenColumns.add(cell)
		known[cell] = true
	}
	reached, err := m.Reachable(here.At, stateOf(w).rules, w.Content.Party().PlanRange,
		func(cell api.Cell) bool { return known[cell] })
	if err != nil {
		return fmt.Errorf("explore: plan view reach: %w", err)
	}
	reachColumns := column{}
	for cell := range reached {
		reachColumns.add(cell)
	}
	hexes := make([]api.Hex, 0, len(visibleColumns)+len(seenColumns))
	for h := range visibleColumns {
		hexes = append(hexes, h)
	}
	for h := range seenColumns {
		if _, dup := visibleColumns[h]; !dup {
			hexes = append(hexes, h)
		}
	}
	slices.SortFunc(hexes, func(a, b api.Hex) int { return cmp.Or(cmp.Compare(a.R, b.R), cmp.Compare(a.Q, b.Q)) })

	view := &api.MapView{Map: here.Map, Name: w.Content.Render(api.Text{ID: stateOf(w).names[here.Map]})}
	if def, found := w.Content.Map(stateOf(w).names[here.Map]); found {
		view.Name = w.Content.Render(api.Text{ID: def.Name})
	}
	for _, h := range hexes {
		best, found, err := nearest(w, m, h, reachColumns[h], here.At.Depth)
		isVisible := found && visible[api.Cell{Hex: h, Depth: best.depth}]
		if err == nil && !found {
			best, found, err = nearest(w, m, h, visibleColumns[h], here.At.Depth)
			isVisible = found
		}
		if err == nil && !found {
			best, found, err = nearest(w, m, h, seenColumns[h], here.At.Depth)
		}
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		view.Cells = append(view.Cells, api.CellView{
			Hex: h, Glyph: best.glyph, Color: best.color, Visible: isVisible, Below: best.depth - here.At.Depth,
		})
	}
	view.Entities = entitiesIn(w, here, visible)
	v.Map = view
	return nil
}

// nearest picks, among the candidate cells at these depths of one hex, the one nearest the focus
// depth, the shallower on a tie.
func nearest(w *sim.World, m *hexmap.Map, h api.Hex, depths []int, focus int) (drawn, bool, error) {
	var best drawn
	found := false
	for _, depth := range depths {
		candidate, ok, err := classify(w, m, api.Cell{Hex: h, Depth: depth})
		if err != nil {
			return drawn{}, false, err
		}
		if !ok {
			continue
		}
		distance, bestDistance := max(depth-focus, focus-depth), max(best.depth-focus, focus-best.depth)
		if !found || distance < bestDistance || (distance == bestDistance && depth < best.depth) {
			best, found = candidate, true
		}
	}
	return best, found, nil
}

// classify says whether a cell can be drawn in the plan view and how. A standable cell is drawn
// with the floor glyph of the solid cell beneath it; a solid cell with a horizontally adjacent open
// cell is drawn with its wall glyph.
func classify(w *sim.World, m *hexmap.Map, cell api.Cell) (drawn, bool, error) {
	terrain, err := m.At(cell)
	if err != nil {
		return drawn{}, false, err
	}
	if terrain.Solid {
		for dir := range api.Dir(api.DirCount) {
			link := m.Topology().Neighbor(cell, dir)
			if link.Map != 0 && link.Map != m.ID() {
				continue
			}
			beside, err := m.At(link.Cell)
			if err != nil {
				return drawn{}, false, err
			}
			if !beside.Solid {
				def := terrainOf(w.Content, terrain.ID)
				return drawn{depth: cell.Depth, glyph: glyphOr(def.WallGlyph), color: def.Color, wall: true}, true, nil
			}
		}
		return drawn{}, false, nil
	}
	standable, err := m.Standable(cell)
	if err != nil || !standable {
		return drawn{}, false, err
	}
	below := m.Topology().Below(cell)
	floor, err := m.At(below.Cell)
	if err != nil {
		return drawn{}, false, err
	}
	def := terrainOf(w.Content, floor.ID)
	return drawn{depth: cell.Depth, glyph: glyphOr(def.FloorGlyph), color: def.Color}, true, nil
}

// glyphOr shows "?" for terrain whose content gives no glyph, so the gap is noticed.
func glyphOr(glyph string) string {
	if glyph == "" {
		return "?"
	}
	return glyph
}

// entitiesIn lists the entities standing in visible cells of the map, by UID.
func entitiesIn(w *sim.World, here comp.Pos, visible map[api.Cell]bool) []api.EntityView {
	identities := ecs.NewMap[comp.Identity](w.ECS)
	query := ecs.NewFilter2[comp.Pos, comp.Appearance](w.ECS).Query()
	var out []api.EntityView
	for query.Next() {
		pos, look := query.Get()
		if pos.Map != here.Map || !visible[pos.At] || !identities.Has(query.Entity()) {
			continue
		}
		out = append(out, api.EntityView{
			UID: identities.Get(query.Entity()).UID, Hex: pos.At.Hex, Glyph: look.Glyph, Color: look.Color,
			Below: pos.At.Depth - here.At.Depth,
		})
	}
	slices.SortFunc(out, func(a, b api.EntityView) int { return cmp.Compare(a.UID, b.UID) })
	return out
}
