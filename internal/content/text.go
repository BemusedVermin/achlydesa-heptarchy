package content

import (
	"strings"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Render turns player-facing text into a string, filling "{{.name}}" placeholders from t.Args.
// An unknown ID renders as "[missing text: <id>]"; Render never panics.
func (r *Registry) Render(t api.Text) string {
	parsed, ok := r.templates[t.ID]
	if !ok {
		return "[missing text: " + string(t.ID) + "]"
	}
	var out strings.Builder
	if err := parsed.Execute(&out, t.Args); err != nil {
		return "[bad text: " + string(t.ID) + "]"
	}
	return out.String()
}
