package content

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestTemplateName(t *testing.T) {
	r := sampleRegistry(t)
	tests := []struct {
		template api.ContentID
		want     api.ContentID
	}{
		{"", "text.template.unknown"},
		{"template.creature.moving_clay", "text.template.creature.moving_clay"},
		{"creature.no_prefix", "text.template.creature.no_prefix"},
	}
	for _, tt := range tests {
		t.Run(string(tt.template), func(t *testing.T) {
			if got := r.TemplateName(tt.template); got.ID != tt.want {
				t.Fatalf("TemplateName(%q) = %q, want %q", tt.template, got.ID, tt.want)
			}
		})
	}
}
