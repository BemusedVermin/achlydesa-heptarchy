package combat

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Names of the choices forms ask for, and the action tag that marks a form needing a graftable target.
const (
	typeChoice  = "type"
	pieceChoice = "piece"
	graftTag    = "graft"
)

// Text IDs of the reasons the type checks give.
const (
	immuneReason       api.ContentID = "text.combat.immune"
	typeBarredReason   api.ContentID = "text.combat.type_barred"
	notGraftableReason api.ContentID = "text.combat.not_graftable"
)

// installTypes registers what makes a type change a form: the choices it is picked through, the
// checks on a target and the changes to timing and effects. It also registers the spawner of fighters.
func installTypes(w *sim.World) error {
	if err := w.RegisterSpawner(sim.Spawner{Name: "combatant", Apply: spawnCombatant}); err != nil {
		return err
	}
	if err := actions.RegisterChoice(w, typeChoice, typeChoices); err != nil {
		return err
	}
	if err := actions.RegisterChoice(w, pieceChoice, pieceChoices); err != nil {
		return err
	}
	actions.AddValidator(w, validateIntent)
	actions.AddModifier(w, modifyIntent)
	return nil
}

// spawnCombatant gives a fighter its side, its innate types and whether it takes grafts.
func spawnCombatant(w *sim.World, e ecs.Entity, t content.Template) error {
	def := t.Combatant
	if def == nil {
		return nil
	}
	ecs.NewMap[comp.Combatant](w.ECS).Add(e, &comp.Combatant{Types: slices.Clone(def.Types)})
	ecs.NewMap[comp.Allegiance](w.ECS).Add(e, &comp.Allegiance{Faction: def.Faction})
	if def.Graftable {
		ecs.NewMap[comp.Graftable](w.ECS).Add(e, &comp.Graftable{})
	}
	return nil
}

// stacksOf is the stacks of the actor's pack.
func stacksOf(w *sim.World, actor ecs.Entity) []comp.ItemStack {
	holder := holderOf(w, actor)
	packs := ecs.NewMap[comp.Inventory](w.ECS)
	if holder.IsZero() || !w.ECS.Alive(holder) || !packs.Has(holder) {
		return nil
	}
	return packs.Get(holder).Stacks
}

// typeChoices is the "type" choice: the body's innate types, plus each consumable type whose item is
// in its party's pack, in ID order.
func typeChoices(w *sim.World, actor ecs.Entity) []string {
	var offered []string
	if combatants := ecs.NewMap[comp.Combatant](w.ECS); w.ECS.Alive(actor) && combatants.Has(actor) {
		for _, id := range combatants.Get(actor).Types {
			offered = append(offered, string(id))
		}
	}
	stacks := stacksOf(w, actor)
	for _, id := range w.Content.TypeIDs() {
		def, _ := w.Content.Type(id)
		if def.Innate || def.Consumes == nil {
			continue
		}
		if slices.ContainsFunc(stacks, func(s comp.ItemStack) bool { return s.Item == *def.Consumes && s.Count > 0 }) {
			offered = append(offered, string(id))
		}
	}
	slices.Sort(offered)
	return slices.Compact(offered)
}

// pieceChoices is the "piece" choice: each piece in the pack that becomes a part, in item ID order.
func pieceChoices(w *sim.World, actor ecs.Entity) []string {
	var offered []string
	for _, stack := range stacksOf(w, actor) {
		if item, ok := w.Content.Item(stack.Item); ok && item.Part != nil && stack.Count > 0 {
			offered = append(offered, string(stack.Item))
		}
	}
	return slices.Compact(offered) // stacks are sorted by item
}

// targetMaterials lists the materials of the parts an intent aims at: the one part, or every attached
// part of the target body when no part is named.
func targetMaterials(w *sim.World, in actions.Intent) []api.ContentID {
	parts := ecs.NewMap[comp.Part](w.ECS)
	var list []ecs.Entity
	switch {
	case !in.TargetPart.IsZero() && w.ECS.Alive(in.TargetPart) && parts.Has(in.TargetPart):
		list = []ecs.Entity{in.TargetPart}
	case !in.Target.IsZero() && w.ECS.Alive(in.Target) && in.Target != in.Actor:
		list = anatomy.Parts(w, in.Target)
	}
	materials := make([]api.ContentID, len(list))
	for i, part := range list {
		materials[i] = parts.Get(part).Material
	}
	return materials
}

// immuneTo says whether everything the intent aims at is immune to the type. Aiming at nothing is not.
func immuneTo(w *sim.World, in actions.Intent, kind api.ContentID) bool {
	materials := targetMaterials(w, in)
	if len(materials) == 0 {
		return false
	}
	return !slices.ContainsFunc(materials, func(id api.ContentID) bool {
		material, ok := w.Content.Material(id)
		return !ok || !slices.Contains(material.ImmuneTo, kind)
	})
}

// ownerOf is whom the type's item is tied to, for the type's condition: the target when the pack
// holds an item bound to it, otherwise whoever the first bound stack of the item names, otherwise 0.
func ownerOf(w *sim.World, in actions.Intent, def content.DamageType) api.UID {
	if def.Consumes == nil {
		return 0
	}
	if item, ok := w.Content.Item(*def.Consumes); !ok || !item.Bound {
		return 0
	}
	var first api.UID
	for _, stack := range stacksOf(w, in.Actor) {
		if stack.Item != *def.Consumes || stack.Count <= 0 {
			continue
		}
		if stack.BoundTo == w.UIDOf(in.Target) {
			return stack.BoundTo
		}
		if first == 0 {
			first = stack.BoundTo
		}
	}
	return first
}

// materialTags lists the tags of the materials of the parts an intent aims at, sorted without repeats.
func materialTags(w *sim.World, in actions.Intent) []string {
	var tags []string
	for _, id := range targetMaterials(w, in) {
		if material, ok := w.Content.Material(id); ok {
			tags = append(tags, material.Tags...)
		}
	}
	slices.Sort(tags)
	return slices.Compact(tags)
}

// validateIntent refuses an intent whose chosen type cannot be used on its target, and a form that
// needs a graftable target aimed at one that is not.
func validateIntent(w *sim.World, in actions.Intent) (bool, api.Text) {
	def, ok := w.Content.Action(in.Action)
	if !ok {
		return true, api.Text{}
	}
	if slices.Contains(def.Tags, graftTag) &&
		(in.Target.IsZero() || !w.ECS.Alive(in.Target) || !ecs.NewMap[comp.Graftable](w.ECS).Has(in.Target)) {
		return false, api.Text{ID: notGraftableReason}
	}
	kind := api.ContentID(in.Choices[typeChoice])
	if kind == "" {
		return true, api.Text{}
	}
	typeDef, ok := w.Content.Type(kind)
	if !ok {
		return false, api.Text{ID: "text.actions.bad_choice"}
	}
	if immuneTo(w, in, kind) {
		return false, api.Text{
			ID:   immuneReason,
			Args: map[string]string{"type": w.Content.Render(api.Text{ID: typeDef.Name})},
		}
	}
	allowed, err := w.Content.TypeAllowed(kind, content.TypeEnv{
		TargetUID: uint64(w.UIDOf(in.Target)), ItemBoundTo: uint64(ownerOf(w, in, typeDef)),
		TargetTags: materialTags(w, in),
	})
	if err != nil || !allowed {
		return false, api.Text{
			ID:   typeBarredReason,
			Args: map[string]string{"type": w.Content.Render(api.Text{ID: typeDef.Name})},
		}
	}
	return true, api.Text{}
}

// modifyIntent adds the chosen type's wind-up, recovery and landing effects to any form it is loaded
// into, and spends the type's item when the swing starts, whether it connects or not.
func modifyIntent(w *sim.World, in actions.Intent) actions.Modifier {
	kind := api.ContentID(in.Choices[typeChoice])
	def, ok := w.Content.Type(kind)
	if !ok {
		return actions.Modifier{}
	}
	env := content.ActionEnv{Speed: timeline.Speed(w, in.Actor), Used: len(in.Channels), Choice: in.Choices}
	windup, recovery, err := w.Content.TypeTicks(kind, env)
	if err != nil {
		windup, recovery = 0, 0 // formulas are compiled at load; one that still fails adds nothing
	}
	mod := actions.Modifier{Windup: windup, Recovery: recovery, OnLand: slices.Clone(def.OnLand)}
	if def.Consumes != nil {
		mod.OnStart = []content.EffectRef{{Do: "consume", Params: map[string]string{"item": string(*def.Consumes)}}}
	}
	return mod
}
