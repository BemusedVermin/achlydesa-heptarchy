package agents

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// rngStream is the stream agents draw from: ties between behaviors, and where to wander.
const rngStream = "agents"

// Primitive is what a behavior does, in Go. Content composes primitives into behaviors.
type Primitive struct {
	Name string
	// Params are the names of the params a behavior using it must give, and the only ones it may.
	Params []string
	// Step does one decision's worth of the behavior: start an action, or report done. Doing
	// neither lets the body wait until the next entry on the timeline.
	Step func(w *sim.World, body ecs.Entity, b content.BehaviorDef, target Target) (done bool, err error)
}

// Target is what a behavior goes for. Body is zero unless the behavior's target is a body; Cell is
// where the target is.
type Target struct {
	Body ecs.Entity
	Cell api.Cell
}

type state struct {
	primitives map[string]Primitive
}

func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{primitives: map[string]Primitive{}}
	res.Add(st)
	return st
}

// RegisterPrimitive makes p available to behaviors as p.Name. Names are unique.
func RegisterPrimitive(w *sim.World, p Primitive) error {
	if p.Name == "" || p.Step == nil {
		return errors.New("agents: a behavior primitive needs a name and a Step function")
	}
	st := stateOf(w)
	if _, dup := st.primitives[p.Name]; dup {
		return fmt.Errorf("agents: behavior primitive %q already registered", p.Name)
	}
	st.primitives[p.Name] = p
	return nil
}

// ValidateContent checks what the loader cannot know: that every behavior names a registered
// primitive and gives exactly the params it declares. It reports every problem.
func ValidateContent(w *sim.World) error {
	st := stateOf(w)
	var problems []error
	for _, id := range w.Content.BehaviorIDs() {
		def, _ := w.Content.Behavior(id)
		p, ok := st.primitives[def.Do]
		if !ok {
			problems = append(problems, fmt.Errorf("behaviors[%q].do: unknown behavior primitive %q", id, def.Do))
			continue
		}
		for _, name := range p.Params {
			if _, given := def.Params[name]; !given {
				problems = append(problems, fmt.Errorf("behaviors[%q].params: %q needs %q", id, def.Do, name))
			}
		}
		for _, name := range slices.Sorted(maps.Keys(def.Params)) {
			if !slices.Contains(p.Params, name) {
				problems = append(problems, fmt.Errorf("behaviors[%q].params: %q takes no %q", id, def.Do, name))
			}
		}
	}
	return errors.Join(problems...)
}

// choice is a behavior an agent could take, with its target and score.
type choice struct {
	id     api.ContentID
	def    content.BehaviorDef
	target Target
	score  float64
}

// chooser weighs one agent's behaviors at one decision. It finds the map's bodies once, however
// many behaviors look for one.
type chooser struct {
	w      *sim.World
	body   ecs.Entity
	pos    comp.Pos
	doing  comp.Doing
	source content.BodyEnv
	phase  string
	bodies []ecs.Entity
	listed bool
}

func newChooser(w *sim.World, body ecs.Entity) *chooser {
	calendar := w.Content.Calendar()
	c := &chooser{
		w: w, body: body,
		pos:    *ecs.NewMap[comp.Pos](w.ECS).Get(body),
		source: effects.BodyEnvOf(w, body, ecs.Entity{}),
		phase:  string(calendar.Phases[clock.New(calendar).Phase(w.Now)].ID),
	}
	if doing := ecs.NewMap[comp.Doing](w.ECS); doing.Has(body) {
		c.doing = *doing.Get(body)
	}
	return c
}

// choose scores each behavior and returns the best: the current one gets agents.keep_bonus, and
// ties are broken by the agents stream. A behavior scoring 0 or less is never chosen, and one whose
// target is not found is not scored.
func (c *chooser) choose(ids []api.ContentID) (choice, bool, error) {
	keep := c.w.Content.Agents().KeepBonus
	var best []choice
	for _, id := range ids {
		def, ok := c.w.Content.Behavior(id)
		if !ok {
			continue
		}
		target, found, err := c.resolve(id, def)
		if err != nil {
			return choice{}, false, fmt.Errorf("behavior %q: %w", id, err)
		}
		if !found {
			continue
		}
		score, err := c.w.Content.BehaviorScore(id, c.env(def, target))
		if err != nil {
			return choice{}, false, err
		}
		if score <= 0 {
			continue
		}
		if id == c.doing.Behavior {
			score += keep
		}
		switch {
		case len(best) == 0 || score > best[0].score:
			best = []choice{{id: id, def: def, target: target, score: score}}
		case score == best[0].score:
			best = append(best, choice{id: id, def: def, target: target, score: score})
		}
	}
	switch len(best) {
	case 0:
		return choice{}, false, nil
	case 1:
		return best[0], true, nil
	}
	return best[c.w.RNG.Get(rngStream).IntN(len(best))], true, nil
}

func (c *chooser) env(def content.BehaviorDef, target Target) content.ConsiderEnv {
	env := content.ConsiderEnv{
		Source: c.source, Target: effects.BodyEnvOf(c.w, target.Body, ecs.Entity{}),
		Phase: c.phase, Now: c.w.Now, Doing: string(c.doing.Behavior),
	}
	if def.Target != nil {
		env.Distance = hexmap.CellDistance(c.pos.At, target.Cell)
	}
	return env
}

// resolve finds the behavior's target. The current behavior keeps the body or cell it went for
// while that still matches, so an agent does not turn to every body that comes a little nearer.
func (c *chooser) resolve(id api.ContentID, def content.BehaviorDef) (Target, bool, error) {
	if def.Target == nil {
		return Target{}, true, nil
	}
	filter := *def.Target
	current := id == c.doing.Behavior
	m := c.w.Maps.Get(c.pos.Map)
	if m == nil {
		return Target{}, false, nil
	}
	switch filter.Kind {
	case content.BehaviorTargetBody:
		if current {
			if other, ok := c.w.ByUID(c.doing.Target); ok && c.matches(other, filter) {
				return Target{Body: other, Cell: ecs.NewMap[comp.Pos](c.w.ECS).Get(other).At}, true, nil
			}
		}
		other, ok := c.nearestBody(filter)
		if !ok {
			return Target{}, false, nil
		}
		return Target{Body: other, Cell: ecs.NewMap[comp.Pos](c.w.ECS).Get(other).At}, true, nil
	case content.BehaviorTargetMark:
		cell, ok := m.Mark(filter.Mark)
		return Target{Cell: cell}, ok, nil
	case content.BehaviorTargetCell:
		if current {
			t, err := m.At(c.doing.Cell)
			if err != nil {
				return Target{}, false, err
			}
			if t.ID == filter.Terrain {
				return Target{Cell: c.doing.Cell}, true, nil
			}
		}
		return c.nearestCell(m, filter.Terrain)
	}
	return Target{}, false, fmt.Errorf("unknown target kind %q", filter.Kind)
}

// matches says whether another body on the agent's map passes the filter.
func (c *chooser) matches(other ecs.Entity, filter content.TargetDef) bool {
	w := c.w
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if other == c.body || !w.Alive(other) || !positions.Has(other) || positions.Get(other).Map != c.pos.Map {
		return false
	}
	if filter.Dead != nil && ecs.NewMap[comp.Dead](w.ECS).Has(other) != *filter.Dead {
		return false
	}
	if filter.Hostile != nil && anatomy.Hostile(w, c.body, other) != *filter.Hostile {
		return false
	}
	if len(filter.Tags) > 0 {
		template, ok := w.Content.Template(w.TemplateOf(other))
		if !ok {
			return false
		}
		for _, tag := range filter.Tags {
			if !slices.Contains(template.Tags, tag) {
				return false
			}
		}
	}
	return true
}

// nearestBody is the matching body nearest by cell distance, then lowest UID.
func (c *chooser) nearestBody(filter content.TargetDef) (ecs.Entity, bool) {
	if !c.listed {
		query := ecs.NewFilter2[comp.Pos, comp.Identity](c.w.ECS).Query()
		for query.Next() {
			if pos, _ := query.Get(); pos.Map == c.pos.Map {
				c.bodies = append(c.bodies, query.Entity())
			}
		}
		c.w.SortByUID(c.bodies)
		c.listed = true
	}
	positions := ecs.NewMap[comp.Pos](c.w.ECS)
	var best ecs.Entity
	bestDistance := 0
	for _, other := range c.bodies {
		if !c.matches(other, filter) {
			continue
		}
		distance := hexmap.CellDistance(c.pos.At, positions.Get(other).At)
		if best.IsZero() || distance < bestDistance {
			best, bestDistance = other, distance
		}
	}
	return best, !best.IsZero()
}

// nearestCell is the cell of the terrain within agents.cell_radius nearest by cell distance, then
// in cell order.
func (c *chooser) nearestCell(m *hexmap.Map, terrain api.ContentID) (Target, bool, error) {
	radius := c.w.Content.Agents().CellRadius
	var best []api.Cell
	bestDistance := 0
	for _, cell := range m.CellsIn(c.pos.At.Hex, radius, c.pos.At.Depth-radius, c.pos.At.Depth+radius) {
		t, err := m.At(cell)
		if err != nil {
			return Target{}, false, err
		}
		if t.ID != terrain {
			continue
		}
		distance := hexmap.CellDistance(c.pos.At, cell)
		switch {
		case len(best) == 0 || distance < bestDistance:
			best, bestDistance = []api.Cell{cell}, distance
		case distance == bestDistance:
			best = append(best, cell)
		}
	}
	if len(best) == 0 {
		return Target{}, false, nil
	}
	return Target{Cell: slices.MinFunc(best, hexmap.CompareCells)}, true, nil
}
