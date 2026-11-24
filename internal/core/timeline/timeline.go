package timeline

import (
	"cmp"
	"container/heap"
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// MaxAITurns is the most non-player decisions one Run takes, so a world with no player cannot hang.
const MaxAITurns = 10_000

// bodyTag is the tag of a body's own channel.
const bodyTag = "body"

// Channel is something that can be busy: a body itself, or one of its parts.
type Channel struct {
	Entity ecs.Entity // the body, or one of its parts
	// Body is the body the channel belongs to. The zero entity means Entity is the body.
	Body ecs.Entity
	Tags []string // "body" for the body itself; a part's are its slot and what it grants
}

// body returns the body that owns c.
func (c Channel) body() ecs.Entity {
	if c.Body.IsZero() {
		return c.Entity
	}
	return c.Body
}

// IsPart reports whether the channel is one of a body's parts rather than the body itself.
func (c Channel) IsPart() bool { return !c.Body.IsZero() && c.Entity != c.Body }

// FirstPart is the entity of the first channel that is a part, or the zero entity when none is.
func FirstPart(channels []Channel) ecs.Entity {
	for _, c := range channels {
		if c.IsPart() {
			return c.Entity
		}
	}
	return ecs.Entity{}
}

// UIDs lists the UIDs of the channels' entities, in the channels' order.
func UIDs(w *sim.World, channels []Channel) []api.UID {
	uids := make([]api.UID, len(channels))
	for i, c := range channels {
		uids[i] = w.UIDOf(c.Entity)
	}
	return uids
}

// slot is one timeline entry. Entries go stale when the body's decision tick changes or it is
// despawned; pop discards them by checking st.queued.
type slot struct {
	tick int64
	uid  api.UID
	e    ecs.Entity
}

// queue is a min-heap of slots ordered by (tick, uid).
type queue []slot

func (q queue) Len() int { return len(q) }
func (q queue) Less(i, j int) bool {
	if q[i].tick != q[j].tick {
		return q[i].tick < q[j].tick
	}
	return q[i].uid < q[j].uid
}
func (q queue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *queue) Push(x any)   { *q = append(*q, x.(slot)) }
func (q *queue) Pop() any {
	old := *q
	last := old[len(old)-1]
	*q = old[:len(old)-1]
	return last
}

// state is everything timeline keeps in the world. It is an Ark resource, so sim never knows about it.
type state struct {
	cfg         content.TimelineConfig
	channels    func(w *sim.World, body ecs.Entity) []Channel
	usable      func(w *sim.World, body ecs.Entity, c Channel) bool
	usableAll   func(w *sim.World, body ecs.Entity, channels []Channel) []bool
	speed       func(w *sim.World, body ecs.Entity) int
	visible     func(w *sim.World, body ecs.Entity) bool
	halt        func(w *sim.World) bool
	controllers map[string]Controller
	line        queue
	queued      map[ecs.Entity]int64 // the tick of each body's one live slot; only ever looked up
	current     ecs.Entity
	hasCurrent  bool
	systemsDue  bool                // the last entry has not had its per-entry systems run yet
	sleeping    map[ecs.Entity]bool // bodies whose last decision only waited; others do not wake for them
	changes     int64               // counts Occupy, Extend and Wait, to tell whether a controller did anything
	actors      *ecs.Map[comp.Actor]
	busy        *ecs.Map[comp.Busy]
	members     *ecs.Map[comp.MemberOf]
	pendings    *ecs.Map[comp.Pending]
	resolvers   map[string]Resolver
	seq         uint64 // the last Seq handed out
}

// stateOf returns the world's timeline state, creating it on first use.
func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{
		controllers: map[string]Controller{IdleController: idle{}},
		queued:      map[ecs.Entity]int64{},
		sleeping:    map[ecs.Entity]bool{},
		actors:      ecs.NewMap[comp.Actor](w.ECS),
		busy:        ecs.NewMap[comp.Busy](w.ECS),
		members:     ecs.NewMap[comp.MemberOf](w.ECS),
		pendings:    ecs.NewMap[comp.Pending](w.ECS),
		resolvers:   map[string]Resolver{},
	}
	st.channels = func(_ *sim.World, body ecs.Entity) []Channel {
		return []Channel{{Entity: body, Tags: []string{bodyTag}}}
	}
	st.usable = func(*sim.World, ecs.Entity, Channel) bool { return true }
	st.speed = func(*sim.World, ecs.Entity) int { return st.cfg.DefaultSpeed }
	st.visible = func(_ *sim.World, body ecs.Entity) bool { return st.members.Has(body) }
	res.Add(st)
	return st
}

// Channels lists a body's channels: the body first, then parts by UID.
func Channels(w *sim.World, body ecs.Entity) []Channel {
	st := stateOf(w)
	channels := slices.Clone(st.channels(w, body))
	for i := range channels {
		if channels[i].Body.IsZero() && channels[i].Entity != body {
			channels[i].Body = body
		}
	}
	slices.SortStableFunc(channels, func(a, b Channel) int {
		return compareChannels(w, body, a, b)
	})
	return channels
}

func compareChannels(w *sim.World, body ecs.Entity, a, b Channel) int {
	rank := func(c Channel) int {
		if c.Entity == body {
			return 0
		}
		return 1
	}
	return cmp.Or(
		cmp.Compare(rank(a), rank(b)),
		w.CompareUID(a.Entity, b.Entity),
		cmp.Compare(a.Entity.ID(), b.Entity.ID()),
	)
}

// busyUntil is the tick c is free again: now when it is free already.
func (st *state) busyUntil(w *sim.World, c Channel) int64 {
	if w.ECS.Alive(c.Entity) && st.busy.Has(c.Entity) {
		return max(st.busy.Get(c.Entity).Until, w.Now)
	}
	return w.Now
}

// Free reports whether the channel is not busy at the current tick.
func Free(w *sim.World, c Channel) bool { return stateOf(w).busyUntil(w, c) <= w.Now }

// StepTicks is the base ticks a step keeps a body's channels busy, from its speed.
func StepTicks(w *sim.World, body ecs.Entity) (int64, error) {
	st := stateOf(w)
	ticks, err := w.Content.StepTicks(content.SpeedEnv{Speed: st.speed(w, body)})
	if err != nil {
		return 0, err
	}
	if ticks < 1 {
		return 0, fmt.Errorf("timeline: step_ticks gave %d; a step must take at least 1 tick", ticks)
	}
	return ticks, nil
}

// decisionTick is when body next decides: the later of its WaitUntil and the earliest time a usable
// channel is free (now, if one is free already). If no channel is usable it uses all of them, so a
// body that can do nothing still gets to wait.
func (st *state) decisionTick(w *sim.World, body ecs.Entity) int64 {
	channels := Channels(w, body)
	usable := make([]Channel, 0, len(channels))
	var flags []bool
	if st.usableAll != nil {
		flags = st.usableAll(w, body, channels)
		if len(flags) != len(channels) {
			flags = nil // a hook that miscounts is ignored; fall back to the per-channel test
		}
	}
	for i, c := range channels {
		if (flags != nil && flags[i]) || (flags == nil && st.usable(w, body, c)) {
			usable = append(usable, c)
		}
	}
	if len(usable) == 0 {
		usable = channels
	}
	earliest := w.Now
	if len(usable) > 0 {
		earliest = st.busyUntil(w, usable[0])
		for _, c := range usable[1:] {
			earliest = min(earliest, st.busyUntil(w, c))
		}
	}
	return max(st.actors.Get(body).WaitUntil, earliest, w.Now)
}

// schedule gives body a live slot at its current decision tick, replacing any other.
func (st *state) schedule(w *sim.World, body ecs.Entity) {
	tick := st.decisionTick(w, body)
	if old, ok := st.queued[body]; ok && old == tick {
		return
	}
	st.queued[body] = tick
	heap.Push(&st.line, slot{tick: tick, uid: w.UIDOf(body), e: body})
}

// live reports whether s is still the body's live slot.
func (st *state) live(w *sim.World, s slot) bool {
	if !w.ECS.Alive(s.e) || !st.actors.Has(s.e) {
		return false
	}
	tick, ok := st.queued[s.e]
	return ok && tick == s.tick
}

// pop removes and returns the body with the earliest live slot (ties by UID). A slot whose decision
// tick has moved since it was queued is requeued rather than dropped.
func (st *state) pop(w *sim.World) (ecs.Entity, bool) {
	for st.line.Len() > 0 {
		next := heap.Pop(&st.line).(slot)
		if !st.live(w, next) {
			continue
		}
		delete(st.queued, next.e)
		if tick := st.decisionTick(w, next.e); tick != next.tick {
			st.schedule(w, next.e)
			continue
		}
		delete(st.sleeping, next.e) // it is awake: it decides now
		return next.e, true
	}
	return ecs.Entity{}, false
}

// peekTick is the tick of the earliest decision, discarding stale entries as pop does.
func (st *state) peekTick(w *sim.World) (int64, bool) {
	for st.line.Len() > 0 {
		next := st.line[0]
		if !st.live(w, next) {
			heap.Pop(&st.line)
			continue
		}
		if tick := st.decisionTick(w, next.e); tick != next.tick {
			heap.Pop(&st.line)
			delete(st.queued, next.e)
			st.schedule(w, next.e)
			continue
		}
		return next.tick, true
	}
	return 0, false
}

// Current returns the body that is deciding now: it has a free usable channel at this tick.
func Current(w *sim.World) (ecs.Entity, bool) {
	st := stateOf(w)
	if st.hasCurrent && w.ECS.Alive(st.current) && st.actors.Has(st.current) &&
		st.decisionTick(w, st.current) <= w.Now {
		return st.current, true
	}
	return ecs.Entity{}, false
}

// checkChannel returns the channel's body, or an error if the channel can't be used.
func (st *state) checkChannel(w *sim.World, c Channel, ticks int64) (ecs.Entity, error) {
	body := c.body()
	switch {
	case ticks < 1:
		return body, fmt.Errorf("timeline: a channel cannot be busy for %d ticks", ticks)
	case !w.ECS.Alive(c.Entity) || !w.ECS.Alive(body) || !st.actors.Has(body):
		return body, fmt.Errorf("timeline: entity %d is not a channel of a living actor", w.UIDOf(c.Entity))
	}
	return body, nil
}

// Occupy makes a channel busy for ticks from now and reschedules its body. The channel must be free.
func Occupy(w *sim.World, c Channel, ticks int64) error {
	st := stateOf(w)
	body, err := st.checkChannel(w, c, ticks)
	if err != nil {
		return err
	}
	if !Free(w, c) {
		return fmt.Errorf("timeline: channel of entity %d is already busy until tick %d",
			w.UIDOf(c.Entity), st.busy.Get(c.Entity).Until)
	}
	st.setBusy(w, c.Entity, w.Now+ticks, false)
	st.changes++
	st.schedule(w, body)
	return nil
}

// Extend adds ticks to a channel's busy time, counting from now if it is free, and reschedules its body.
// It leaves alone whether a lock holds the channel; use Lock for that.
func Extend(w *sim.World, c Channel, ticks int64) error { return extend(w, c, ticks, false) }

// Lock is Extend for the lock effect: the channel then counts as held (see Locked) until its busy
// time ends or it is released.
func Lock(w *sim.World, c Channel, ticks int64) error { return extend(w, c, ticks, true) }

// Locked reports whether the channel is busy because a lock holds it, not because of its own action.
func Locked(w *sim.World, c Channel) bool {
	st := stateOf(w)
	return w.ECS.Alive(c.Entity) && st.busy.Has(c.Entity) && !Free(w, c) && st.busy.Get(c.Entity).Locked
}

func extend(w *sim.World, c Channel, ticks int64, locked bool) error {
	st := stateOf(w)
	body, err := st.checkChannel(w, c, ticks)
	if err != nil {
		return err
	}
	st.setBusy(w, c.Entity, st.busyUntil(w, c)+ticks, locked)
	st.changes++
	st.schedule(w, body)
	return nil
}

// setBusy records the busy time. A lock flag left from busy time that has already ended is dropped;
// an active one stays unless this call sets it.
func (st *state) setBusy(w *sim.World, e ecs.Entity, until int64, locked bool) {
	if st.busy.Has(e) {
		busy := st.busy.Get(e)
		busy.Locked = locked || (busy.Locked && busy.Until > w.Now)
		busy.Until = until
		return
	}
	st.busy.Add(e, &comp.Busy{Until: until, Locked: locked})
}

// waitUntil sets the body's WaitUntil and reschedules it.
func (st *state) waitUntil(w *sim.World, body ecs.Entity, tick int64) {
	st.actors.Get(body).WaitUntil = tick
	st.changes++
	st.schedule(w, body)
}

// nextOtherEntry is the tick of the earliest live entry that is not body's, but never before the
// next tick, so waiting always moves time on. Entries of sleeping bodies (whose last decision only
// waited) are skipped, so bodies that do nothing don't wake each other a tick at a time. With no
// other entry it is the next tick.
func (st *state) nextOtherEntry(w *sim.World, body ecs.Entity) int64 {
	best := w.Now + 1
	found := false
	for _, s := range st.line {
		if s.e == body || st.sleeping[s.e] || !st.live(w, s) {
			continue
		}
		if !found || s.tick < best {
			found, best = true, s.tick
		}
	}
	return max(best, w.Now+1)
}

// Run plays the timeline forward until a player body decides, and returns it: the zero entity when
// SetHalt's check holds. It errors after MaxAITurns non-player decisions.
func Run(w *sim.World, clk *clock.Clock) (current ecs.Entity, err error) {
	st := stateOf(w)
	aiTurns := 0
	for {
		if st.halt != nil && st.halt(w) {
			return ecs.Entity{}, nil
		}
		if body, ok := Current(w); ok && st.actors.Get(body).Controller == PlayerController &&
			!st.landingDue(w) {
			return body, nil
		}
		if st.systemsDue {
			st.systemsDue = false
			if err := w.Systems.Run(w, sim.PerEntry); err != nil {
				return ecs.Entity{}, fmt.Errorf("timeline: per-entry systems: %w", err)
			}
		}
		if landed, err := st.landNext(w, clk); err != nil {
			return ecs.Entity{}, err
		} else if landed {
			continue
		}
		body, ok := st.pop(w)
		if !ok {
			return ecs.Entity{}, errors.New("timeline: nobody is on the timeline")
		}
		if err := st.open(w, clk, body); err != nil {
			return ecs.Entity{}, err
		}
		controller := st.actors.Get(body).Controller
		if controller == PlayerController {
			return body, nil
		}
		if aiTurns == MaxAITurns {
			return ecs.Entity{}, fmt.Errorf("timeline: %d decisions passed without the player acting", MaxAITurns)
		}
		aiTurns++
		if err := st.decide(w, body, controller); err != nil {
			return ecs.Entity{}, err
		}
	}
}

// open moves the clock to body's decision and announces it.
func (st *state) open(w *sim.World, clk *clock.Clock, body ecs.Entity) error {
	if tick := st.decisionTick(w, body); tick > w.Now {
		if err := clk.Advance(w, tick-w.Now); err != nil {
			return err
		}
	}
	st.current, st.hasCurrent = body, true
	st.systemsDue = true
	w.Bus.Publish(api.Ready{Source: w.UIDOf(body), Tick: w.Now})
	if _, err := w.Bus.Drain(w); err != nil {
		return fmt.Errorf("timeline: ready: %w", err)
	}
	return nil
}

// decide runs a non-player body's controller, and makes the body wait if it did nothing.
func (st *state) decide(w *sim.World, body ecs.Entity, name string) error {
	controller, ok := st.controllers[name]
	if !ok {
		return fmt.Errorf("timeline: entity %d has unknown controller %q", w.UIDOf(body), name)
	}
	before := st.changes
	if err := controller.Decide(w, body); err != nil {
		return fmt.Errorf("timeline: controller %q: %w", name, err)
	}
	if st.changes != before {
		delete(st.sleeping, body)
	} else if w.ECS.Alive(body) && st.actors.Has(body) {
		st.sleeping[body] = true
		st.waitUntil(w, body, st.nextOtherEntry(w, body))
	}
	return nil
}

// Release frees a busy channel at once and reschedules its body. A free channel is left alone.
func Release(w *sim.World, c Channel) error {
	st := stateOf(w)
	body, err := st.checkChannel(w, c, 1)
	if err != nil {
		return err
	}
	if st.busy.Has(c.Entity) {
		st.busy.Remove(c.Entity)
	}
	st.changes++
	st.schedule(w, body)
	return nil
}

// Reschedule recomputes a body's decision tick and requeues it. Call it after a hook's answer
// changes, such as when parts are lost or attached. Deciding also takes a body off the timeline
// until it acts or waits, so a body whose controller changed while the player was deciding for it
// needs this to be decided for again. A body with no Actor is left alone.
func Reschedule(w *sim.World, body ecs.Entity) {
	st := stateOf(w)
	if w.ECS.Alive(body) && st.actors.Has(body) {
		st.schedule(w, body)
	}
}

// Speed is the body's speed, as the world's speed hook gives it.
func Speed(w *sim.World, body ecs.Entity) int { return stateOf(w).speed(w, body) }
