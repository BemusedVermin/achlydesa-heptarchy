package effects

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// allTag is the channel tag that matches every channel of a body.
const allTag = "all"

// BodySet is which bodies a selector means.
type BodySet string

// The body selectors. Allies, enemies, others and everyone all mean "in sight of the source";
// witnesses means in sight of the target, for an event such as a death that has no source.
const (
	BodiesSelf   BodySet = "self"
	BodiesTarget BodySet = "target"
	// In sight, not hostile to the source, not the source.
	BodiesAllies BodySet = "allies"
	// In sight and hostile to the source.
	BodiesEnemies BodySet = "enemies"
	// In sight, not the source.
	BodiesOthers BodySet = "others"
	// In sight, and the source.
	BodiesEveryone BodySet = "everyone"
	// In sight of the target, not the target.
	BodiesWitnesses BodySet = "witnesses"
)

var bodySets = []BodySet{
	BodiesSelf, BodiesTarget, BodiesAllies, BodiesEnemies, BodiesOthers, BodiesEveryone, BodiesWitnesses,
}

// BodySelector is a parsed body selector: a set of bodies, optionally narrowed to those whose
// template carries a tag, written <set>[<tag>] such as others[crawls].
type BodySelector struct {
	Set BodySet
	// When set, only bodies whose template has this tag.
	Tag string
}

// ChannelSelector is a parsed channel selector: the action's own channels, the target part, or the
// channels of some bodies that carry a tag.
type ChannelSelector struct {
	Used       bool
	TargetPart bool
	Bodies     BodySelector
	// "all" matches every channel.
	Tag string
}

// ParseBodies parses a body selector: self, target, allies, enemies, others, everyone or witnesses, each
// optionally followed by a template tag in brackets, such as others[crawls].
func ParseBodies(src string) (BodySelector, error) {
	name, tag := src, ""
	if open := strings.IndexByte(src, '['); open >= 0 {
		if !strings.HasSuffix(src, "]") || !validTag(src[open+1:len(src)-1]) {
			return BodySelector{}, fmt.Errorf("bad body selector %q: want <set>[<tag>] with a lower-case tag", src)
		}
		name, tag = src[:open], src[open+1:len(src)-1]
	}
	set := BodySet(name)
	if !slices.Contains(bodySets, set) {
		return BodySelector{}, fmt.Errorf("bad body selector %q: want one of %v", src, bodySets)
	}
	return BodySelector{Set: set, Tag: tag}, nil
}

// ParseChannels parses a channel selector: used, target_part, or <bodies>:<tag>, such as target:all,
// target:arm or self:body.
func ParseChannels(src string) (ChannelSelector, error) {
	switch src {
	case "used":
		return ChannelSelector{Used: true}, nil
	case "target_part":
		return ChannelSelector{TargetPart: true}, nil
	}
	bodies, tag, ok := strings.Cut(src, ":")
	if !ok {
		return ChannelSelector{}, fmt.Errorf("bad channel selector %q: want used, target_part or <bodies>:<tag>", src)
	}
	set, err := ParseBodies(bodies)
	if err != nil {
		return ChannelSelector{}, fmt.Errorf("bad channel selector %q: %w", src, err)
	}
	if !validTag(tag) {
		return ChannelSelector{}, fmt.Errorf("bad channel selector %q: tag %q must be lower-case words", src, tag)
	}
	return ChannelSelector{Bodies: set, Tag: tag}, nil
}

func validTag(tag string) bool {
	if tag == "" || tag[0] < 'a' || tag[0] > 'z' {
		return false
	}
	for _, r := range tag {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

// Resolve returns the selected bodies, sorted by UID.
func (s BodySelector) Resolve(ctx *Context) ([]ecs.Entity, error) {
	w := ctx.World
	var out []ecs.Entity
	switch s.Set {
	case BodiesSelf:
		out = livingOnly(w, ctx.Source)
	case BodiesTarget:
		out = livingOnly(w, ctx.Target)
	case BodiesWitnesses:
		inSight, err := bodiesInSight(w, ctx.Target)
		if err != nil {
			return nil, err
		}
		out = slices.DeleteFunc(inSight, func(body ecs.Entity) bool { return body == ctx.Target })
	default:
		inSight, err := bodiesInSight(w, ctx.Source)
		if err != nil {
			return nil, err
		}
		hostile := stateOf(w).hostile
		for _, body := range inSight {
			isSource := body == ctx.Source
			switch s.Set {
			case BodiesEveryone:
				out = append(out, body)
			case BodiesOthers:
				if !isSource {
					out = append(out, body)
				}
			case BodiesAllies:
				if !isSource && !hostile(w, ctx.Source, body) {
					out = append(out, body)
				}
			case BodiesEnemies:
				if !isSource && hostile(w, ctx.Source, body) {
					out = append(out, body)
				}
			case BodiesSelf, BodiesTarget, BodiesWitnesses:
			}
		}
	}
	if s.Tag != "" {
		out = slices.DeleteFunc(out, func(body ecs.Entity) bool {
			return !slices.Contains(BodyEnvOf(w, body, ecs.Entity{}).Template.Tags, s.Tag)
		})
	}
	w.SortByUID(out)
	return out, nil
}

// Resolve returns the selected channels, sorted by the UID of the channel's entity. A tag nothing
// carries gives an empty list.
func (s ChannelSelector) Resolve(ctx *Context) ([]timeline.Channel, error) {
	w := ctx.World
	var out []timeline.Channel
	switch {
	case s.Used:
		for _, c := range ctx.Used {
			if w.Alive(c.Entity) {
				out = append(out, c)
			}
		}
	case s.TargetPart:
		if w.Alive(ctx.TargetPart) && w.Alive(ctx.Target) {
			out = append(out, partChannel(w, ctx.Target, ctx.TargetPart))
		}
	default:
		bodies, err := s.Bodies.Resolve(ctx)
		if err != nil {
			return nil, err
		}
		for _, body := range bodies {
			for _, c := range timeline.Channels(w, body) {
				if s.Tag == allTag || slices.Contains(c.Tags, s.Tag) {
					out = append(out, c)
				}
			}
		}
	}
	slices.SortStableFunc(out, func(a, b timeline.Channel) int {
		return w.CompareUID(a.Entity, b.Entity)
	})
	return out, nil
}

// partChannel finds the part among the body's channels, so it comes with its tags.
func partChannel(w *sim.World, body, part ecs.Entity) timeline.Channel {
	for _, c := range timeline.Channels(w, body) {
		if c.Entity == part {
			return c
		}
	}
	return timeline.Channel{Entity: part, Body: body}
}

func livingOnly(w *sim.World, e ecs.Entity) []ecs.Entity {
	if w.Alive(e) {
		return []ecs.Entity{e}
	}
	return nil
}

// bodiesInSight lists the bodies (entities on the timeline with a position) standing in a cell the
// viewer can see, the viewer itself included. Sight is hexmap.Visible out to the party's sight
// radius, and as many depths up and down, until content says otherwise.
func bodiesInSight(w *sim.World, viewer ecs.Entity) ([]ecs.Entity, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	m := w.Map
	if !w.Alive(viewer) || !positions.Has(viewer) || m == nil {
		return nil, nil
	}
	from := positions.Get(viewer)
	radius := w.Content.Party().SightRadius
	cells, err := m.Visible(from.At, radius, w.Content.Party().SightDepth)
	if err != nil {
		return nil, fmt.Errorf("effects: sight: %w", err)
	}
	seen := make(map[api.Cell]bool, len(cells))
	for _, cell := range cells {
		seen[cell] = true
	}
	var out []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for query.Next() {
		pos, _ := query.Get()
		if seen[pos.At] {
			out = append(out, query.Entity())
		}
	}
	return out, nil
}
