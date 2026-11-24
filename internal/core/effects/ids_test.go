package effects_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
)

func TestKnownIDs(t *testing.T) {
	lookup := func(id api.ContentID) (struct{}, bool) { return struct{}{}, id == "status.slow" }
	known := map[string]func(api.ContentID) bool{"status": effects.Known(lookup), "item": effects.Known(lookup)}
	tests := []struct {
		name string
		raw  map[string]string
		want string
	}{
		{"known", map[string]string{"status": "status.slow"}, ""},
		{"absent", map[string]string{"ticks": "3"}, ""},
		{"unknown", map[string]string{"status": "status.nope"}, `unknown status "status.nope"`},
		{
			"both unknown, in name order",
			map[string]string{"status": "status.nope", "item": "item.nope"},
			"unknown item \"item.nope\"\nunknown status \"status.nope\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := effects.KnownIDs(tt.raw, known)
			got := ""
			if err != nil {
				got = err.Error()
			}
			if got != tt.want {
				t.Errorf("KnownIDs = %q, want %q", got, tt.want)
			}
		})
	}
}
