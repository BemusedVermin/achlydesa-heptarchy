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

// The body selectors. Allies, enemies, others and everyone all mean "in sight of the source".
const (
	BodiesSelf     BodySet = "self"
	BodiesTarget   BodySet = "target"
	BodiesAllies   BodySet = "allies"   // in sight, not hostile to the source, not the source
	BodiesEnemies  BodySet = "enemies"  // in sight and hostile to the source
	BodiesOthers   BodySet = "others"   // in sight, not the source
	BodiesEveryone BodySet = "everyone" // in sight, and the source
)

var bodySets = []BodySet{BodiesSelf, BodiesTarget, BodiesAllies, BodiesEnemies, BodiesOthers, BodiesEveryone}

// BodySelector is a parsed body selector.
type BodySelector struct{ Set BodySet }

// ChannelSelector is a parsed channel selector: the action's own channels, the target part, or the
// channels of some bodies that carry a tag.
type ChannelSelector struct {
	Used       bool
	TargetPart bool
	Bodies     BodySelector
	Tag        string // "all" matches every channel
}

// ParseBodies parses a body selector: self, target, allies, enemies, others or everyone.
func ParseBodies(src string) (BodySelector, error) {
	set := BodySet(src)
	if !slices.Contains(bodySets, set) {
		return BodySelector{}, fmt.Errorf("bad body selector %q: want one of %v", src, bodySets)
	}
	return BodySelector{Set: set}, nil
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
	default:
		inSight, err := bodiesInSight(ctx)
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
			}
		}
	}
	sortByUID(w, out)
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
			if alive(w, c.Entity) {
				out = append(out, c)
			}
		}
	case s.TargetPart:
		if alive(w, ctx.TargetPart) && alive(w, ctx.Target) {
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
		return compareUID(w.UIDOf(a.Entity), w.UIDOf(b.Entity))
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
	if alive(w, e) {
		return []ecs.Entity{e}
	}
	return nil
}

func sortByUID(w *sim.World, entities []ecs.Entity) {
	slices.SortStableFunc(entities, func(a, b ecs.Entity) int { return compareUID(w.UIDOf(a), w.UIDOf(b)) })
}

func compareUID(a, b api.UID) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// bodiesInSight lists the bodies (entities on the timeline with a position) standing in a cell the
// source can see, the source itself included. Sight is hexmap.Visible out to the party's sight
// radius, and as many depths up and down, until content says otherwise.
func bodiesInSight(ctx *Context) ([]ecs.Entity, error) {
	w := ctx.World
	st := stateOf(w)
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !alive(w, ctx.Source) || !positions.Has(ctx.Source) || st.maps == nil {
		return nil, nil
	}
	from := positions.Get(ctx.Source)
	m := st.maps(from.Map)
	if m == nil {
		return nil, nil
	}
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
		if pos.Map == from.Map && seen[pos.At] {
			out = append(out, query.Entity())
		}
	}
	return out, nil
}
