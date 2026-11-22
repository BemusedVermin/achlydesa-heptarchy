package combat

import (
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
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
	severedReason     api.ContentID = "text.combat.severed"
	packShortReason   api.ContentID = "text.combat.pack_short"
	nothingToAttach   api.ContentID = "text.combat.no_severed_part"
	defaultRevealFrom api.ContentID = "source.sweep"
	materialClaim     api.ContentID = "claim.material"
)

// expireSystem is the name of the per-turn system that removes expired statuses.
const expireSystem = "combat.expire_statuses"

// Install registers the combat events, the primitives and the system that expires statuses. It
// needs effects installed first.
func Install(w *sim.World) error {
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.PartDamaged],
		api.RegisterEvent[api.PartSevered],
		api.RegisterEvent[api.PartAttached],
		api.RegisterEvent[api.StatusAdded],
	} {
		if err := register(w.Codec); err != nil {
			return err
		}
	}
	for _, p := range primitives() {
		if err := effects.Register(w, p); err != nil {
			return err
		}
	}
	return w.Systems.Register(sim.System{Name: expireSystem, Cadence: sim.PerTurn, Run: expireStatuses})
}

func primitives() []effects.Primitive {
	parts := map[string]effects.Kind{"parts": effects.KindChannels}
	return []effects.Primitive{
		{Name: "damage", Params: []string{"parts", "amount"}, Kinds: parts, Apply: applyDamage},
		{Name: "sever", Params: []string{"parts"}, Optional: []string{"when"}, Kinds: parts, Apply: applySever},
		{
			Name: "add_status", Params: []string{"bodies", "status", "ticks"},
			Kinds:         map[string]effects.Kind{"bodies": effects.KindBodies, "status": effects.KindText},
			VerifyContent: verifyAddStatusContent, Apply: applyAddStatus,
		},
		{
			Name: "attach_part", Params: []string{"from"}, Optional: []string{"item"},
			Kinds:  map[string]effects.Kind{"from": effects.KindText, "item": effects.KindText},
			Verify: verifyAttach, VerifyContent: verifyAttachContent, Apply: applyAttach, Check: checkAttach,
		},
		{Name: "raise_max_integrity", Params: []string{"parts", "amount"}, Kinds: parts, Apply: applyRaiseMaxIntegrity},
		{
			Name: "reveal", Params: []string{"claim"}, Optional: []string{"source"},
			Kinds:  map[string]effects.Kind{"claim": effects.KindText, "source": effects.KindText},
			Verify: verifyReveal, Apply: applyReveal,
		},
		{
			Name: "consume", Params: []string{"item"}, Kinds: map[string]effects.Kind{"item": effects.KindText},
			VerifyContent: verifyConsumeContent, Apply: applyConsume, Check: checkConsume,
		},
		{
			Name: "spawn_from_part", Params: []string{"faction", "controller"},
			Kinds:         map[string]effects.Kind{"faction": effects.KindText, "controller": effects.KindText},
			VerifyContent: verifySpawnContent, Apply: applySpawnFromPart,
		},
	}
}

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
		at := *positions.Get(body)
		if positions.Has(part) {
			*positions.Get(part) = at
		} else {
			positions.Add(part, &at)
		}
	}
	w.Bus.Publish(
		api.PartSevered{Body: w.UIDOf(body), Part: w.UIDOf(part), Cause: cause, Source: w.UIDOf(source), Tick: w.Now},
	)
	timeline.Reschedule(w, body)
}

func applyDamage(ctx *effects.Context, p effects.Params) error {
	channels, err := p.Channels("parts")
	if err != nil {
		return err
	}
	amount, err := p.Int("amount")
	if err != nil {
		return err
	}
	if amount < 0 {
		return fmt.Errorf("amount must not be negative, got %d", amount)
	}
	if amount == 0 {
		return nil
	}
	w := ctx.World
	parts := ecs.NewMap[comp.Part](w.ECS)
	for _, c := range channels {
		body := bodyOf(w, c.Entity)
		if body.IsZero() || !parts.Has(c.Entity) || parts.Get(c.Entity).Integrity <= 0 {
			continue
		}
		part := parts.Get(c.Entity)
		lost := min(amount, part.Integrity)
		part.Integrity -= lost
		w.Bus.Publish(api.PartDamaged{
			Body: w.UIDOf(body), Part: w.UIDOf(c.Entity), Amount: lost, Integrity: part.Integrity,
			Type: api.ContentID(ctx.Choices["type"]), Source: w.UIDOf(ctx.Source), Tick: w.Now,
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
	amount, err := p.Int("amount")
	if err != nil {
		return err
	}
	if amount < 0 {
		return fmt.Errorf("amount must not be negative, got %d", amount)
	}
	parts := ecs.NewMap[comp.Part](ctx.World.ECS)
	for _, c := range channels {
		if ctx.World.ECS.Alive(c.Entity) && parts.Has(c.Entity) {
			parts.Get(c.Entity).MaxIntegrity += amount
		}
	}
	return nil
}

func applySever(ctx *effects.Context, p effects.Params) error {
	if p.Has("when") {
		holds, err := p.Bool("when")
		if err != nil {
			return err
		}
		if !holds {
			return nil
		}
	}
	channels, err := p.Channels("parts")
	if err != nil {
		return err
	}
	parts := ecs.NewMap[comp.Part](ctx.World.ECS)
	for _, c := range channels {
		if parts.Has(c.Entity) {
			sever(ctx.World, ctx.Source, c.Entity, api.SeveredCut)
		}
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
	if _, ok := w.Content.Status(api.ContentID(status)); !ok {
		return fmt.Errorf("unknown status %q", status)
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
			return compareIDs(a.Status, b.Status)
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
			},
		)
	}
	return nil
}

func compareIDs(a, b api.ContentID) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
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
	switch raw["from"] {
	case api.AttachedFromPack:
		if !hasItem {
			return errors.New("from pack needs an item")
		}
	case api.AttachedFromSevered:
		if hasItem {
			return errors.New("from severed takes no item")
		}
	default:
		return fmt.Errorf("from must be %q or %q, got %q", api.AttachedFromPack, api.AttachedFromSevered, raw["from"])
	}
	return nil
}

// holderOf is who keeps the source's pack: its party when it is a member, otherwise itself.
func holderOf(w *sim.World, source ecs.Entity) ecs.Entity {
	if source.IsZero() || !w.ECS.Alive(source) {
		return ecs.Entity{}
	}
	members := ecs.NewMap[comp.MemberOf](w.ECS)
	if members.Has(source) {
		return members.GetRelation(source)
	}
	return source
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
	return slices.MinFunc(found, func(a, b ecs.Entity) int { return int(w.UIDOf(a)) - int(w.UIDOf(b)) })
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
	return packHas(ctx, api.ContentID(item))
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
	if grafts := ecs.NewMap[comp.Graft](w.ECS); !grafts.Has(part) {
		grafts.Add(part, &comp.Graft{})
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

// takePiece takes the piece from the pack and makes the part it becomes, not yet attached.
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
	if !inventory.Take(w, holderOf(w, ctx.Source), item, 0, 1) {
		return ecs.Entity{}, fmt.Errorf("the pack has no %q", item)
	}
	part := w.NewIdentified(*def.Part)
	ecs.NewMap[comp.Part](w.ECS).Add(part, &comp.Part{
		Template: *def.Part, Slot: template.Slot, Material: template.Material,
		Integrity: template.Integrity, MaxIntegrity: template.Integrity,
		Grants: append([]comp.Grant(nil), template.Grants...), Vital: template.Vital,
	})
	return part, nil
}

func verifyReveal(raw map[string]string) error {
	if raw["claim"] != string(materialClaim) {
		return fmt.Errorf("can only reveal %q, got %q", materialClaim, raw["claim"])
	}
	return nil
}

func applyReveal(ctx *effects.Context, p effects.Params) error {
	w := ctx.World
	knower := holderOf(w, ctx.Source)
	if knower.IsZero() || !w.ECS.Alive(ctx.Target) {
		return nil
	}
	claim, err := p.String("claim")
	if err != nil {
		return err
	}
	source := string(defaultRevealFrom)
	if p.Has("source") {
		if source, err = p.String("source"); err != nil {
			return err
		}
	}
	parts := ecs.NewMap[comp.Part](w.ECS)
	for _, part := range anatomy.Parts(w, ctx.Target) {
		knowledge.Learn(w, knower, api.Fact{
			Subject: w.UIDOf(part), Claim: api.ContentID(claim), Value: string(parts.Get(part).Material),
			Source: api.ContentID(source), Confidence: 1,
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

// packHas says whether the source's pack holds the item, and gives the reason when it does not.
func packHas(ctx *effects.Context, item api.ContentID) (bool, api.Text) {
	w := ctx.World
	if inventory.Count(w, holderOf(w, ctx.Source), item, boundTo(ctx, item)) > 0 {
		return true, api.Text{}
	}
	args := map[string]string{"item": string(item)}
	if def, ok := w.Content.Item(item); ok {
		args["item"] = w.Content.Render(api.Text{ID: def.Name})
	}
	return false, api.Text{ID: packShortReason, Args: args}
}

func checkConsume(ctx *effects.Context, p effects.Params) (bool, api.Text) {
	item, _ := p.String("item")
	return packHas(ctx, api.ContentID(item))
}

func applyConsume(ctx *effects.Context, p effects.Params) error {
	item, err := p.String("item")
	if err != nil {
		return err
	}
	id := api.ContentID(item)
	if _, ok := ctx.World.Content.Item(id); !ok {
		return fmt.Errorf("unknown item %q", item)
	}
	inventory.Take(ctx.World, holderOf(ctx.World, ctx.Source), id, boundTo(ctx, id), 1)
	return nil
}

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
	if _, ok := w.Content.Faction(api.ContentID(faction)); !ok {
		return fmt.Errorf("unknown faction %q", faction)
	}
	part := severedPart(ctx)
	if part.IsZero() {
		return nil
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	identities := ecs.NewMap[comp.Identity](w.ECS)
	at := *positions.Get(part)
	body := w.NewIdentified(identities.Get(part).Template)
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

func verifyAddStatusContent(reg *content.Registry, raw map[string]string) error {
	if _, ok := reg.Status(api.ContentID(raw["status"])); !ok {
		return fmt.Errorf("unknown status %q", raw["status"])
	}
	return nil
}

func verifyAttachContent(reg *content.Registry, raw map[string]string) error {
	id, ok := raw["item"]
	if !ok {
		return nil
	}
	item, found := reg.Item(api.ContentID(id))
	if !found {
		return fmt.Errorf("unknown item %q", id)
	}
	if item.Part == nil {
		return fmt.Errorf("item %q is not a piece that becomes a part", id)
	}
	return nil
}

func verifyConsumeContent(reg *content.Registry, raw map[string]string) error {
	if _, ok := reg.Item(api.ContentID(raw["item"])); !ok {
		return fmt.Errorf("unknown item %q", raw["item"])
	}
	return nil
}

func verifySpawnContent(reg *content.Registry, raw map[string]string) error {
	if _, ok := reg.Faction(api.ContentID(raw["faction"])); !ok {
		return fmt.Errorf("unknown faction %q", raw["faction"])
	}
	return nil
}
