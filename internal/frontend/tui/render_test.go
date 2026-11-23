package tui_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/frontend/tui"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func config(t *testing.T) game.Config {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	// These tests walk and read goldens drawn for a party that can only step.
	for _, id := range []string{"template.person.reaper", "template.person.salken"} {
		person := bundle.Templates[id]
		person.Actions = []api.ContentID{"action.step"}
		bundle.Templates[id] = person
	}
	// The sample moving clay hunts; these goldens were drawn with it standing still.
	movingClay := bundle.Templates["template.creature.moving_clay"]
	movingClay.Actor = &content.ActorDef{Controller: "still"}
	movingClay.Actions = nil
	movingClay.Lists = nil
	bundle.Templates["template.creature.moving_clay"] = movingClay
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	return game.Config{Seed: 1, Content: reg}
}

func newApp(t *testing.T) *tui.App {
	t.Helper()
	g, err := game.New(config(t))
	if err != nil {
		t.Fatal(err)
	}
	app, err := tui.NewApp(g)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestContentMsgSwapsOrLogsTheError(t *testing.T) {
	app := newApp(t)
	reg := contenttest.Sample(t)
	app.Update(tui.ContentMsg{Registry: reg})
	last := app.State.Log[len(app.State.Log)-1].Text
	if !strings.Contains(last, "reloaded") {
		t.Errorf("last log line = %q, want the reload warning", last)
	}
	app.Update(tui.ContentMsg{Err: os.ErrInvalid})
	last = app.State.Log[len(app.State.Log)-1].Text
	if !strings.Contains(last, "reload failed") {
		t.Errorf("last log line = %q, want the failure", last)
	}
	if app.Err() != nil {
		t.Errorf("a failed reload must not end the game: %v", app.Err())
	}
}

// press sends a key to the app as the terminal would.
func press(app *tui.App, text string) {
	code := []rune(text)[0]
	if text == "tab" {
		code = tea.KeyTab
	}
	app.Update(tea.KeyPressMsg{Code: code, Text: text})
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the golden file; run go test ./internal/frontend/tui -update\n got:\n%s", name, got)
	}
}

func TestRenderMapStart(t *testing.T) {
	app := newApp(t)
	golden(t, "map_start", tui.RenderMap(app.State.Map, app.State.Party, 44, 23))
}

func TestRenderMapAfterWalkingAwayDimsMemory(t *testing.T) {
	app := newApp(t)
	for range 8 {
		press(app, "d")
	}
	out := tui.RenderMap(app.State.Map, app.State.Party, 44, 23)
	golden(t, "map_walked", out)
	remembered := 0
	for _, c := range app.State.Map.Cells {
		if !c.Visible {
			remembered++
		}
	}
	if remembered == 0 {
		t.Error("no remembered cells after walking away")
	}
}

func TestRenderMapSlopeShadesSmoothly(t *testing.T) {
	m := &api.MapView{}
	for q := range 13 {
		m.Cells = append(
			m.Cells,
			api.CellView{Hex: api.Hex{Q: q}, Glyph: ".", Color: "green", Visible: true, Below: q - 4},
		)
	}
	out := tui.RenderMap(m, nil, 30, 1)
	golden(t, "map_slope", out)
	// Every step has its own color: thirteen cells, thirteen distinct shades.
	seen := map[string]bool{}
	for _, part := range strings.Split(out, "\x1b[") {
		if i := strings.IndexByte(part, 'm'); i > 0 && strings.HasPrefix(part, "38;2;") {
			seen[part[:i]] = true
		}
	}
	if len(seen) != 13 {
		t.Errorf("got %d distinct shades, want 13", len(seen))
	}
}

func TestRenderMapSmallPaneScrollsToTheParty(t *testing.T) {
	app := newApp(t)
	out := tui.RenderMap(app.State.Map, app.State.Party, 20, 8)
	golden(t, "map_small", out)
	if lines := strings.Count(out, "\n") + 1; lines != 8 {
		t.Errorf("%d lines, want 8", lines)
	}
	if !strings.Contains(out, "@") {
		t.Error("the party scrolled out of view")
	}
}

func TestRenderTimeline(t *testing.T) {
	tl := &api.TimelineView{Now: 10, Window: 6, Entries: []api.TimelineEntry{
		{Tick: 10, Kind: api.EntryReady, Glyph: "@", Color: "red"},
		{Tick: 12, Kind: api.EntryReady, Glyph: "@", Color: "cyan"},
		{Tick: 12, Kind: api.EntryLands, Glyph: "r", Color: "red"},
		{Tick: 12, Kind: api.EntryReady, Glyph: "r", Color: "red"},
		{Tick: 15, Kind: api.EntryLands, Glyph: "r", Color: "red"},
	}}
	golden(t, "timeline", tui.RenderTimeline(tl, 40, 2))
	if got := len(tui.EntriesAt(tl, 2)); got != 3 {
		t.Errorf("entries at offset 2 = %d, want 3", got)
	}
}

func TestDisabledMoveShowsItsReasonAndSendsNothing(t *testing.T) {
	app := newApp(t)
	var blocked string
	for _, o := range app.State.Options {
		if strings.HasPrefix(o.ID, "step/") && !o.Enabled {
			blocked = o.ID
		}
	}
	// Walk east until the wall disables a step.
	for range 30 {
		if blocked != "" {
			break
		}
		press(app, "d")
		for _, o := range app.State.Options {
			if o.ID == "step/e" && !o.Enabled {
				blocked = o.ID
			}
		}
	}
	if blocked != "step/e" {
		t.Fatalf("never blocked going east (blocked = %q)", blocked)
	}
	before := len(app.Game.History())
	press(app, "d")
	if app.Status == "" {
		t.Error("status line is empty after a blocked move")
	}
	if got := len(app.Game.History()); got != before {
		t.Errorf("%d commands sent, want none", got-before)
	}
}

func TestPlayRecordsAReplayThatReproducesTheHash(t *testing.T) {
	app := newApp(t)
	for _, k := range []string{"d", "d", "e", ".", ",", "tab", "f", "a"} {
		press(app, k)
	}
	if app.Err() != nil {
		t.Fatal(app.Err())
	}
	if len(app.Game.History()) < 5 {
		t.Fatalf("only %d commands recorded", len(app.Game.History()))
	}
	var log bytes.Buffer
	if err := tui.WriteReplay(app.Game, &log); err != nil {
		t.Fatal(err)
	}
	cmds, err := app.Game.Codec().ReadLog(&log)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := game.Replay(config(t), cmds)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := app.Game.Hash()
	got, _ := replayed.Hash()
	if want != got {
		t.Errorf("replay hash %x, want %x", got, want)
	}
}

func TestViewDrawsEveryPane(t *testing.T) {
	app := newApp(t)
	app.Width, app.Height = 100, 30
	out := app.View().Content
	for _, want := range []string{"Day 1", "Control:", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("view lacks %q:\n%s", want, out)
		}
	}
}
