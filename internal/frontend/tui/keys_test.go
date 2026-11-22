package tui

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func stepView(states map[string]bool) api.View {
	var v api.View
	for id, enabled := range states {
		v.Options = append(v.Options, api.Option{ID: id, Enabled: enabled})
	}
	return v
}

func TestVerticalOption(t *testing.T) {
	open := map[string]bool{"step/ne": true, "step/nw": true, "step/se": true, "step/sw": true}
	tests := []struct {
		name   string
		view   map[string]bool
		up     bool
		last   string
		want   string
		wantOK bool
	}{
		{"first up prefers NE", open, true, "", "step/ne", true},
		{"up flips to NW", open, true, "step/ne", "step/nw", true},
		{"up flips back to NE", open, true, "step/nw", "step/ne", true},
		{"first down prefers SE", open, false, "", "step/se", true},
		{"down flips to SW", open, false, "step/se", "step/sw", true},
		{"down after up starts at SE", open, false, "step/ne", "step/se", true},
		{"NE blocked falls back to NW", map[string]bool{"step/ne": false, "step/nw": true}, true, "", "step/nw", true},
		{"NE absent falls back to NW", map[string]bool{"step/nw": true}, true, "", "step/nw", true},
		{
			"both blocked picks an offered one",
			map[string]bool{"step/ne": false, "step/nw": false},
			true,
			"",
			"step/ne",
			true,
		},
		{"none offered", map[string]bool{"step/e": true}, true, "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := VerticalOption(stepView(tt.view), tt.up, tt.last)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("got (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
