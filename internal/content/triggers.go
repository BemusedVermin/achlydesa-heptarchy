package content

import (
	"fmt"
	"slices"

	"github.com/expr-lang/expr"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Kinds of content that can carry triggers.
const (
	OwnerTemplate = "template"
	OwnerMaterial = "material"
	OwnerStatus   = "status"
	OwnerItem     = "item"
)

// TriggerSite is one trigger and the content that carries it.
type TriggerSite struct {
	// ID names the trigger: the owner's ID, then "#", then the trigger's index in its list.
	ID      string
	Owner   string        // one of the Owner kinds
	OwnerID api.ContentID // the template, material, status or item ID
	Path    string        // where it sits in content, for error messages
	Trigger Trigger
}

// TriggerID is the ID of the index-th trigger of the owner with the given ID.
func TriggerID(owner api.ContentID, index int) string { return fmt.Sprintf("%s#%d", owner, index) }

// TriggerSites lists every trigger in content: templates, then materials, statuses and items, each
// in ID order and each owner's triggers in the order written.
func (r *Registry) TriggerSites() []TriggerSite {
	var sites []TriggerSite
	add := func(owner, collection, id string, triggers []Trigger) {
		for i, trigger := range triggers {
			sites = append(sites, TriggerSite{
				ID: TriggerID(api.ContentID(id), i), Owner: owner, OwnerID: api.ContentID(id),
				Path: fmt.Sprintf("%s[%q].triggers[%d]", collection, id, i), Trigger: trigger,
			})
		}
	}
	for _, id := range sortedKeys(r.bundle.Templates) {
		add(OwnerTemplate, "templates", id, r.bundle.Templates[id].Triggers)
	}
	for _, id := range sortedKeys(r.bundle.Materials) {
		add(OwnerMaterial, "materials", id, r.bundle.Materials[id].Triggers)
	}
	for _, id := range sortedKeys(r.bundle.Statuses) {
		add(OwnerStatus, "statuses", id, r.bundle.Statuses[id].Triggers)
	}
	for _, id := range sortedKeys(r.bundle.Items) {
		add(OwnerItem, "items", id, r.bundle.Items[id].Triggers)
	}
	return sites
}

// TemplateTriggerIDs lists the IDs of a template's triggers, or none.
func (r *Registry) TemplateTriggerIDs(template api.ContentID) []string {
	var ids []string
	for i := range r.bundle.Templates[string(template)].Triggers {
		ids = append(ids, TriggerID(template, i))
	}
	return ids
}

// ItemTriggerIDs lists the IDs of an item's triggers, or none.
func (r *Registry) ItemTriggerIDs(item api.ContentID) []string {
	var ids []string
	for i := range r.bundle.Items[string(item)].Triggers {
		ids = append(ids, TriggerID(item, i))
	}
	return ids
}

// checkTriggers checks each trigger's role and effects, and compiles its condition. Whether the
// event kind exists is for the game to check: only it knows which kinds are registered.
func (r *Registry) checkTriggers() []error {
	p := r.problems()
	for _, site := range r.TriggerSites() {
		trigger := site.Trigger
		if trigger.On == "" {
			p.addf("%s.on: required", site.Path)
		}
		if !slices.Contains(triggerRoles(), trigger.Role) {
			p.addf("%s.role: unknown role %q (want one of %v)", site.Path, trigger.Role, triggerRoles())
		}
		if trigger.If != "" {
			if _, err := r.Compile(trigger.If, EffectEnv{}, expr.AsBool()); err != nil {
				p.addf("%s.if: %w", site.Path, err)
			}
		}
		if len(trigger.Do) == 0 {
			p.addf("%s.do: at least one effect is required", site.Path)
		}
	}
	return p.errs
}
