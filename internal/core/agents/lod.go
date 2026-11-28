package agents

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"sync"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

const (
	farSystem = "agents.far"
	// farRouteNodes bounds the local areas a far route searches, so a destination walled off from the
	// agent is given up on rather than searched for across the whole world.
	farRouteNodes = 2000
	// placeReach bounds how far from its believed cell a promoted agent looks for ground to stand on.
	placeReach = 2 * hexmap.ChunkSize
)

// lod is what the far system remembers between entries. It is only a cache: a world loaded from a save
// starts without it and works it out again, to the same result.
type lod struct {
	valid   bool
	members []hexmap.ChunkKey
	// due is the earliest far decision, while dueKnown.
	due      int64
	dueKnown bool
}

func lodOf(w *sim.World) *lod {
	res := ecs.NewResource[lod](w.ECS)
	if !res.Has() {
		res.Add(&lod{})
	}
	return res.Get()
}

func installLOD(w *sim.World) error {
	return w.Systems.Register(sim.System{Name: farSystem, Cadence: sim.PerEntry, Run: runFar})
}

// Near reports whether the body is within world.near_chunks chunks of a party member, by the chunk it
// stands in. A body with no position counts as near, and every body counts as far while no party member
// stands anywhere.
func Near(w *sim.World, body ecs.Entity) bool {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !positions.Has(body) {
		return true
	}
	return nearMembers(w, hexmap.KeyOf(positions.Get(body).At), memberChunks(w))
}

func nearMembers(w *sim.World, key hexmap.ChunkKey, members []hexmap.ChunkKey) bool {
	radius := w.Content.World().NearChunks
	return slices.ContainsFunc(members, func(m hexmap.ChunkKey) bool { return hexmap.ChunkDistance(key, m) <= radius })
}

// memberChunks lists the chunks party members stand in, each once, in key order.
func memberChunks(w *sim.World) []hexmap.ChunkKey {
	var out []hexmap.ChunkKey
	query := ecs.NewFilter2[comp.MemberOf, comp.Pos](w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		out = append(out, hexmap.KeyOf(pos.At))
	}
	slices.SortFunc(out, hexmap.CompareKeys)
	return slices.Compact(out)
}

// demote takes a body that has walked out of reach of the party off the timeline. It drops its plan,
// which a far decision makes again at goal level at once.
func demote(w *sim.World, body ecs.Entity) {
	drop(w, body)
	sim.Put(w, body, comp.Far{Decides: w.Now})
	timeline.Park(w, body)
	lodOf(w).dueKnown = false
}

// promote puts a far body that has come near the party on the free standable cell nearest the one it is
// believed in, and back on the timeline to decide cell by cell. Its far plan is dropped.
func promote(w *sim.World, body ecs.Entity) error {
	pos := ecs.NewMap[comp.Pos](w.ECS).Get(body)
	at, err := placeNear(w, body, pos.At)
	if err != nil {
		return fmt.Errorf("agents: promoting body %d: %w", w.UIDOf(body), err)
	}
	ecs.NewMap[comp.Pos](w.ECS).Get(body).At = at
	if trails := ecs.NewMap[comp.Trail](w.ECS); trails.Has(body) {
		trails.Get(body).Last = at
	}
	ecs.NewMap[comp.Far](w.ECS).Remove(body)
	drop(w, body)
	timeline.Unpark(w, body)
	return nil
}

// placeNear finds the free standable cell nearest from by [hexmap.CellDistance], ties by cell order.
func placeNear(w *sim.World, body ecs.Entity, from api.Cell) (api.Cell, error) {
	if w.Map == nil {
		return from, nil
	}
	held := cellsHeldByOthers(w, body)
	for radius := range placeReach {
		var ring []api.Cell
		for depth := from.Depth - radius; depth <= from.Depth+radius; depth++ {
			ring = appendRing(
				ring,
				api.Cell{Hex: from.Hex, Depth: depth},
				radius-max(depth-from.Depth, from.Depth-depth),
			)
		}
		slices.SortFunc(ring, hexmap.CompareCells)
		for _, cell := range ring {
			ok, err := w.Map.Standable(cell)
			if err != nil {
				return api.Cell{}, err
			}
			if ok && !held[cell] {
				return cell, nil
			}
		}
	}
	return api.Cell{}, fmt.Errorf("no open ground within %d cells of %v", placeReach, from)
}

// runFar promotes the far bodies the party has come near, then lets each far body whose turn has come
// decide. It runs after every timeline entry but does little until the party changes chunk or a far
// decision is due.
func runFar(w *sim.World) error {
	st := lodOf(w)
	members := memberChunks(w)
	if !st.valid || !slices.Equal(members, st.members) {
		st.members, st.valid = members, true
		for _, body := range farBodies(w) {
			if nearMembers(w, hexmap.KeyOf(ecs.NewMap[comp.Pos](w.ECS).Get(body).At), members) {
				if err := promote(w, body); err != nil {
					return err
				}
			}
		}
		st.dueKnown = false
	}
	if st.dueKnown && w.Now < st.due {
		return nil
	}
	bodies := farBodies(w)
	fars := ecs.NewMap[comp.Far](w.ECS)
	var due []ecs.Entity
	for _, body := range bodies {
		if fars.Get(body).Decides <= w.Now {
			due = append(due, body)
		}
	}
	if err := decideFar(w, due, members); err != nil {
		return err
	}
	st.due, st.dueKnown = math.MaxInt64, true
	query := ecs.NewFilter1[comp.Far](w.ECS).Query()
	for query.Next() {
		st.due = min(st.due, query.Get().Decides)
	}
	return nil
}

// farBodies lists the far bodies with a position, in UID order.
func farBodies(w *sim.World) []ecs.Entity {
	var out []ecs.Entity
	query := ecs.NewFilter2[comp.Far, comp.Pos](w.ECS).Query()
	for query.Next() {
		out = append(out, query.Entity())
	}
	return w.SortByUID(out)
}

// planJob is one far body's planning, done on another goroutine.
type planJob struct {
	body     ecs.Entity
	planner  Planner
	snapshot Snapshot
	goal     api.ContentID
	plan     Plan
	found    bool
}

// decideFar has each due body arrive where its route has brought it, choose a goal, and act on its plan:
// the plans it needs are made on other goroutines from snapshots, then adopted in UID order.
func decideFar(w *sim.World, due []ecs.Entity, members []hexmap.ChunkKey) error {
	if len(due) == 0 {
		return nil
	}
	fars := ecs.NewMap[comp.Far](w.ECS)
	minds := ecs.NewMap[comp.Mind](w.ECS)
	dead := ecs.NewMap[comp.Dead](w.ECS)
	marks := landmarkIndex(w)
	next := w.Now + int64(w.Content.Agents().FarTicks)
	var deciding []ecs.Entity
	var jobs []planJob
	for _, body := range due {
		if dead.Has(body) || !minds.Has(body) {
			fars.Remove(body)
			continue
		}
		fars.Get(body).Decides = next
		promoted, err := arrive(w, body, members, marks)
		if err != nil {
			return err
		}
		if promoted || len(fars.Get(body).Hops) > 0 {
			continue
		}
		if err := think(w, body); err != nil {
			return err
		}
		mind := minds.Get(body)
		if len(mind.Plan.Steps) > 0 && (mind.Step >= len(mind.Plan.Steps) || !planHolds(w, body)) {
			drop(w, body)
		}
		deciding = append(deciding, body)
		if mind := minds.Get(body); len(mind.Plan.Steps) == 0 && mind.Goal != "" {
			planner, snapshot, err := planInput(w, body)
			if err != nil {
				return err
			}
			jobs = append(jobs, planJob{body: body, planner: planner, snapshot: snapshot, goal: mind.Goal})
		}
	}
	planAll(jobs)
	planned := map[ecs.Entity]*planJob{}
	for i := range jobs {
		planned[jobs[i].body] = &jobs[i]
	}
	for _, body := range deciding {
		if job, ok := planned[body]; ok {
			if _, err := adopt(w, body, job.plan, job.found); err != nil {
				return err
			}
		}
		if err := runFarSteps(w, body); err != nil {
			return err
		}
	}
	return nil
}

// planAll runs the planners on several goroutines. Each job reads only its own snapshot.
func planAll(jobs []planJob) {
	count := min(runtime.GOMAXPROCS(0), len(jobs))
	var wg sync.WaitGroup
	for i := range count {
		from, to := i*len(jobs)/count, (i+1)*len(jobs)/count
		wg.Go(func() {
			for j := from; j < to; j++ {
				jobs[j].plan, jobs[j].found = jobs[j].planner.Plan(jobs[j].snapshot, jobs[j].goal)
			}
		})
	}
	wg.Wait()
}

// landmarkIndex is every landmark that stands somewhere, by the chunk it stands in, each chunk's in UID
// order: what a far body learns on reaching a chunk, in place of looking.
func landmarkIndex(w *sim.World) map[hexmap.ChunkKey][]sighting {
	deaths := ecs.NewMap[comp.Dead](w.ECS)
	var all []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Affordance](w.ECS).Query()
	for query.Next() {
		all = append(all, query.Entity())
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	affordances := ecs.NewMap[comp.Affordance](w.ECS)
	index := map[hexmap.ChunkKey][]sighting{}
	for _, mark := range w.SortByUID(all) {
		at := positions.Get(mark).At
		index[hexmap.KeyOf(at)] = append(index[hexmap.KeyOf(at)], sighting{
			subject: w.UIDOf(mark), at: at, dead: deaths.Has(mark), gives: affordances.Get(mark).Tags,
		})
	}
	return index
}

// arrive moves the body through the stops of its route it has reached by now, learning the landmarks of
// each chunk it reaches. It reports true when a stop brought it near the party and it was promoted.
func arrive(
	w *sim.World,
	body ecs.Entity,
	members []hexmap.ChunkKey,
	marks map[hexmap.ChunkKey][]sighting,
) (bool, error) {
	cfg := w.Content.Agents()
	uid := w.UIDOf(body)
	for {
		far := ecs.NewMap[comp.Far](w.ECS).Get(body)
		if len(far.Hops) == 0 || far.Hops[0].Arrive > w.Now {
			return false, nil
		}
		at := far.Hops[0].At
		far.Hops = slices.Delete(far.Hops, 0, 1)
		ecs.NewMap[comp.Pos](w.ECS).Get(body).At = at
		for _, mark := range marks[hexmap.KeyOf(at)] {
			if mark.subject != uid {
				learnSighting(w, body, cfg, mark)
			}
		}
		if nearMembers(w, hexmap.KeyOf(at), members) {
			return true, promote(w, body)
		}
	}
}

// runFarSteps carries out the body's plan at goal level: a step to a cell starts a route over the chunk
// graph and waits for it, a wandering step heads for a neighboring chunk, and any other step does its
// action's effects at once, with no cell path and no timeline. A step that cannot be done drops the plan.
func runFarSteps(w *sim.World, body ecs.Entity) error {
	minds := ecs.NewMap[comp.Mind](w.ECS)
	for {
		skipFinished(w, body)
		mind := minds.Get(body)
		if mind.Step >= len(mind.Plan.Steps) {
			drop(w, body)
			return nil
		}
		step := mind.Plan.Steps[mind.Step]
		op, ok := w.Content.Operator(step.Operator)
		if !ok {
			return fmt.Errorf("agents: body %d's plan names unknown operator %q", w.UIDOf(body), step.Operator)
		}
		beliefs, err := StateOf(w, body)
		if err != nil {
			return err
		}
		if !Holds(op, beliefs) {
			drop(w, body)
			return nil
		}
		var done bool
		switch op.Does.Target {
		case GoToPicker:
			done, err = travel(w, body, step)
		case WanderPicker:
			done, err = wanderFar(w, body)
			if err == nil && done {
				minds.Get(body).Step++
			}
		default:
			done, err = act(w, body, step, op)
			if err == nil && done {
				minds.Get(body).Step++
				continue
			}
		}
		if err != nil {
			return err
		}
		if !done {
			drop(w, body)
		}
		return nil
	}
}

// graphOf is the world's chunk graph, or nil when it has none.
func graphOf(w *sim.World) *worldgen.Graph {
	res := ecs.NewResource[worldgen.Graph](w.ECS)
	if !res.Has() {
		return nil
	}
	return res.Get()
}

// travel routes the body to the step's target cell over the chunk graph, a stop at each chunk on the
// way, arriving as fast as its steps would carry it along the route's cost. It reports false when no
// route is found.
func travel(w *sim.World, body ecs.Entity, step Step) (bool, error) {
	graph := graphOf(w)
	if step.Target == nil || graph == nil {
		return false, nil
	}
	from := ecs.NewMap[comp.Pos](w.ECS).Get(body).At
	goal := step.Target.At
	ticks, err := timeline.StepTicks(w, body)
	if err != nil {
		return false, err
	}
	var hops []comp.FarHop
	if hexmap.KeyOf(from) == hexmap.KeyOf(goal) {
		hops = []comp.FarHop{{At: goal, Arrive: w.Now + ticks*int64(max(1, hexmap.CellDistance(from, goal)))}}
	} else {
		route, found, err := graph.RouteFrom(from, hexmap.KeyOf(goal), farRouteNodes)
		if err != nil || !found {
			return false, err
		}
		legs := int64(len(route.Chunks) - 1)
		for i := range route.Chunks[1:] {
			at := route.Entries[i]
			if int64(i+1) == legs {
				at = goal
			}
			hops = append(hops, comp.FarHop{At: at, Arrive: w.Now + ticks*int64(route.Cost)*int64(i+1)/legs})
		}
	}
	ecs.NewMap[comp.Far](w.ECS).Get(body).Hops = hops
	return true, nil
}

// wanderFar sends the body to a neighboring chunk the graph joins to its own, drawn from the body's own
// sequence.
func wanderFar(w *sim.World, body ecs.Entity) (bool, error) {
	graph := graphOf(w)
	if graph == nil {
		return false, nil
	}
	from := ecs.NewMap[comp.Pos](w.ECS).Get(body).At
	links, err := graph.Neighbors(hexmap.KeyOf(from))
	if err != nil || len(links) == 0 {
		return false, err
	}
	ticks, err := timeline.StepTicks(w, body)
	if err != nil {
		return false, err
	}
	link := links[w.RNG.Derived(wanderStream, planSalt(w.UIDOf(body), w.Now)).IntN(len(links))]
	hop := comp.FarHop{At: worldgen.Center(link.To), Arrive: w.Now + ticks*int64(link.Cost)}
	ecs.NewMap[comp.Far](w.ECS).Get(body).Hops = []comp.FarHop{hop}
	return true, nil
}

// act does a step's action at once, as a far body does: its start and landing effects apply, when they
// could apply now, as the operator's [FarPicker] aims them, or else at the step's target or the body
// itself. It reports false when they could not, or there is nothing to aim at.
func act(w *sim.World, body ecs.Entity, step Step, op content.Operator) (bool, error) {
	action, ok := w.Content.Action(op.Does.Action)
	if !ok {
		return false, fmt.Errorf("agents: operator %q names unknown action %q", step.Operator, op.Does.Action)
	}
	intent, ok, err := aimFar(w, body, step, op)
	if err != nil || !ok {
		return false, err
	}
	refs := slices.Concat(action.OnStart, action.OnLand)
	ctx := &effects.Context{
		World: w, Source: body, Target: intent.Target, Dir: intent.Dir, Choices: intent.Choices,
		Action: op.Does.Action,
	}
	if ok, _ := effects.Check(ctx, refs); !ok {
		return false, nil
	}
	if err := effects.Apply(ctx, refs); err != nil {
		return false, fmt.Errorf("agents: body %d acting far: %w", w.UIDOf(body), err)
	}
	return true, nil
}

// aimFar picks what a far step's action lands on: what the operator's [FarPicker] picks, else the step's
// target, else the body. It reports false when the named target is gone or the picker finds nothing.
func aimFar(w *sim.World, body ecs.Entity, step Step, op content.Operator) (actions.Intent, bool, error) {
	if picker, ok := mindsOf(w).pickers[op.Does.Target].(FarPicker); ok {
		return picker.PickFar(w, body, step)
	}
	if step.Target == nil || step.Target.Subject == 0 {
		return actions.Intent{Target: body}, true, nil
	}
	found, ok := w.ByUID(step.Target.Subject)
	if !ok || !w.Alive(found) {
		return actions.Intent{}, false, nil
	}
	return actions.Intent{Target: found}, true, nil
}
