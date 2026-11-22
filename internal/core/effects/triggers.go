package effects

import (
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// triggerSystem names the bus subscriber that runs content triggers.
const triggerSystem = "effects.triggers"

// triggerIndex is the content's triggers grouped by the event kind they listen for, in the order
// content lists them. It is rebuilt when the world's content is swapped.
type triggerIndex struct {
	reg    *content.Registry
	byKind map[string][]content.TriggerSite
}

// installTriggers gives new entities the triggers of their template and subscribes the runner.
//
// The runner is one subscriber to every event rather than one per kind: a typed subscription needs
// the event's Go type, which content cannot name, and one subscriber that looks the kind up in the
// index also keeps working when content is swapped for one that listens to more kinds.
func installTriggers(w *sim.World) error {
	err := w.RegisterSpawner(sim.Spawner{Name: "triggers", Apply: spawnTriggers})
	if err != nil {
		return err
	}
	w.Bus.SubscribeAll(triggerSystem, runTriggers)
	return nil
}

// spawnTriggers records the trigger IDs of the template an entity came from.
func spawnTriggers(w *sim.World, e ecs.Entity, t content.Template) error {
	if len(t.Triggers) == 0 {
		return nil
	}
	identity := ecs.NewMap[comp.Identity](w.ECS).Get(e)
	ecs.NewMap[comp.Triggers](w.ECS).Add(e, &comp.Triggers{IDs: w.Content.TemplateTriggerIDs(identity.Template)})
	return nil
}

func indexOf(w *sim.World) *triggerIndex {
	st := stateOf(w)
	if st.triggers != nil && st.triggers.reg == w.Content {
		return st.triggers
	}
	index := &triggerIndex{reg: w.Content, byKind: map[string][]content.TriggerSite{}}
	for _, site := range w.Content.TriggerSites() {
		index.byKind[site.Trigger.On] = append(index.byKind[site.Trigger.On], site)
	}
	st.triggers = index
	return index
}

// validateTriggers checks that every trigger listens for a registered event kind.
func validateTriggers(w *sim.World) []error {
	var problems []error
	for _, site := range w.Content.TriggerSites() {
		if site.Trigger.On != "" && !w.Codec.HasEvent(site.Trigger.On) {
			problems = append(problems, fmt.Errorf("%s.on: unknown event kind %q", site.Path, site.Trigger.On))
		}
	}
	return problems
}

// firing is one trigger that matched, and the entity it matched through.
type firing struct {
	carrier ecs.Entity
	uid     api.UID
	site    content.TriggerSite
	order   int // the site's place in the index, for a stable order among one carrier's triggers
}

// runTriggers applies the triggers that listen for e. Matches fire in UID order of the entity that
// carries them, then in the order content lists them.
func runTriggers(w *sim.World, e api.Event) {
	subjected, ok := e.(api.Subjected)
	if !ok {
		return
	}
	sites := indexOf(w).byKind[e.EventKind()]
	if len(sites) == 0 {
		return
	}
	subjects := subjected.Subjects()
	var firings []firing
	for order, site := range sites {
		for _, uid := range roleUIDs(subjects, site.Trigger.Role) {
			carrier, alive := w.ByUID(uid)
			if alive && carries(w, carrier, site) {
				firings = append(firings, firing{carrier: carrier, uid: uid, site: site, order: order})
			}
		}
	}
	slices.SortFunc(firings, func(a, b firing) int {
		if a.uid != b.uid {
			if a.uid < b.uid {
				return -1
			}
			return 1
		}
		return a.order - b.order
	})
	var choices map[string]string
	if chosen, ok := e.(api.Chosen); ok {
		choices = chosen.Choices()
	}
	for _, f := range firings {
		ctx := &Context{World: w, Choices: choices}
		ctx.Source = entityOf(w, subjects[api.SubjectSource])
		ctx.Target = entityOf(w, subjects[api.SubjectTarget])
		ctx.TargetPart = entityOf(w, subjects[api.SubjectPart])
		if err := fire(ctx, f.site.Trigger); err != nil {
			w.Bus.Fail(fmt.Errorf("effects: trigger %s: %w", f.site.Path, err))
			return
		}
	}
}

// fire checks one trigger's condition and applies its effects.
func fire(ctx *Context, trigger content.Trigger) error {
	if trigger.If != "" {
		st := stateOf(ctx.World)
		holds, err := (&params{ctx: ctx, st: st, raw: map[string]string{"if": trigger.If}}).Bool("if")
		if err != nil {
			return err
		}
		if !holds {
			return nil
		}
	}
	return Apply(ctx, trigger.Do)
}

// roleUIDs lists the subjects a trigger of the role could apply to, in UID order.
func roleUIDs(subjects map[string]api.UID, role string) []api.UID {
	var uids []api.UID
	if role == content.RoleAny {
		for _, uid := range subjects {
			if uid != 0 && !slices.Contains(uids, uid) {
				uids = append(uids, uid)
			}
		}
		slices.Sort(uids)
		return uids
	}
	if uid := subjects[role]; uid != 0 {
		uids = append(uids, uid)
	}
	return uids
}

func entityOf(w *sim.World, uid api.UID) ecs.Entity {
	if uid == 0 {
		return ecs.Entity{}
	}
	e, _ := w.ByUID(uid)
	return e
}

// carries reports whether the entity is something the trigger's owner applies to: a template's
// entity, a part made of the material, a body under the status or a holder of the item.
func carries(w *sim.World, e ecs.Entity, site content.TriggerSite) bool {
	switch site.Owner {
	case content.OwnerTemplate:
		triggers := ecs.NewMap[comp.Triggers](w.ECS)
		return triggers.Has(e) && slices.Contains(triggers.Get(e).IDs, site.ID)
	case content.OwnerMaterial:
		parts := ecs.NewMap[comp.Part](w.ECS)
		return parts.Has(e) && parts.Get(e).Material == site.OwnerID
	case content.OwnerStatus:
		statuses := ecs.NewMap[comp.Statuses](w.ECS)
		return statuses.Has(e) && slices.ContainsFunc(statuses.Get(e).Active, func(entry comp.StatusEntry) bool {
			return entry.Status == site.OwnerID
		})
	case content.OwnerItem:
		packs := ecs.NewMap[comp.Inventory](w.ECS)
		return packs.Has(e) && slices.ContainsFunc(packs.Get(e).Stacks, func(stack comp.ItemStack) bool {
			return stack.Item == site.OwnerID && stack.Count > 0
		})
	}
	return false
}
