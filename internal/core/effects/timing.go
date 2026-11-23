package effects

import (
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// interrupted is the reason given when an effect cancels an action.
const interrupted api.ContentID = "text.effects.interrupted"

func registerTiming(w *sim.World) error {
	channelsKind := map[string]Kind{"channels": KindChannels}
	for _, p := range []Primitive{
		{Name: "lock", Params: []string{"channels", "ticks"}, Kinds: channelsKind, Apply: applyLock},
		{
			Name: "interrupt", Optional: []string{"bodies", "channels", "when"},
			Kinds:  map[string]Kind{"bodies": KindBodies, "channels": KindChannels},
			Verify: verifyInterrupt, Apply: applyInterrupt,
		},
		{Name: "delay", Params: []string{"channels", "ticks"}, Kinds: channelsKind, Apply: applyDelay},
		{Name: "ready", Params: []string{"channels", "count"}, Kinds: channelsKind, Apply: applyReady},
	} {
		if err := Register(w, p); err != nil {
			return err
		}
	}
	return nil
}

func uidsOf(w *sim.World, channels []timeline.Channel) []api.UID {
	uids := make([]api.UID, len(channels))
	for i, c := range channels {
		uids[i] = w.UIDOf(c.Entity)
	}
	return uids
}

// windingUp lists the pending actions that use the channel: those that name it as their part, and
// actions from content, which may occupy several channels at once.
func windingUp(w *sim.World, c timeline.Channel) []ecs.Entity {
	pendings := ecs.NewMap[comp.Pending](w.ECS)
	landings := ecs.NewMap[comp.Landing](w.ECS)
	uid := w.UIDOf(c.Entity)
	body := c.Body
	if body.IsZero() {
		body = c.Entity
	}
	var out []ecs.Entity
	for _, e := range timeline.PendingOf(w, w.UIDOf(body)) {
		if pendings.Get(e).Part == uid || (landings.Has(e) && slices.Contains(landings.Get(e).Channels, uid)) {
			out = append(out, e)
		}
	}
	return out
}

// WindingUp lists the pending actions that use the channel, for primitives elsewhere that take a
// channel away, such as a severed arm.
func WindingUp(w *sim.World, c timeline.Channel) []ecs.Entity { return windingUp(w, c) }

func ticksParam(p Params) (int64, error) {
	ticks, err := p.Int("ticks")
	if err != nil {
		return 0, err
	}
	if ticks < 1 {
		return 0, fmt.Errorf("ticks must be at least 1, got %d", ticks)
	}
	return int64(ticks), nil
}

func applyLock(ctx *Context, p Params) error {
	channels, err := p.Channels("channels")
	if err != nil {
		return err
	}
	ticks, err := ticksParam(p)
	if err != nil {
		return err
	}
	w := ctx.World
	for _, c := range channels {
		for _, pending := range windingUp(w, c) {
			timeline.Cancel(w, pending, interrupted)
		}
		if err := timeline.Lock(w, c, ticks); err != nil {
			return err
		}
	}
	if len(channels) > 0 {
		w.Bus.Publish(api.ChannelsLocked{Channels: uidsOf(w, channels), Ticks: ticks, Tick: w.Now})
	}
	return nil
}

func verifyInterrupt(raw map[string]string) error {
	_, bodies := raw["bodies"]
	_, channels := raw["channels"]
	if bodies == channels {
		return errors.New("give exactly one of bodies or channels")
	}
	return nil
}

func applyInterrupt(ctx *Context, p Params) error {
	if p.Has("when") {
		holds, err := p.Bool("when")
		if err != nil {
			return err
		}
		if !holds {
			return nil
		}
	}
	w := ctx.World
	var pendings []ecs.Entity
	switch {
	case p.Has("bodies") && !p.Has("channels"):
		bodies, err := p.Bodies("bodies")
		if err != nil {
			return err
		}
		for _, body := range bodies {
			pendings = append(pendings, timeline.PendingOf(w, w.UIDOf(body))...)
		}
	case p.Has("channels") && !p.Has("bodies"):
		channels, err := p.Channels("channels")
		if err != nil {
			return err
		}
		for _, c := range channels {
			pendings = append(pendings, windingUp(w, c)...)
		}
	default:
		return errors.New("give exactly one of bodies or channels")
	}
	pendingOf := ecs.NewMap[comp.Pending](w.ECS)
	actors := map[api.UID]bool{}
	for _, e := range pendings {
		actors[pendingOf.Get(e).Actor] = true
	}
	for _, e := range pendings {
		timeline.Cancel(w, e, interrupted)
	}
	if len(actors) > 0 {
		list := make([]api.UID, 0, len(actors))
		for uid := range actors {
			list = append(list, uid)
		}
		slices.Sort(list)
		w.Bus.Publish(api.ActionsInterrupted{Actors: list, Tick: w.Now})
	}
	return nil
}

func applyDelay(ctx *Context, p Params) error {
	channels, err := p.Channels("channels")
	if err != nil {
		return err
	}
	ticks, err := ticksParam(p)
	if err != nil {
		return err
	}
	w := ctx.World
	for _, c := range channels {
		if err := timeline.Extend(w, c, ticks); err != nil {
			return err
		}
	}
	if len(channels) > 0 {
		w.Bus.Publish(api.ChannelsDelayed{Channels: uidsOf(w, channels), Ticks: ticks, Tick: w.Now})
	}
	return nil
}

func applyReady(ctx *Context, p Params) error {
	channels, err := p.Channels("channels")
	if err != nil {
		return err
	}
	count, err := p.Int("count")
	if err != nil {
		return err
	}
	if count < 1 {
		return fmt.Errorf("count must be at least 1, got %d", count)
	}
	w := ctx.World
	var freed []timeline.Channel
	for _, c := range channels {
		if len(freed) == count {
			break
		}
		if timeline.Free(w, c) || len(windingUp(w, c)) > 0 {
			continue
		}
		if err := timeline.Release(w, c); err != nil {
			return err
		}
		freed = append(freed, c)
	}
	if len(freed) > 0 {
		w.Bus.Publish(api.ChannelsReadied{Channels: uidsOf(w, freed), Tick: w.Now})
	}
	return nil
}
