package agents

import (
	"errors"
	"fmt"
	"hash/fnv"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const (
	// MindController is the controller name templates use, as actor.controller, for bodies that
	// decide by goal and plan.
	MindController = "mind"
	// SelfPicker is the target picker that aims a step's action at the agent itself.
	SelfPicker  = "self"
	planStream  = "agents.plan"
	mindSpawner = "mind"
)

// Snapshot is what a planner may read: the agent's beliefs and the content, copied out of the world
// so that planning can run on another goroutine.
type Snapshot struct {
	Agent   api.UID
	Tick    int64
	Beliefs State
	Content *content.Registry
	// Rand is the planner's own sequence, derived from the agent's UID and the tick, so planning draws
	// no shared stream.
	Rand *rng.Generator
}

// Planner makes a plan for a goal from a snapshot alone. Plan may run on several goroutines at once
// and must change no shared state. It reports false when it finds none. A plan with no steps counts
// as none.
type Planner interface {
	Plan(snapshot Snapshot, goal api.ContentID) (Plan, bool)
}

// Picker aims a step's action: it returns the intent with its target, direction or choices filled
// in, and false when there is nothing to aim at, which fails the step. The mind controller sets the
// intent's actor and action.
type Picker interface {
	Pick(w *sim.World, body ecs.Entity, step Step) (actions.Intent, bool, error)
}

// Finisher is a picker for a step that takes many actions: the step repeats until Finished says it
// is done, then the plan moves on.
type Finisher interface {
	Picker
	Finished(w *sim.World, body ecs.Entity, step Step) bool
}

type selfPicker struct{}

func (selfPicker) Pick(_ *sim.World, body ecs.Entity, _ Step) (actions.Intent, bool, error) {
	return actions.Intent{Target: body}, true, nil
}

// minds is the mind controller's registries, kept on the world as an Ark resource.
type minds struct {
	planners map[string]Planner
	pickers  map[string]Picker
}

func mindsOf(w *sim.World) *minds {
	res := ecs.NewResource[minds](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &minds{planners: map[string]Planner{}, pickers: map[string]Picker{}}
	res.Add(st)
	return st
}

// RegisterPlanner makes p available to templates as mind.planner. Names are unique.
func RegisterPlanner(w *sim.World, name string, p Planner) error {
	st := mindsOf(w)
	switch _, dup := st.planners[name]; {
	case name == "" || p == nil:
		return errors.New("agents: a planner needs a name and an implementation")
	case dup:
		return fmt.Errorf("agents: planner %q already registered", name)
	}
	st.planners[name] = p
	return nil
}

// RegisterPicker makes p available to operators as does.target. Names are unique.
func RegisterPicker(w *sim.World, name string, p Picker) error {
	st := mindsOf(w)
	switch _, dup := st.pickers[name]; {
	case name == "" || p == nil:
		return errors.New("agents: a picker needs a name and an implementation")
	case dup:
		return fmt.Errorf("agents: picker %q already registered", name)
	}
	st.pickers[name] = p
	return nil
}

// ValidateContent checks that every planner templates name and every picker operators name is
// registered. Run it once every package and hook is installed.
func ValidateContent(w *sim.World) error {
	st := mindsOf(w)
	var errs []error
	for _, id := range w.Content.RecordIDs(content.CollectionTemplates) {
		template, _ := w.Content.Template(id)
		if template.Mind == nil {
			if template.Actor != nil && template.Actor.Controller == MindController {
				errs = append(errs, fmt.Errorf("templates[%q].mind: required by the %q controller", id, MindController))
			}
			continue
		}
		if _, ok := st.planners[template.Mind.Planner]; !ok {
			errs = append(errs, fmt.Errorf("templates[%q].mind.planner: unknown planner %q", id, template.Mind.Planner))
		}
	}
	usesGOAP := false
	for _, id := range w.Content.RecordIDs(content.CollectionTemplates) {
		template, _ := w.Content.Template(id)
		usesGOAP = usesGOAP || (template.Mind != nil && template.Mind.Planner == GOAPPlanner)
	}
	for _, id := range w.Content.GeographyIDs() {
		geography, _ := w.Content.Geography(id)
		for i, spawn := range geography.Spawns {
			if spawn.Mind == nil {
				continue
			}
			usesGOAP = usesGOAP || spawn.Mind.Planner == GOAPPlanner
			if _, ok := st.planners[spawn.Mind.Planner]; !ok {
				errs = append(errs, fmt.Errorf("geographies[%q].spawns[%d].mind.planner: unknown planner %q",
					id, i, spawn.Mind.Planner))
			}
		}
	}
	if cfg := w.Content.Agents(); usesGOAP && (cfg.GoapNodes < 1 || cfg.GoapDepth < 1 || cfg.GoapCache < 1) {
		errs = append(errs, fmt.Errorf("agents.goap_nodes, goap_depth and goap_cache: must be at least 1 when "+
			"a template or spawn uses the %q planner, got %d, %d, %d", GOAPPlanner, cfg.GoapNodes, cfg.GoapDepth, cfg.GoapCache))
	}
	for _, id := range w.Content.OperatorIDs() {
		op, _ := w.Content.Operator(id)
		if _, ok := st.pickers[op.Does.Target]; !ok {
			errs = append(errs, fmt.Errorf("operators[%q].does.target: unknown picker %q", id, op.Does.Target))
		}
	}
	return errors.Join(errs...)
}

func installMind(w *sim.World) error {
	err := w.RegisterSpawner(
		sim.Spawner{Name: mindSpawner, Apply: func(w *sim.World, e ecs.Entity, t content.Template) error {
			if t.Mind != nil {
				ecs.NewMap[comp.Mind](w.ECS).Add(e, &comp.Mind{Planner: t.Mind.Planner})
			}
			return nil
		}},
	)
	if err != nil {
		return err
	}
	if err := RegisterPicker(w, SelfPicker, selfPicker{}); err != nil {
		return err
	}
	if err := RegisterPlanner(w, GOAPPlanner, newGOAP()); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.PlanMade](w.Codec); err != nil {
		return err
	}
	return timeline.RegisterController(w, MindController, mindController{})
}

// mindController holds a plan for its body and starts the plan's steps one at a time, each once the
// last has finished with the body's channels. A plan that stops holding is dropped and made again.
type mindController struct{}

func (mindController) Decide(w *sim.World, body ecs.Entity) error {
	if !w.Alive(body) || ecs.NewMap[comp.Dead](w.ECS).Has(body) {
		return nil
	}
	minds := ecs.NewMap[comp.Mind](w.ECS)
	if !minds.Has(body) {
		return fmt.Errorf("agents: body %d has the mind controller but its template has no mind", w.UIDOf(body))
	}
	for _, c := range timeline.Channels(w, body) {
		if !timeline.Free(w, c) {
			return nil
		}
	}
	if !Near(w, body) {
		demote(w, body)
		return nil
	}
	if err := think(w, body); err != nil {
		return err
	}
	if chat := w.Content.Agents().ChatGoal; chat != "" && minds.Get(body).Goal == chat {
		drop(w, body)
		started, err := startChat(w, body)
		if !started {
			minds.Get(body).Goal = ""
		}
		return err
	}
	if mind := minds.Get(body); len(mind.Plan.Steps) > 0 && (mind.Step >= len(mind.Plan.Steps) || !planHolds(w, body)) {
		drop(w, body)
	}
	if len(minds.Get(body).Plan.Steps) == 0 {
		if made, err := makePlan(w, body); err != nil || !made {
			return err
		}
	}
	started, err := startStep(w, body)
	if err != nil {
		return err
	}
	if !started {
		drop(w, body)
	}
	return nil
}

// planHolds says whether the body's plan still stands: its goal is still the body's, no export has
// begun since it was made, and it still holds every claim it made on a thing that still exists.
func planHolds(w *sim.World, body ecs.Entity) bool {
	mind := ecs.NewMap[comp.Mind](w.ECS).Get(body)
	if mind.Plan.Goal != mind.Goal || mind.Exports != w.ExportCount {
		return false
	}
	var held []comp.Claim
	if claims := ecs.NewMap[comp.Claims](w.ECS); claims.Has(body) {
		held = claims.Get(body).Held
	}
	if !slices.Equal(held, mind.Plan.Claims) {
		return false
	}
	for _, c := range held {
		if target, ok := w.ByUID(c.Target); !ok || !w.Alive(target) {
			return false
		}
	}
	return true
}

// drop ends the body's plan and the claims it held.
func drop(w *sim.World, body ecs.Entity) {
	mind := ecs.NewMap[comp.Mind](w.ECS).Get(body)
	mind.Plan, mind.Step = comp.Plan{}, 0
	if claims := ecs.NewMap[comp.Claims](w.ECS); claims.Has(body) {
		claims.Remove(body)
	}
}

// makePlan asks the body's planner for a plan to its goal and claims what the plan needs. It reports
// false, leaving the body without a plan, when there is no goal, no plan, or a claim clashes with
// another agent's; the body tries again at its next decision.
func makePlan(w *sim.World, body ecs.Entity) (bool, error) {
	mind := *ecs.NewMap[comp.Mind](w.ECS).Get(body)
	if mind.Goal == "" {
		return false, nil
	}
	planner, snapshot, err := planInput(w, body)
	if err != nil {
		return false, err
	}
	plan, found := planner.Plan(snapshot, mind.Goal)
	return adopt(w, body, plan, found)
}

// planInput is the body's planner and the snapshot it plans from.
func planInput(w *sim.World, body ecs.Entity) (Planner, Snapshot, error) {
	mind := ecs.NewMap[comp.Mind](w.ECS).Get(body)
	planner, ok := mindsOf(w).planners[mind.Planner]
	if !ok {
		return nil, Snapshot{}, fmt.Errorf("agents: body %d has unknown planner %q", w.UIDOf(body), mind.Planner)
	}
	beliefs, err := StateOf(w, body)
	if err != nil {
		return nil, Snapshot{}, err
	}
	uid := w.UIDOf(body)
	return planner, Snapshot{
		Agent:   uid,
		Tick:    w.Now,
		Beliefs: beliefs,
		Content: w.Content,
		Rand:    w.RNG.Derived(planStream, planSalt(uid, w.Now)),
	}, nil
}

// adopt gives the body a plan its planner found for its goal, if the claims it makes can be had.
func adopt(w *sim.World, body ecs.Entity, plan Plan, found bool) (bool, error) {
	if !found || len(plan.Steps) == 0 {
		return false, nil
	}
	minds := ecs.NewMap[comp.Mind](w.ECS)
	plan.Goal = minds.Get(body).Goal
	if claimed, err := claim(w, body, plan.Claims); err != nil || !claimed {
		return false, err
	}
	stored := minds.Get(body)
	stored.Plan, stored.Step, stored.Exports = plan, 0, w.ExportCount
	if w.Content.Agents().TracePlans {
		steps := make([]api.ContentID, len(plan.Steps))
		for i, step := range plan.Steps {
			steps[i] = step.Operator
		}
		w.Bus.Publish(api.PlanMade{Agent: w.UIDOf(body), Goal: plan.Goal, Steps: steps, Tick: w.Now})
	}
	return true, nil
}

func planSalt(uid api.UID, tick int64) uint64 {
	hasher := fnv.New64a()
	_, _ = fmt.Fprintf(hasher, "%d@%d", uid, tick)
	return hasher.Sum64()
}

type claimKey struct {
	target api.UID
	item   api.ContentID
}

// claim gives the body the claims if no other living agent holds them, and reports whether it did.
// Bodies decide one at a time in UID order, so at one tick the lower UID wins a clash.
func claim(w *sim.World, body ecs.Entity, wanted []comp.Claim) (bool, error) {
	if len(wanted) == 0 {
		return true, nil
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	totals := map[claimKey]int{}
	whole, claimed := map[api.UID]bool{}, map[api.UID]bool{}
	note := func(c comp.Claim) {
		totals[claimKey{c.Target, c.Item}] += c.N
		whole[c.Target] = whole[c.Target] || c.Item == ""
		claimed[c.Target] = true
	}
	query := ecs.NewFilter1[comp.Claims](w.ECS).Query()
	for query.Next() {
		if holder := query.Entity(); holder == body || dead.Has(holder) {
			continue
		}
		for _, c := range query.Get().Held {
			note(c)
		}
	}
	for _, c := range wanted {
		target, ok := w.ByUID(c.Target)
		switch {
		case c.Item != "" && c.N < 1:
			return false, fmt.Errorf("agents: body %d's plan claims %d of %q", w.UIDOf(body), c.N, c.Item)
		case !ok || !w.Alive(target) || whole[c.Target] || (c.Item == "" && claimed[c.Target]):
			return false, nil
		}
		note(c)
		if c.Item == "" {
			continue
		}
		if totals[claimKey{c.Target, c.Item}] > stockOf(w, target, c.Item) {
			return false, nil
		}
	}
	held := &comp.Claims{Held: slices.Clone(wanted)}
	if claims := ecs.NewMap[comp.Claims](w.ECS); claims.Has(body) {
		claims.Get(body).Held = held.Held
	} else {
		claims.Add(body, held)
	}
	return true, nil
}

// stockOf is how many of the item a thing holds, in its source and its inventory.
func stockOf(w *sim.World, thing ecs.Entity, item api.ContentID) int {
	n := inventory.Count(w, thing, item, 0)
	if sources := ecs.NewMap[comp.Source](w.ECS); sources.Has(thing) && sources.Get(thing).Item == item {
		n += sources.Get(thing).Stock
	}
	return n
}

// startStep starts the plan's next step as a timeline action, and reports false when it cannot: a
// belief the step needs is contradicted, the picker finds nothing to aim at, or the action is refused.
func startStep(w *sim.World, body ecs.Entity) (bool, error) {
	skipFinished(w, body)
	mind := *ecs.NewMap[comp.Mind](w.ECS).Get(body)
	if mind.Step >= len(mind.Plan.Steps) {
		return false, nil
	}
	step := mind.Plan.Steps[mind.Step]
	op, ok := w.Content.Operator(step.Operator)
	if !ok {
		return false, fmt.Errorf("agents: body %d's plan names unknown operator %q", w.UIDOf(body), step.Operator)
	}
	beliefs, err := StateOf(w, body)
	if err != nil || !Holds(op, beliefs) {
		return false, err
	}
	picker, ok := mindsOf(w).pickers[op.Does.Target]
	if !ok {
		return false, fmt.Errorf("agents: operator %q names unknown picker %q", step.Operator, op.Does.Target)
	}
	intent, ok, err := picker.Pick(w, body, step)
	if err != nil || !ok {
		return false, err
	}
	intent.Actor, intent.Action = body, op.Does.Action
	if ok, _ := actions.Validate(w, intent); !ok {
		return false, nil
	}
	if err := actions.Start(w, intent); err != nil {
		return false, err
	}
	if _, repeats := picker.(Finisher); !repeats {
		ecs.NewMap[comp.Mind](w.ECS).Get(body).Step++
	}
	return true, nil
}

// skipFinished moves the body past the steps whose picker says they are done.
func skipFinished(w *sim.World, body ecs.Entity) {
	mind := ecs.NewMap[comp.Mind](w.ECS).Get(body)
	for mind.Step < len(mind.Plan.Steps) {
		step := mind.Plan.Steps[mind.Step]
		op, ok := w.Content.Operator(step.Operator)
		if !ok {
			return
		}
		finisher, ok := mindsOf(w).pickers[op.Does.Target].(Finisher)
		if !ok || !finisher.Finished(w, body, step) {
			return
		}
		mind.Step++
	}
}
