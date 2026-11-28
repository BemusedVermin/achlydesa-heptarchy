package agents

import (
	"cmp"
	"container/heap"
	"slices"
	"sync"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// GOAPPlanner is the name templates give as mind.planner to plan by search.
const GOAPPlanner = "goap"

// goap plans by uniform-cost search forward over believed states, from the agent's beliefs to a state
// where the goal's want holds.
type goap struct {
	mu    sync.Mutex
	cache map[goapKey]goapResult
}

type goapKey struct {
	state string
	goal  api.ContentID
}

type goapResult struct {
	steps []api.ContentID
	found bool
}

func newGOAP() *goap { return &goap{cache: map[goapKey]goapResult{}} }

// Plan finds the cheapest plan within the node and depth caps. Equal costs break by depth, then by
// the order states were reached, and operators are tried in ID order, so the same state and goal
// always give the same plan. The cache is keyed by the whole state, so a belief that changes misses it.
func (g *goap) Plan(snapshot Snapshot, goal api.ContentID) (Plan, bool) {
	key := goapKey{state: snapshot.Beliefs.Key(), goal: goal}
	g.mu.Lock()
	result, hit := g.cache[key]
	g.mu.Unlock()
	if !hit {
		result = searchGOAP(snapshot, goal)
		cfg := snapshot.Content.Agents()
		g.mu.Lock()
		if len(g.cache) >= cfg.GoapCache {
			clear(g.cache)
		}
		g.cache[key] = result
		g.mu.Unlock()
	}
	if !result.found {
		return Plan{}, false
	}
	steps := make([]Step, len(result.steps))
	for i, id := range result.steps {
		steps[i] = Step{Operator: id}
	}
	return Plan{Steps: steps}, true
}

type goapNode struct {
	state  State
	parent *goapNode
	op     api.ContentID
	cost   float64
	depth  int
	seq    int
}

type goapQueue []*goapNode

func (q goapQueue) Len() int { return len(q) }
func (q goapQueue) Less(i, j int) bool {
	return cmp.Or(
		cmp.Compare(q[i].cost, q[j].cost),
		cmp.Compare(q[i].depth, q[j].depth),
		cmp.Compare(q[i].seq, q[j].seq),
	) < 0
}
func (q goapQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *goapQueue) Push(x any)   { *q = append(*q, x.(*goapNode)) }
func (q *goapQueue) Pop() any {
	old := *q
	last := old[len(old)-1]
	*q = old[:len(old)-1]
	return last
}

func searchGOAP(snapshot Snapshot, goalID api.ContentID) goapResult {
	reg := snapshot.Content
	goal, ok := reg.Goal(goalID)
	if !ok || len(goal.Want) == 0 || holdsAll(goal.Want, snapshot.Beliefs) {
		return goapResult{}
	}
	cfg := reg.Agents()
	type operator struct {
		id api.ContentID
		op content.Operator
	}
	var ops []operator
	for _, id := range reg.OperatorIDs() {
		op, _ := reg.Operator(id)
		ops = append(ops, operator{id, op})
	}
	seq := 0
	open := &goapQueue{{state: snapshot.Beliefs}}
	best := map[string]float64{snapshot.Beliefs.Key(): 0}
	for expanded := 0; open.Len() > 0 && expanded < cfg.GoapNodes; expanded++ {
		node := heap.Pop(open).(*goapNode)
		if holdsAll(goal.Want, node.state) {
			var steps []api.ContentID
			for n := node; n.parent != nil; n = n.parent {
				steps = append(steps, n.op)
			}
			slices.Reverse(steps)
			return goapResult{steps: steps, found: true}
		}
		if node.depth >= cfg.GoapDepth || node.cost > best[node.state.Key()] {
			continue
		}
		for _, o := range ops {
			if !Holds(o.op, node.state) {
				continue
			}
			next := Apply(o.op, node.state)
			cost := node.cost + o.op.Cost
			nextKey := next.Key()
			if known, seen := best[nextKey]; seen && known <= cost {
				continue
			}
			best[nextKey] = cost
			seq++
			heap.Push(open, &goapNode{state: next, parent: node, op: o.id, cost: cost, depth: node.depth + 1, seq: seq})
		}
	}
	return goapResult{}
}
