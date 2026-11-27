package combat

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const bodiesView = "combat.bodies"

// unknownMaterial is shown for a part whose material the party has not learned.
const unknownMaterial = "?"

// seenCells is the set of cells a body at a position sees, by the alert's sight.
func seenCells(w *sim.World, from comp.Pos) (map[api.Cell]bool, error) {
	cfg := w.Content.Combat()
	m := w.Map
	if m == nil {
		return nil, nil
	}
	cells, err := m.Visible(from.At, cfg.SightRadius, cfg.SightDepth)
	if err != nil {
		return nil, fmt.Errorf("combat: sight: %w", err)
	}
	seen := make(map[api.Cell]bool, len(cells))
	for _, cell := range cells {
		seen[cell] = true
	}
	return seen, nil
}

// visibleBodies lists the standing bodies any living member of the party sees, and the whole party,
// by UID.
func visibleBodies(w *sim.World, party ecs.Entity) ([]ecs.Entity, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	chosen := map[ecs.Entity]bool{}
	for _, member := range anatomy.Members(w, party) {
		chosen[member] = true
	}
	candidates := hostileCandidates(w)
	for _, member := range livingMembers(w, party) {
		if !positions.Has(member) {
			continue
		}
		from := *positions.Get(member)
		seen, err := seenCells(w, from)
		if err != nil {
			return nil, err
		}
		for _, other := range candidates {
			if seen[positions.Get(other).At] {
				chosen[other] = true
			}
		}
	}
	return slices.SortedFunc(maps.Keys(chosen), w.CompareUID), nil
}

func buildBodies(w *sim.World, v *api.View) error {
	party, ok := anatomy.ControlledParty(w)
	if !ok {
		return nil
	}
	bodies, err := visibleBodies(w, party)
	if err != nil {
		return err
	}
	view := &api.BodiesView{Bodies: []api.BodyView{}, Threats: []api.ThreatView{}}
	if current, ok := timeline.Current(w); ok {
		view.Current = w.UIDOf(current)
	}
	members := anatomy.Members(w, party)
	for _, body := range bodies {
		view.Bodies = append(view.Bodies, bodyView(w, party, members, body))
	}
	for _, member := range members {
		view.Threats = append(view.Threats, threatsAt(w, member)...)
	}
	slices.SortStableFunc(view.Threats, func(a, b api.ThreatView) int {
		return cmp.Or(cmp.Compare(a.Tick, b.Tick), cmp.Compare(a.Attacker, b.Attacker))
	})
	v.Bodies = view
	return nil
}

func bodyView(w *sim.World, party ecs.Entity, members []ecs.Entity, body ecs.Entity) api.BodyView {
	view := api.BodyView{UID: w.UIDOf(body), Name: nameOf(w, body), Channels: []api.ChannelView{}}
	view.Member = slices.Contains(members, body)
	for _, member := range livingMembers(w, party) {
		if anatomy.Hostile(w, member, body) {
			view.Hostile = true
			break
		}
	}
	if allegiances := ecs.NewMap[comp.Allegiance](w.ECS); allegiances.Has(body) {
		if faction, ok := w.Content.Faction(allegiances.Get(body).Faction); ok {
			view.Faction = w.Content.Render(api.Text{ID: faction.Name})
		}
	}
	if statuses := ecs.NewMap[comp.Statuses](w.ECS); statuses.Has(body) {
		for _, entry := range statuses.Get(body).Active {
			if def, ok := w.Content.Status(entry.Status); ok {
				view.Statuses = append(view.Statuses, w.Content.Render(api.Text{ID: def.Name}))
			}
		}
	}
	windups := windupsOf(w, body)
	for _, part := range anatomy.Parts(w, body) {
		view.Channels = append(view.Channels, channelView(w, party, view.Member, part, windups))
	}
	return view
}

// windupsOf maps each channel UID of a body to the action it is winding up.
func windupsOf(w *sim.World, body ecs.Entity) map[api.UID]*api.WindupView {
	out := map[api.UID]*api.WindupView{}
	landings := ecs.NewMap[comp.Landing](w.ECS)
	pendings := ecs.NewMap[comp.Pending](w.ECS)
	for _, pending := range timeline.PendingOf(w, w.UIDOf(body)) {
		if !landings.Has(pending) {
			continue
		}
		pend, landing := pendings.Get(pending), landings.Get(pending)
		wind := &api.WindupView{
			Action: actionName(w, landing.Action), Target: pend.Target, Lands: pend.Lands - w.Now,
			When: w.Content.Render(api.Text{
				ID: "text.combat.lands_at", Args: map[string]string{"tick": strconv.FormatInt(pend.Lands, 10)},
			}),
		}
		for _, channel := range landing.Channels {
			if _, taken := out[channel]; !taken {
				out[channel] = wind
			}
		}
	}
	return out
}

func actionName(w *sim.World, id api.ContentID) string {
	def, ok := w.Content.Action(id)
	if !ok {
		return string(id)
	}
	return w.Content.Render(api.Text{ID: def.Name})
}

func channelView(
	w *sim.World, party ecs.Entity, member bool, part ecs.Entity, windups map[api.UID]*api.WindupView,
) api.ChannelView {
	data := ecs.NewMap[comp.Part](w.ECS).Get(part)
	uid := w.UIDOf(part)
	view := api.ChannelView{
		UID: uid, Slot: data.Slot, Integrity: data.Integrity, MaxIntegrity: data.MaxIntegrity,
		Graft: ecs.NewMap[comp.Graft](w.ECS).Has(part), Material: unknownMaterial, Windup: windups[uid],
	}
	if template, ok := w.Content.Part(data.Template); ok {
		view.Name = w.Content.Render(api.Text{ID: template.Name})
	}
	if member {
		view.Material = materialName(w, data.Material)
	} else if fact, known := knowledge.Best(w, party, uid, w.Content.Combat().MaterialClaim); known {
		view.Material = w.Content.RenderValue(fact.Claim, fact.Value)
	}
	if busy := ecs.NewMap[comp.Busy](w.ECS); busy.Has(part) {
		if b := busy.Get(part); b.Until > w.Now {
			view.FreeIn, view.Locked = b.Until-w.Now, b.Locked
		}
	}
	view.State = channelState(w, view)
	return view
}

func channelState(w *sim.World, view api.ChannelView) string {
	state := api.Text{ID: "text.combat.channel.ready"}
	switch {
	case view.Locked:
		state = api.Text{ID: "text.combat.channel.locked"}
	case view.FreeIn > 0:
		state = api.Text{
			ID: "text.combat.channel.free_in", Args: map[string]string{"ticks": strconv.FormatInt(view.FreeIn, 10)},
		}
	}
	rendered := w.Content.Render(state)
	if view.Graft {
		return w.Content.Render(api.Text{ID: "text.combat.channel.graft", Args: map[string]string{"state": rendered}})
	}
	return rendered
}

func materialName(w *sim.World, id api.ContentID) string {
	if material, ok := w.Content.Material(id); ok {
		return w.Content.Render(api.Text{ID: material.Name})
	}
	return string(id)
}

// threatsAt lists the actions in flight that are aimed at the body.
func threatsAt(w *sim.World, target ecs.Entity) []api.ThreatView {
	var out []api.ThreatView
	landings := ecs.NewMap[comp.Landing](w.ECS)
	pendings := ecs.NewMap[comp.Pending](w.ECS)
	for _, pending := range timeline.AimedAt(w, w.UIDOf(target)) {
		if !landings.Has(pending) {
			continue
		}
		pend, landing := pendings.Get(pending), landings.Get(pending)
		out = append(out, api.ThreatView{
			Attacker: pend.Actor, Target: pend.Target, TargetPart: landing.TargetPart,
			Action: actionName(w, landing.Action), Lands: pend.Lands - w.Now, Tick: pend.Lands,
		})
	}
	return out
}

// nameOf is the text of a body's name, "text.template.<template without its prefix>".
func nameOf(w *sim.World, e ecs.Entity) string {
	return w.Content.Render(w.Content.TemplateName(w.TemplateOf(e)))
}
