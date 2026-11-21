package sim

import (
	"errors"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestViewsBuildInSortedOrder(t *testing.T) {
	w := NewWorld(1)
	w.Now = 42
	appendLabel := func(label string) func(*World, *api.View) error {
		return func(_ *World, v *api.View) error {
			v.Options = append(v.Options, api.Option{ID: label})
			return nil
		}
	}
	for _, b := range []ViewBuilder{
		{Name: "z.last", After: []string{"a.first"}, Build: appendLabel("z")},
		{Name: "m.mid", After: []string{"a.first"}, Build: appendLabel("m")},
		{Name: "a.first", Build: appendLabel("a")},
	} {
		if err := w.Views.Register(b); err != nil {
			t.Fatal(err)
		}
	}
	view, err := w.Views.Build(w)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, option := range view.Options {
		ids = append(ids, option.ID)
	}
	if got := strings.Join(ids, ","); got != "a,m,z" {
		t.Errorf("builder order = %s, want a,m,z", got)
	}
	if view.Tick != 42 {
		t.Errorf("Tick = %d, want 42", view.Tick)
	}
}

func TestViewsErrors(t *testing.T) {
	w := NewWorld(1)
	build := func(*World, *api.View) error { return nil }
	if err := w.Views.Register(ViewBuilder{Name: "a.a", Build: build}); err != nil {
		t.Fatal(err)
	}
	if err := w.Views.Register(ViewBuilder{Name: "a.a", Build: build}); err == nil {
		t.Error("duplicate builder should error")
	}

	if err := w.Views.Register(ViewBuilder{Name: "b.b", After: []string{"a.ghost"}, Build: build}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Views.Build(w); err == nil || !strings.Contains(err.Error(), "a.ghost") {
		t.Errorf("unknown After: err = %v", err)
	}

	w = NewWorld(1)
	boom := errors.New("boom")
	_ = w.Views.Register(ViewBuilder{Name: "a.a", Build: func(*World, *api.View) error { return boom }})
	if _, err := w.Views.Build(w); !errors.Is(err, boom) {
		t.Errorf("builder error: err = %v, want wrapped boom", err)
	}
}
