package sim

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// NameOf is the rendered name of a body, from its template.
func (w *World) NameOf(e ecs.Entity) string {
	return w.Content.Render(w.Content.TemplateName(w.TemplateOf(e)))
}

// ActionName is the rendered name of an action, or empty when the action is unknown.
func (w *World) ActionName(id api.ContentID) string {
	def, ok := w.Content.Action(id)
	if !ok {
		return ""
	}
	return w.Content.Render(api.Text{ID: def.Name})
}
