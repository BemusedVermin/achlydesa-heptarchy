package turns

import (
	"container/heap"
	"errors"
	"fmt"
	"sort"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// MaxAITurns is the most non-player turns one Run takes, so a world with no player cannot hang.
const MaxAITurns = 10_000

// slot is one timeline entry. Entries go stale when a turn opens or the entity is despawned;
// pop discards them by checking the component.
type slot struct {
	next int64
	uid  api.UID
	e    ecs.Entity
}

// timeline is a min-heap of slots ordered by (next, uid).
type timeline []slot

func (t timeline) Len() int { return len(t) }
func (t timeline) Less(i, j int) bool {
	if t[i].next != t[j].next {
		return t[i].next < t[j].next
	}
	return t[i].uid < t[j].uid
}
func (t timeline) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t *timeline) Push(x any)   { *t = append(*t, x.(slot)) }
func (t *timeline) Pop() any {
	old := *t
	last := old[len(old)-1]
	*t = old[:len(old)-1]
	return last
}

// state is everything turns keeps in the world. It is an Ark resource, so sim never knows about turns.
type state struct {
	cfg         content.TurnsConfig
	speed       func(w *sim.World, e ecs.Entity) int
	budget      func(w *sim.World, e ecs.Entity) (steps, actions int)
	controllers map[string]Controller
	line        timeline
	current     ecs.Entity
	hasCurrent  bool
	turns       *ecs.Map[comp.Turn]
}

// stateOf returns the world's turn state, creating it on first use.
func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{
		controllers: map[string]Controller{IdleController: idle{}},
		turns:       ecs.NewMap[comp.Turn](w.ECS),
	}
	st.speed = func(*sim.World, ecs.Entity) int { return st.cfg.DefaultSpeed }
	st.budget = func(*sim.World, ecs.Entity) (int, int) { return 1, 0 }
	res.Add(st)
	return st
}

// pop removes and returns the live slot with the smallest (next, uid), skipping stale ones.
func (st *state) pop(w *sim.World) (ecs.Entity, bool) {
	for st.line.Len() > 0 {
		next := heap.Pop(&st.line).(slot)
		if st.valid(w, next) {
			return next.e, true
		}
	}
	return ecs.Entity{}, false
}

// valid reports whether s still describes a waiting turn.
func (st *state) valid(w *sim.World, s slot) bool {
	if !w.ECS.Alive(s.e) || !st.turns.Has(s.e) {
		return false
	}
	turn := st.turns.Get(s.e)
	return !turn.Open && turn.Next == s.next
}

func (st *state) schedule(w *sim.World, e ecs.Entity) {
	heap.Push(&st.line, slot{next: st.turns.Get(e).Next, uid: w.UIDOf(e), e: e})
}

// turnTicks is the delay from one turn to the next for e, before extra costs.
func (st *state) turnTicks(w *sim.World, e ecs.Entity) (int64, error) {
	ticks, err := w.Content.TurnTicks(content.TurnEnv{Speed: st.speed(w, e)})
	if err != nil {
		return 0, err
	}
	if ticks < 1 {
		return 0, fmt.Errorf("turns: turn_ticks gave %d; a turn must take at least 1 tick", ticks)
	}
	return ticks, nil
}

// hasOpenTurn reports whether e is alive with an open turn.
func (st *state) hasOpenTurn(w *sim.World, e ecs.Entity) bool {
	return w.ECS.Alive(e) && st.turns.Has(e) && st.turns.Get(e).Open
}

// Current returns the entity whose turn is open.
func Current(w *sim.World) (ecs.Entity, bool) {
	st := stateOf(w)
	if st.hasCurrent && st.hasOpenTurn(w, st.current) {
		return st.current, true
	}
	return ecs.Entity{}, false
}

// Run continues the timeline. If the current turn is still open and player-controlled, it returns at
// once. Otherwise it repeatedly takes the entity with the smallest Next (ties by UID), advances the
// clock to that tick if it's in the future (which runs per-phase and per-day systems), opens its
// turn and publishes TurnStarted. A player actor stops the loop; any other runs its controller and
// ends its turn. It errors after MaxAITurns AI turns in one call, so a world with no player can't hang.
func Run(w *sim.World, clk *clock.Clock) (current ecs.Entity, err error) {
	st := stateOf(w)
	aiTurns := 0
	for {
		actor, open := Current(w)
		if !open {
			next, ok := st.pop(w)
			if !ok {
				return ecs.Entity{}, errors.New("turns: nobody is on the timeline")
			}
			if err := st.open(w, clk, next); err != nil {
				return ecs.Entity{}, err
			}
			actor = next
		}
		controller := st.turns.Get(actor).Controller
		if controller == PlayerController {
			return actor, nil
		}
		if aiTurns == MaxAITurns {
			return ecs.Entity{}, fmt.Errorf("turns: %d turns passed without the player acting", MaxAITurns)
		}
		aiTurns++
		if err := st.takeAITurn(w, actor, controller); err != nil {
			return ecs.Entity{}, err
		}
	}
}

// open moves the clock to e's turn and starts it.
func (st *state) open(w *sim.World, clk *clock.Clock, e ecs.Entity) error {
	if next := st.turns.Get(e).Next; next > w.Now {
		if err := clk.Advance(w, next-w.Now); err != nil {
			return err
		}
	}
	st.turns.Get(e).Open = true
	st.current, st.hasCurrent = e, true
	w.Bus.Publish(api.TurnStarted{Actor: w.UIDOf(e), Tick: w.Now})
	if _, err := w.Bus.Drain(w); err != nil {
		return fmt.Errorf("turns: turn started: %w", err)
	}
	return nil
}

func (st *state) takeAITurn(w *sim.World, actor ecs.Entity, name string) error {
	controller, ok := st.controllers[name]
	if !ok {
		return fmt.Errorf("turns: entity %d has unknown controller %q", w.UIDOf(actor), name)
	}
	if err := controller.TakeTurn(w, actor); err != nil {
		return fmt.Errorf("turns: controller %q: %w", name, err)
	}
	if st.hasOpenTurn(w, actor) {
		return EndTurn(w, actor)
	}
	return nil
}

// Spend records use of the current turn's budget; over budget is an error (a bug, since commands
// are validated first).
func Spend(w *sim.World, actor ecs.Entity, step bool, actions int) error {
	steps, arms := Remaining(w, actor)
	if actions < 0 || (step && steps < 1) || actions > arms {
		return fmt.Errorf("turns: entity %d cannot spend step=%t actions=%d with %d steps and %d actions left",
			w.UIDOf(actor), step, actions, steps, arms)
	}
	turn := stateOf(w).turns.Get(actor)
	turn.Stepped = turn.Stepped || step
	turn.ActionsUsed += actions
	return nil
}

// Remaining is what the actor may still do this turn. It is 0, 0 outside the actor's open turn.
func Remaining(w *sim.World, actor ecs.Entity) (steps, actions int) {
	st := stateOf(w)
	if !st.hasOpenTurn(w, actor) {
		return 0, 0
	}
	turn := st.turns.Get(actor)
	steps, actions = st.budget(w, actor)
	if turn.Stepped {
		steps = 0
	}
	return min(max(steps, 0), 1), max(actions-turn.ActionsUsed, 0)
}

// AddExtra delays the actor's next turn by ticks more (terrain, climbing, waiting). The delay
// applies once: EndTurn clears it.
func AddExtra(w *sim.World, actor ecs.Entity, ticks int64) {
	st := stateOf(w)
	if w.ECS.Alive(actor) && st.turns.Has(actor) {
		st.turns.Get(actor).Extra += ticks
	}
}

// EndTurn closes the turn: Next = Now + turn_ticks(speed) + Extra, the budget resets, TurnEnded is
// published and PerTurn systems run.
func EndTurn(w *sim.World, actor ecs.Entity) error {
	st := stateOf(w)
	if !st.hasOpenTurn(w, actor) {
		return fmt.Errorf("turns: entity %d has no open turn to end", w.UIDOf(actor))
	}
	ticks, err := st.turnTicks(w, actor)
	if err != nil {
		return err
	}
	turn := st.turns.Get(actor)
	turn.Next = w.Now + ticks + max(turn.Extra, 0)
	turn.Extra, turn.Stepped, turn.ActionsUsed, turn.Open = 0, false, 0, false
	st.current, st.hasCurrent = ecs.Entity{}, false
	st.schedule(w, actor)
	w.Bus.Publish(api.TurnEnded{Actor: w.UIDOf(actor)})
	if _, err := w.Bus.Drain(w); err != nil {
		return fmt.Errorf("turns: turn ended: %w", err)
	}
	if err := w.Systems.Run(w, sim.PerTurn); err != nil {
		return fmt.Errorf("turns: per-turn systems: %w", err)
	}
	return nil
}

// upcoming lists up to n waiting actors in timeline order.
func (st *state) upcoming(w *sim.World, n int) []api.UID {
	live := make(timeline, 0, len(st.line))
	for _, s := range st.line {
		if st.valid(w, s) {
			live = append(live, s)
		}
	}
	sort.Slice(live, live.Less)
	live = live[:min(len(live), n)]
	uids := make([]api.UID, len(live))
	for i, s := range live {
		uids[i] = s.uid
	}
	return uids
}
