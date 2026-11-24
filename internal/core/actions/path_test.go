package actions_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestEveryOptionPathNamesEachSegmentOfItsID(t *testing.T) {
	fx := newFixture(t, 1, opts{contentChoices: true, mutate: withWater})
	checked := 0
	for _, option := range fx.view().Options {
		if option.ID == api.OptionWaitEvent || option.ID == api.OptionWaitPhase {
			continue
		}
		segments := strings.Split(option.ID, "/")
		if len(option.Path) != len(segments) {
			t.Errorf(
				"%s: path %q has %d labels for %d segments",
				option.ID,
				option.Path,
				len(option.Path),
				len(segments),
			)
			continue
		}
		for i, label := range option.Path {
			_, notNumber := strconv.ParseUint(segments[i], 10, 64)
			raw := label == segments[i] && (notNumber == nil || strings.Contains(segments[i], "."))
			if label == "" || raw || strings.Contains(label, "missing text") {
				t.Errorf("%s: segment %q reads %q, want a rendered name", option.ID, segments[i], label)
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no option to check")
	}
}
