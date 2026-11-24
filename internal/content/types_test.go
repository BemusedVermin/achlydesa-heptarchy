package content

import (
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestSampleTypesLoadAndNameIsNeverRefused(t *testing.T) {
	reg := sampleRegistry(t)
	tests := []struct {
		held map[string]int
		want bool
	}{{map[string]int{"item.name": 1}, true}, {map[string]int{}, true}}
	for _, tc := range tests {
		got, _, err := reg.ModsRequires("type.name", ActionEnv{Inventory: tc.held})
		if err != nil || got != tc.want {
			t.Errorf("name with inventory %v = %v, %v; want %v", tc.held, got, err, tc.want)
		}
	}
	if got, _, err := reg.ModsRequires("type.cut", ActionEnv{}); err != nil || !got {
		t.Errorf("cut = %v, %v; want always allowed", got, err)
	}
	for _, warning := range reg.Warnings() {
		if strings.Contains(warning, "hostile") {
			t.Errorf("warning %q, want no hostility warnings", warning)
		}
	}
}

func TestOneSidedHostilityWarnsButLoads(t *testing.T) {
	bundle := sampleBundle(t)
	deep := bundle.Factions["faction.wild"]
	deep.HostileTo = []api.ContentID{"faction.crawling"} // only the party's side of the feud is left one-sided
	bundle.Factions["faction.wild"] = deep
	reg, err := New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	hostility := 0
	for _, warning := range reg.Warnings() {
		if strings.Contains(warning, "hostile") {
			hostility++
		}
	}
	if hostility != 1 {
		t.Errorf("warnings = %v, want one about hostility", reg.Warnings())
	}
}

func TestBadTypeContentIsRejected(t *testing.T) {
	bundle := sampleBundle(t)
	bad := bundle.Types["type.press"]
	bad.Requires = "no_such_field == 1"
	bundle.Types["type.press"] = bad
	burn := bundle.Types["type.burn"]
	burn.Lists = map[string][]string{"consumes": {"item.nothing"}}
	bundle.Types["type.burn"] = burn
	if _, err := New(bundle); err == nil {
		t.Error("content with a bad condition and an unknown item loaded")
	}
}
