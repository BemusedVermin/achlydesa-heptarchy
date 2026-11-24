package explore

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Install registers stepping, the maps, party control and their views. It must run after effects
// and timeline.
func Install(w *sim.World) error {
	if w.Content == nil {
		return errors.New("explore: install: world has no content")
	}
	stateOf(w)
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.Moved],
		api.RegisterEvent[api.Fell],
		api.RegisterEvent[api.ControlChanged],
		api.RegisterEvent[api.LeaderChanged],
		api.RegisterEvent[api.PartyWoke],
	} {
		if err := register(w.Codec); err != nil {
			return err
		}
	}
	err := effects.Register(w, effects.Primitive{Name: "step", Apply: applyStep, Check: checkStep})
	if err != nil {
		return err
	}
	effects.SetMaps(w, w.Maps.Get)
	timeline.SetVisible(w, visibleBody)
	if err := installControl(w); err != nil {
		return err
	}
	if err := w.Views.Register(sim.ViewBuilder{Name: "explore.party", Build: buildParty}); err != nil {
		return err
	}
	return w.Views.Register(sim.ViewBuilder{
		Name: "explore.map", After: []string{"timeline", "explore.party"}, Build: buildMap,
	})
}

// onMap returns the body's position and map. Bodies that are not on a map (test fixtures, or bodies
// in transit) report false and step in place.
func onMap(w *sim.World, body ecs.Entity) (*comp.Pos, *hexmap.Map, bool) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !w.Alive(body) || !positions.Has(body) {
		return nil, nil, false
	}
	pos := positions.Get(body)
	m := w.Maps.Get(pos.Map)
	return pos, m, m != nil
}

func bump() api.Text { return api.Text{ID: "text.explore.bump"} }

// checkStep fails with text.explore.bump when the hexmap rules offer no step that way.
func checkStep(ctx *effects.Context, _ effects.Params) (bool, api.Text) {
	pos, m, ok := onMap(ctx.World, ctx.Source)
	if !ok {
		return true, api.Text{}
	}
	if ctx.Dir == nil {
		return false, bump()
	}
	step, can, err := m.Step(pos.At, *ctx.Dir, stateOf(ctx.World).rules)
	if err != nil {
		return false, api.Text{ID: "text.explore.broken", Args: map[string]string{"error": err.Error()}}
	}
	if !can {
		return false, bump()
	}
	if _, occupied, swappable := occupant(ctx.World, ctx.Source, pos, step); occupied && !swappable {
		return false, api.Text{ID: "text.explore.occupied"}
	}
	return true, api.Text{}
}

// occupant finds the body in the cell a step leads to. A step onto another member of the same party
// swaps the two, unless it is a drop or leaves the map, which cannot be reversed; a step onto anyone
// else is refused.
func occupant(
	w *sim.World,
	source ecs.Entity,
	pos *comp.Pos,
	step hexmap.Step,
) (other ecs.Entity, occupied, swappable bool) {
	dest := pos.Map
	if step.To.Map != 0 {
		dest = step.To.Map
	}
	other, occupied = bodyAt(w, dest, step.To.Cell, source)
	if !occupied {
		return ecs.Entity{}, false, false
	}
	party, member := anatomy.PartyOf(w, source)
	otherParty, otherMember := anatomy.PartyOf(w, other)
	swappable = member && otherMember && party == otherParty && step.Kind != hexmap.StepDrop && dest == pos.Map
	return other, true, swappable
}

func setTrail(w *sim.World, e ecs.Entity, cell api.Cell) {
	if trails := ecs.NewMap[comp.Trail](w.ECS); trails.Has(e) {
		trails.Get(e).Last = cell
	}
}

// reverse is the kind of step that undoes a step of this kind.
func reverse(kind string) string {
	switch kind {
	case api.MoveUp:
		return api.MoveDown
	case api.MoveDown:
		return api.MoveUp
	}
	return kind
}

// swapWith moves other into the cell the source just left, and publishes the move. The source has
// already moved to dest.
func swapWith(w *sim.World, other ecs.Entity, from, dest api.Cell, kind string) error {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	pos := positions.Get(other)
	pos.At = from
	setTrail(w, other, dest)
	w.Bus.Publish(api.Moved{Source: w.UIDOf(other), Map: pos.Map, From: dest, To: from, Kind: reverse(kind)})
	party, _ := anatomy.PartyOf(w, other)
	cells, err := sightOf(w, *pos)
	if err != nil {
		return err
	}
	markSeen(w, party, pos.Map, cells)
	return nil
}

var moveKinds = map[hexmap.StepKind]string{
	hexmap.StepLevel: api.MoveLevel,
	hexmap.StepUp:    api.MoveUp,
	hexmap.StepDown:  api.MoveDown,
	hexmap.StepDrop:  api.MoveDrop,
}

// applyStep moves the body at once, publishes what happened, records the new sight, and then keeps
// the channels the action uses busy for the cost of the ground and of any climb. Moving first and
// paying after lets a body step out of reach of a swing that has not landed yet.
func applyStep(ctx *effects.Context, _ effects.Params) error {
	w := ctx.World
	pos, m, ok := onMap(w, ctx.Source)
	if !ok {
		return nil
	}
	if ctx.Dir == nil {
		return errors.New("explore: step needs a direction")
	}
	step, can, err := m.Step(pos.At, *ctx.Dir, stateOf(w).rules)
	switch {
	case err != nil:
		return fmt.Errorf("explore: step: %w", err)
	case !can:
		return errors.New("explore: step is blocked")
	}
	other, occupied, swappable := occupant(w, ctx.Source, pos, step)
	if occupied && !swappable {
		return errors.New("explore: step is blocked by a body")
	}
	from := pos.At
	if step.To.Map != 0 {
		pos.Map = step.To.Map
		if m = w.Maps.Get(pos.Map); m == nil {
			return fmt.Errorf("explore: step leads to unknown map %d", pos.Map)
		}
	}
	pos.At = step.To.Cell
	setTrail(w, ctx.Source, from)
	actor := w.UIDOf(ctx.Source)
	w.Bus.Publish(api.Moved{Source: actor, Map: pos.Map, From: from, To: pos.At, Kind: moveKinds[step.Kind]})
	if occupied {
		if err := swapWith(w, other, from, pos.At, moveKinds[step.Kind]); err != nil {
			return err
		}
	}
	if step.Kind == hexmap.StepDrop {
		w.Bus.Publish(api.Fell{Body: actor, Cells: step.Fall})
	}
	if party, member := anatomy.PartyOf(w, ctx.Source); member {
		cells, err := sightOf(w, *pos)
		if err != nil {
			return err
		}
		markSeen(w, party, pos.Map, cells)
	}
	cost, err := stepCost(w, ctx.Source, m, pos.At, step.Kind)
	if err != nil {
		return err
	}
	if cost < 1 {
		return nil
	}
	for _, channel := range ctx.Used {
		if err := timeline.Extend(w, channel, cost); err != nil {
			return fmt.Errorf("explore: step: %w", err)
		}
	}
	return nil
}

// stepCost is the ticks beyond the action's own recovery: the move cost of the floor under the
// destination, plus the climb cost of a step up.
func stepCost(w *sim.World, body ecs.Entity, m *hexmap.Map, to api.Cell, kind hexmap.StepKind) (int64, error) {
	var cost int64
	if floor, local := m.Topology().Below(to), true; local {
		terrain, err := m.At(floor.Cell)
		if err != nil {
			return 0, err
		}
		if _, known := w.Content.Terrain(terrain.ID); known && terrain.Solid {
			if cost, err = w.Content.MoveCost(terrain.ID, content.MoveEnv{}); err != nil {
				return 0, err
			}
		}
	}
	if kind == hexmap.StepUp {
		speed, err := timeline.Speed(w, body)
		if err != nil {
			return 0, err
		}
		climb, err := w.Content.ClimbCost(content.SpeedEnv{Speed: speed})
		if err != nil {
			return 0, err
		}
		cost += climb
	}
	return cost, nil
}

// visibleBody is the timeline's visibility hook: party members, and any body standing in a cell
// the party can see now.
func visibleBody(w *sim.World, body ecs.Entity) bool {
	if ecs.NewMap[comp.MemberOf](w.ECS).Has(body) {
		return true
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	party, ok := anatomy.ControlledParty(w)
	if !ok || !w.Alive(body) || !positions.Has(body) {
		return false
	}
	sight, err := partySight(w, party)
	if err != nil {
		return false
	}
	pos := positions.Get(body)
	return sight[pos.Map][pos.At]
}
