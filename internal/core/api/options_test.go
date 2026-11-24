package api

import "testing"

func TestParseLeadOption(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		want   UID
		wantOK bool
	}{
		{"round trip", LeadOption(42), 42, true},
		{"another option", ControlOption(ControlFollow), 0, false},
		{"bare prefix", OptionLeadPrefix, 0, false},
		{"bad UID", OptionLeadPrefix + "x7", 0, false},
		{"negative UID", OptionLeadPrefix + "-1", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseLeadOption(tt.id)
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("ParseLeadOption(%q) = %d, %v; want %d, %v", tt.id, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
