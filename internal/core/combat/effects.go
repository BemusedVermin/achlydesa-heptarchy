package combat

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Text IDs of the reasons and cancelations this package gives.
const (
	severedReason        api.ContentID = "text.combat.severed"
	inventoryShortReason api.ContentID = "text.combat.inventory_short"
	nothingToAttach      api.ContentID = "text.combat.no_severed_part"
)

// bodyOf is the body a part is attached to, or the zero entity.
func bodyOf(w *sim.World, part ecs.Entity) ecs.Entity {
	partOf := ecs.NewMap[comp.PartOf](w.ECS)
	if !w.ECS.Alive(part) || !partOf.Has(part) {
		return ecs.Entity{}
	}
	return partOf.GetRelation(part)
}

// sever takes an attached part off its body: its pending actions are canceled, it keeps its
// integrity and stays where the body stands, and the body is rescheduled for its lost channel.
func sever(w *sim.World, source, part ecs.Entity, cause string) {
	body := bodyOf(w, part)
	if body.IsZero() {
		return
	}
	for _, pending := range effects.WindingUp(w, timeline.Channel{Entity: part, Body: body}) {
		timeline.Cancel(w, pending, severedReason)
	}
	ecs.NewMap[comp.PartOf](w.ECS).Remove(part)
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if positions.Has(body) {
		sim.Put(w, part, *positions.Get(body))
	}
	w.Bus.Publish(api.PartSevered{
		Body: w.UIDOf(body), Part: w.UIDOf(part), Cause: cause, Source: w.UIDOf(source), Tick: w.Now,
		Line: partLine(w, "text.combat.part_severed", source, body, part, nil),
	})
	timeline.Reschedule(w, body)
}

func applyDamage(ctx *effects.Context, p effects.Params) error {
	channels, err := p.Channels("parts")
	if err != nil {
		return err
	}
	w := ctx.World
	parts := ecs.NewMap[comp.Part](w.ECS)
	for _, c := range channels {
		body := bodyOf(w, c.Entity)
		if body.IsZero() || !parts.Has(c.Entity) || parts.Get(c.Entity).Integrity <= 0 {
			continue
		}
		amount, err := nonNegative(p.For(c), "amount")
		if err != nil {
			return err
		}
		if amount == 0 {
			continue
		}
		part := parts.Get(c.Entity)
		lost := min(amount, part.Integrity)
		part.Integrity -= lost
		w.Bus.Publish(api.PartDamaged{
			Body: w.UIDOf(body), Part: w.UIDOf(c.Entity), Amount: lost, Integrity: part.Integrity,
			Chosen: maps.Clone(ctx.Choices), Source: w.UIDOf(ctx.Source), Tick: w.Now,
			Line: damageLine(ctx, body, c.Entity, lost, part.Integrity),
		})
		if part.Integrity == 0 {
			sever(w, ctx.Source, c.Entity, api.SeveredWorn)
		}
	}
	return nil
}

// applyRaiseMaxIntegrity raises the ceiling of each part's integrity. Integrity itself stays as it
// is, so the part is sturdier but not healed.
func applyRaiseMaxIntegrity(ctx *effects.Context, p effects.Params) error {
	channels, err := p.Channels("parts")
	if err != nil {
		return err
	}
	parts := ecs.NewMap[comp.Part](ctx.World.ECS)
	for _, c := range channels {
		if !ctx.World.ECS.Alive(c.Entity) || !parts.Has(c.Entity) {
			continue
		}
		amount, err := nonNegative(p.For(c), "amount")
		if err != nil {
			return err
		}
		parts.Get(c.Entity).MaxIntegrity += amount
	}
	return nil
}

func applySever(ctx *effects.Context, p effects.Params) error {
	channels, err := p.Channels("parts")
	if err != nil {
		return err
	}
	parts := ecs.NewMap[comp.Part](ctx.World.ECS)
	for _, c := range channels {
		if !parts.Has(c.Entity) {
			continue
		}
		if p.Has("when") {
			holds, err := p.For(c).Bool("when")
			if err != nil {
				return err
			}
			if !holds {
				continue
			}
		}
		sever(ctx.World, ctx.Source, c.Entity, api.SeveredCut)
	}
	return nil
}

func applyAddStatus(ctx *effects.Context, p effects.Params) error {
	bodies, err := p.Bodies("bodies")
	if err != nil {
		return err
	}
	status, err := p.String("status")
	if err != nil {
		return err
	}
	ticks, err := p.Int("ticks")
	if err != nil {
		return err
	}
	w := ctx.World
	if err := effects.KnownIDs(map[string]string{"status": status}, knownIDs(w.Content)); err != nil {
		return err
	}
	if ticks < 1 {
		return fmt.Errorf("ticks must be at least 1, got %d", ticks)
	}
	statuses := ecs.NewMap[comp.Statuses](w.ECS)
	until := w.Now + int64(ticks)
	for _, body := range bodies {
		if !statuses.Has(body) {
			statuses.Add(body, &comp.Statuses{})
		}
		active := statuses.Get(body)
		entry := comp.StatusEntry{Status: api.ContentID(status), Until: until}
		at, found := slices.BinarySearchFunc(active.Active, entry, func(a, b comp.StatusEntry) int {
			return cmp.Compare(a.Status, b.Status)
		})
		if found {
			active.Active[at].Until = max(active.Active[at].Until, until)
			entry.Until = active.Active[at].Until
		} else {
			active.Active = slices.Insert(active.Active, at, entry)
		}
		w.Bus.Publish(
			api.StatusAdded{
				Body:   w.UIDOf(body),
				Status: entry.Status,
				Until:  entry.Until,
				Source: w.UIDOf(ctx.Source),
				Tick:   w.Now,
				Line:   statusLine(w, body, entry),
			},
		)
	}
	return nil
}

// expireStatuses removes every status whose tick has come, and the component once none is left.
func expireStatuses(w *sim.World) error {
	statuses := ecs.NewMap[comp.Statuses](w.ECS)
	var holders []ecs.Entity
	query := ecs.NewFilter1[comp.Statuses](w.ECS).Query()
	for query.Next() {
		holders = append(holders, query.Entity())
	}
	for _, holder := range holders {
		active := statuses.Get(holder)
		active.Active = slices.DeleteFunc(active.Active, func(e comp.StatusEntry) bool { return e.Until <= w.Now })
		if len(active.Active) == 0 {
			statuses.Remove(holder)
		}
	}
	return nil
}

func verifyAttach(raw map[string]string) error {
	_, hasItem := raw["item"]
	if effects.IsFormula(raw["from"]) {
		return nil
	}
	switch raw["from"] {
	case api.AttachedFromInventory:
		if !hasItem {
			return errors.New("from inventory needs an item")
		}
	case api.AttachedFromSevered:
		if hasItem {
			return errors.New("from severed takes no item")
		}
	default:
		return fmt.Errorf(
			"from must be %q or %q, got %q",
			api.AttachedFromInventory,
			api.AttachedFromSevered,
			raw["from"],
		)
	}
	return nil
}

// severedNear is the severed part with the lowest UID lying where the body stands, or zero.
func severedNear(w *sim.World, body ecs.Entity) ecs.Entity {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !w.ECS.Alive(body) || !positions.Has(body) {
		return ecs.Entity{}
	}
	at := *positions.Get(body)
	partOf := ecs.NewMap[comp.PartOf](w.ECS)
	var found []ecs.Entity
	query := ecs.NewFilter2[comp.Part, comp.Pos](w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		if *pos == at && !partOf.Has(query.Entity()) {
			found = append(found, query.Entity())
		}
	}
	if len(found) == 0 {
		return ecs.Entity{}
	}
	return slices.MinFunc(found, w.CompareUID)
}

func checkAttach(ctx *effects.Context, p effects.Params) (bool, api.Text) {
	w := ctx.World
	if !w.ECS.Alive(ctx.Target) {
		return false, api.Text{ID: nothingToAttach}
	}
	from, _ := p.String("from")
	if from == api.AttachedFromSevered {
		if severedNear(w, ctx.Target).IsZero() {
			return false, api.Text{ID: nothingToAttach}
		}
		return true, api.Text{}
	}
	item, _ := p.String("item")
	return inventoryHas(ctx, api.ContentID(item))
}

func applyAttach(ctx *effects.Context, p effects.Params) error {
	if ok, _ := checkAttach(ctx, p); !ok {
		return nil
	}
	w := ctx.World
	from, err := p.String("from")
	if err != nil {
		return err
	}
	body := ctx.Target
	var part ecs.Entity
	if from == api.AttachedFromSevered {
		part = severedNear(w, body)
		ecs.NewMap[comp.Pos](w.ECS).Remove(part)
	} else {
		item, err := p.String("item")
		if err != nil {
			return err
		}
		if part, err = takePiece(ctx, api.ContentID(item)); err != nil {
			return err
		}
	}
	ecs.NewMap[comp.PartOf](w.ECS).Add(part, &comp.PartOf{}, body)
	grafts, reattached := ecs.NewMap[comp.Graft](w.ECS), ecs.NewMap[comp.Reattached](w.ECS)
	switch {
	case from == api.AttachedFromSevered:
		if !grafts.Has(part) && !reattached.Has(part) {
			reattached.Add(part, &comp.Reattached{})
		}
	case !grafts.Has(part):
		item, err := p.String("item")
		if err != nil {
			return err
		}
		graft := &comp.Graft{Piece: api.ContentID(item)}
		// The piece's triggers now belong to the part, so they fire for it and not its holder.
		if ids := w.Content.ItemTriggerIDs(graft.Piece); len(ids) > 0 {
			ecs.NewMap[comp.Triggers](w.ECS).Add(part, &comp.Triggers{IDs: ids})
		}
		grafts.Add(part, graft)
	}
	w.Bus.Publish(
		api.PartAttached{
			Body:   w.UIDOf(body),
			Part:   w.UIDOf(part),
			From:   from,
			Source: w.UIDOf(ctx.Source),
			Tick:   w.Now,
		},
	)
	timeline.Reschedule(w, body)
	return nil
}

// takePiece takes the piece from the inventory and makes the part it becomes, not yet attached.
func takePiece(ctx *effects.Context, item api.ContentID) (ecs.Entity, error) {
	w := ctx.World
	def, ok := w.Content.Item(item)
	if !ok || def.Part == nil {
		return ecs.Entity{}, fmt.Errorf("item %q is not a piece that becomes a part", item)
	}
	template, ok := w.Content.Part(*def.Part)
	if !ok {
		return ecs.Entity{}, fmt.Errorf("item %q becomes unknown part %q", item, *def.Part)
	}
	if !inventory.Take(w, inventory.HolderOf(w, ctx.Source), item, 0, 1) {
		return ecs.Entity{}, fmt.Errorf("the inventory has no %q", item)
	}
	part := w.NewIdentified(*def.Part)
	ecs.NewMap[comp.Part](w.ECS).Add(part, &comp.Part{
		Template: *def.Part, Slot: template.Slot, Material: template.Material,
		Integrity: template.Integrity, MaxIntegrity: template.Integrity,
		Grants: append([]comp.Grant(nil), template.Grants...), Vital: template.Vital,
	})
	return part, nil
}

func applyReveal(ctx *effects.Context, p effects.Params) error {
	w := ctx.World
	knower := inventory.HolderOf(w, ctx.Source)
	if knower.IsZero() || !w.ECS.Alive(ctx.Target) {
		return nil
	}
	claim, err := p.String("claim")
	if err != nil {
		return err
	}
	origin, err := p.String("origin")
	if err != nil {
		return err
	}
	if err := effects.KnownIDs(map[string]string{"claim": claim, "origin": origin}, knownIDs(w.Content)); err != nil {
		return err
	}
	var channels []timeline.Channel
	if p.Has("parts") {
		channels, err = p.Channels("parts")
	} else {
		var all effects.ChannelSelector
		if all, err = effects.ParseChannels("target:all"); err == nil {
			channels, err = all.Resolve(ctx)
		}
	}
	if err != nil {
		return err
	}
	parts := ecs.NewMap[comp.Part](w.ECS)
	for _, c := range channels {
		if !parts.Has(c.Entity) {
			continue
		}
		value, err := p.For(c).String("value")
		if err != nil {
			return err
		}
		knowledge.Learn(w, knower, api.Fact{
			Subject: w.UIDOf(c.Entity), Claim: api.ContentID(claim), Value: value,
			Origin: api.ContentID(origin), Confidence: 1,
		})
	}
	return nil
}

// boundTo is whom the item is tied to when used in the context: the target, for a bound item.
func boundTo(ctx *effects.Context, item api.ContentID) api.UID {
	if def, ok := ctx.World.Content.Item(item); ok && def.Bound {
		return ctx.World.UIDOf(ctx.Target)
	}
	return 0
}

// inventoryHas says whether the source's inventory holds the item, and gives the reason when it does not.
func inventoryHas(ctx *effects.Context, item api.ContentID) (bool, api.Text) {
	w := ctx.World
	if inventory.Count(w, inventory.HolderOf(w, ctx.Source), item, boundTo(ctx, item)) > 0 {
		return true, api.Text{}
	}
	args := map[string]string{"item": string(item)}
	if def, ok := w.Content.Item(item); ok {
		args["item"] = w.Content.Render(api.Text{ID: def.Name})
	}
	return false, api.Text{ID: inventoryShortReason, Args: args}
}

// consumedItem is the item a consume names, or else the one its choice picked.
func consumedItem(ctx *effects.Context, p effects.Params) (string, error) {
	if p.Has("item") {
		return p.String("item")
	}
	choice, err := p.String("choice")
	if err != nil {
		return "", fmt.Errorf("consume needs an item or a choice: %w", err)
	}
	return ctx.Choices[choice], nil
}

func verifyConsume(raw map[string]string) error {
	_, item := raw["item"]
	_, choice := raw["choice"]
	if item == choice {
		return errors.New("consume needs exactly one of item and choice")
	}
	return nil
}

func checkConsume(ctx *effects.Context, p effects.Params) (bool, api.Text) {
	item, _ := consumedItem(ctx, p)
	id := api.ContentID(item)
	// A bound item held for some other body doesn't refuse: refusing would tell the player whose
	// keepsake they hold. The swing lands and the stack bound to this target, if any, is used.
	if def, ok := ctx.World.Content.Item(id); ok && def.Bound &&
		inventory.Usable(ctx.World, ctx.Source, ecs.Entity{})[item] > 0 {
		return true, api.Text{}
	}
	return inventoryHas(ctx, id)
}

func applyConsume(ctx *effects.Context, p effects.Params) error {
	item, err := consumedItem(ctx, p)
	if err != nil {
		return err
	}
	if err := effects.KnownIDs(map[string]string{"item": item}, knownIDs(ctx.World.Content)); err != nil {
		return err
	}
	id := api.ContentID(item)
	if !inventory.Take(ctx.World, inventory.HolderOf(ctx.World, ctx.Source), id, boundTo(ctx, id), 1) {
		return nil
	}
	def, _ := ctx.World.Content.Item(id)
	return effects.Apply(&effects.Context{World: ctx.World, Source: ctx.Source, Target: ctx.Source}, def.OnConsume)
}

// applySpawnFromPart makes a severed part a body of its own, at the cell where it lies. The body's
// identity is the template the part crawls as (its crawls_as), with that template's appearance, or
// else the part's own template; its faction and controller come from the parameters.
func applySpawnFromPart(ctx *effects.Context, p effects.Params) error {
	faction, err := p.String("faction")
	if err != nil {
		return err
	}
	controller, err := p.String("controller")
	if err != nil {
		return err
	}
	w := ctx.World
	if err := effects.KnownIDs(map[string]string{"faction": faction}, knownIDs(w.Content)); err != nil {
		return err
	}
	part := severedPart(ctx)
	if part.IsZero() {
		return nil
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	identities := ecs.NewMap[comp.Identity](w.ECS)
	at := *positions.Get(part)
	identity := identities.Get(part).Template
	var appearance *comp.Appearance
	if def, ok := w.Content.Part(identity); ok && def.CrawlsAs != "" {
		identity = def.CrawlsAs
		template, ok := w.Content.Template(identity)
		if !ok {
			return fmt.Errorf("part %q crawls as unknown template %q", def.Name, identity)
		}
		if template.Appearance != nil {
			looks := *template.Appearance
			appearance = &looks
		}
	}
	body := w.NewIdentified(identity)
	if appearance != nil {
		ecs.NewMap[comp.Appearance](w.ECS).Add(body, appearance)
	}
	ecs.NewMap[comp.Body](w.ECS).Add(body, &comp.Body{})
	positions.Add(body, &at)
	positions.Remove(part)
	ecs.NewMap[comp.Allegiance](w.ECS).Add(body, &comp.Allegiance{Faction: api.ContentID(faction)})
	ecs.NewMap[comp.Actor](w.ECS).Add(body, &comp.Actor{Controller: controller})
	ecs.NewMap[comp.PartOf](w.ECS).Add(part, &comp.PartOf{}, body)
	timeline.Reschedule(w, body)
	return nil
}

// severedPart is the target part when it is severed and lying somewhere, otherwise the severed part
// nearest the target body.
func severedPart(ctx *effects.Context) ecs.Entity {
	w := ctx.World
	partOf := ecs.NewMap[comp.PartOf](w.ECS)
	positions := ecs.NewMap[comp.Pos](w.ECS)
	parts := ecs.NewMap[comp.Part](w.ECS)
	if part := ctx.TargetPart; w.ECS.Alive(part) && !part.IsZero() &&
		parts.Has(part) && !partOf.Has(part) && positions.Has(part) {
		return part
	}
	return severedNear(w, ctx.Target)
}

func knownIDs(reg *content.Registry) map[string]func(api.ContentID) bool {
	return map[string]func(api.ContentID) bool{
		"claim":   effects.Known(reg.Claim),
		"faction": effects.Known(reg.Faction),
		"item":    effects.Known(reg.Item),
		"origin":  effects.Known(reg.Origin),
		"status":  effects.Known(reg.Status),
	}
}

func verifyKnownIDs(reg *content.Registry, raw map[string]string) error {
	return effects.KnownIDs(raw, knownIDs(reg))
}

func verifyAttachContent(reg *content.Registry, raw map[string]string) error {
	if err := verifyKnownIDs(reg, raw); err != nil {
		return err
	}
	if id, ok := raw["item"]; ok {
		if item, _ := reg.Item(api.ContentID(id)); item.Part == nil {
			return fmt.Errorf("item %q is not a piece that becomes a part", id)
		}
	}
	return nil
}

func nonNegative(p effects.Params, name string) (int, error) {
	value, err := p.Int(name)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must not be negative, got %d", name, value)
	}
	return value, nil
}

// partLine is the log text about a part of a body, or empty when the party perceives neither the
// body nor the source.
func partLine(w *sim.World, id api.ContentID, source, body, part ecs.Entity, extra map[string]string) api.Text {
	if !timeline.IsVisible(w, body) && !timeline.IsVisible(w, source) {
		return api.Text{}
	}
	args := map[string]string{"target": nameOf(w, body), "part": partName(w, part)}
	maps.Copy(args, extra)
	return api.Text{ID: id, Args: args}
}

func partName(w *sim.World, part ecs.Entity) string {
	template, ok := w.Content.Part(ecs.NewMap[comp.Part](w.ECS).Get(part).Template)
	if !ok {
		return ""
	}
	return w.Content.Render(api.Text{ID: template.Name})
}

// damageLine names who dealt the damage and with what when the context says so.
func damageLine(ctx *effects.Context, body, part ecs.Entity, lost, left int) api.Text {
	w := ctx.World
	args := map[string]string{"amount": strconv.Itoa(lost), "left": strconv.Itoa(left)}
	id := api.ContentID("text.combat.part_damaged_plain")
	if !ctx.Source.IsZero() && ctx.Action != "" {
		id = "text.combat.part_damaged"
		args["actor"], args["action"] = nameOf(w, ctx.Source), w.ActionName(ctx.Action)
	}
	return partLine(w, id, ctx.Source, body, part, args)
}

func statusLine(w *sim.World, body ecs.Entity, entry comp.StatusEntry) api.Text {
	if !timeline.IsVisible(w, body) {
		return api.Text{}
	}
	var name string
	if status, ok := w.Content.Status(entry.Status); ok {
		name = w.Content.Render(api.Text{ID: status.Name})
	}
	return api.Text{ID: "text.combat.status_added", Args: map[string]string{
		"target": nameOf(w, body), "status": name, "until": strconv.FormatInt(entry.Until, 10),
	}}
}

// applyRemoveCorpse removes the target body from the world if it is dead, and does nothing to a
// living one.
func applyRemoveCorpse(ctx *effects.Context, _ effects.Params) error {
	w := ctx.World
	if ctx.Target.IsZero() || !w.Alive(ctx.Target) || !ecs.NewMap[comp.Dead](w.ECS).Has(ctx.Target) {
		return nil
	}
	w.Despawn(ctx.Target)
	return nil
}
