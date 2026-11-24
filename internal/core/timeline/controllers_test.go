package timeline_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

func TestControllerNamedFindsRegisteredControllersButNeverThePlayer(t *testing.T) {
	fx := newFixture(t, 1)
	tests := []struct {
		name string
		want bool
	}{
		{"record", true},
		{timeline.IdleController, true},
		{timeline.PlayerController, false},
		{"no.such.controller", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if c, ok := timeline.ControllerNamed(fx.w, tt.name); ok != tt.want || (c != nil) != tt.want {
				t.Errorf("ControllerNamed(%q) = %v, %v; want found %v", tt.name, c, ok, tt.want)
			}
		})
	}
}
