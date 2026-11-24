package save_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/save"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

// rig is a game on the sample content's start scenario, with the world kept for tests that edit it.
type rig struct {
	cfg game.Config
	g   *game.Game
	w   *sim.World
}

func newBundle(t testing.TB) content.Bundle {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.StepOnlyParty(&bundle)
	contenttest.StandingCrawler(&bundle)
	return bundle
}

func newRig(t testing.TB, seed uint64, bundle content.Bundle) *rig {
	t.Helper()
	reg := contenttest.Registry(t, bundle)
	r := &rig{}
	r.cfg = game.Config{Seed: seed, Content: reg}
	r.g, r.w, _ = gametest.Start(t, r.cfg)
	return r
}

// reload saves g and loads the bytes into a new game with the same config.
func reload(t testing.TB, cfg game.Config, g *game.Game) (*rig, []string) {
	t.Helper()
	loaded, w, warnings := gametest.Reload(t, cfg, g)
	return &rig{cfg: cfg, g: loaded, w: w}, warnings
}

func hash(t rapid.TB, g *game.Game) [32]byte {
	t.Helper()
	sum, err := g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	return sum
}

// nth is a policy that takes the enabled option at its index, wrapping around.
type nth int

func (n nth) Choose(view api.View) (api.Option, bool) {
	var enabled []api.Option
	for _, option := range view.Options {
		if option.Enabled {
			enabled = append(enabled, option)
		}
	}
	if len(enabled) == 0 {
		return api.Option{}, false
	}
	return enabled[int(n)%len(enabled)], true
}

// pick applies the enabled option of the view that choice selects, wrapping around.
func pick(t rapid.TB, g *game.Game, choice int) {
	t.Helper()
	if _, ok, err := harness.Step(g, nth(choice)); err != nil || !ok {
		t.Fatalf("pick %d: ok %v, err %v", choice, ok, err)
	}
}

func TestSaveThenLoadAtTickZeroHasTheSameHash(t *testing.T) {
	r := newRig(t, 1, newBundle(t))
	loaded, warnings := reload(t, r.cfg, r.g)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if hash(t, loaded.g) != hash(t, r.g) {
		t.Error("hash changed across save and load")
	}
	if loaded.g.Current() != r.g.Current() || loaded.g.Current() == 0 {
		t.Errorf("current = %d, want %d", loaded.g.Current(), r.g.Current())
	}
}

func TestSaveThenLoadWithASpawnedActorHasTheSameHash(t *testing.T) {
	bundle := newBundle(t)
	contenttest.NoStart(&bundle)
	cfg := game.Config{
		Seed:    1,
		Content: contenttest.Registry(t, bundle),
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			_, err := w.Spawn("template.person.reaper", nil)
			return err
		},
	}
	g, _, cfg := gametest.Start(t, cfg)
	loaded, warnings := reload(t, cfg, g)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if hash(t, loaded.g) != hash(t, g) {
		t.Error("hash changed across save and load")
	}
	if loaded.g.Current() != g.Current() || g.Current() == 0 {
		t.Errorf("current = %d, want %d", loaded.g.Current(), g.Current())
	}
}

func TestSaveThenLoadWithThePartyAndAFactHasTheSameHash(t *testing.T) {
	bundle := newBundle(t)
	contenttest.NoStart(&bundle)
	var party ecs.Entity
	cfg := game.Config{
		Seed:    1,
		Content: contenttest.Registry(t, bundle),
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			var err error
			party, err = w.Spawn("template.party.main", nil)
			return err
		},
	}
	g, w, cfg := gametest.Start(t, cfg)
	before := hash(t, g)
	members := anatomy.Members(w, party)
	if len(members) == 0 {
		t.Fatal("the party has no members")
	}
	knowledge.Learn(w, members[0], api.Fact{
		Subject: w.UIDOf(party), Claim: "claim.material", Value: "material.clay", Origin: "origin.sweep",
		Confidence: 0.8,
	})
	if hash(t, g) == before {
		t.Fatal("the fact is not in the hash")
	}
	loaded, warnings := reload(t, cfg, g)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if hash(t, loaded.g) != hash(t, g) {
		t.Error("hash changed across save and load")
	}
}

func TestSavingTwiceGivesTheSameBytes(t *testing.T) {
	r := newRig(t, 1, newBundle(t))
	for i := range 20 {
		pick(t, r.g, i)
	}
	var first, second bytes.Buffer
	if err := r.g.Save(&first); err != nil {
		t.Fatal(err)
	}
	if err := r.g.Save(&second); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Error("two saves of one game differ")
	}
}

func TestAMapEditSurvivesSaveAndLoad(t *testing.T) {
	r := newRig(t, 1, newBundle(t))
	at := api.Cell{Hex: api.Hex{Q: 2, R: 2}, Depth: 1}
	m := r.w.Maps.Get(1)
	before, err := m.At(at)
	if err != nil {
		t.Fatal(err)
	}
	want := hexmap.Terrain{ID: "terrain.packed_clay", Solid: true, Opaque: true}
	if before.ID == want.ID {
		want = hexmap.Terrain{ID: "terrain.air"}
	}
	if err := m.Set(at, want); err != nil {
		t.Fatal(err)
	}
	if hash(t, r.g) == hash(t, newRig(t, 1, newBundle(t)).g) {
		t.Fatal("the edit is not in the hash")
	}
	loaded, _ := reload(t, r.cfg, r.g)
	got, err := loaded.w.Maps.Get(1).At(at)
	if err != nil || got != want {
		t.Errorf("cell = %+v, %v; want %+v", got, err, want)
	}
	if hash(t, loaded.g) != hash(t, r.g) {
		t.Error("hash changed across save and load")
	}
}

// continuesIdentically plays before, saves and loads, then plays after on both games, and fails
// when their hashes differ after the load or at the end. The hash covers the history and the random
// streams, so games that part never meet again, and checking at the end catches a split anywhere.
func continuesIdentically(t rapid.TB, cfg game.Config, before, after []int) {
	t.Helper()
	g, err := game.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, choice := range before {
		pick(t, g, choice)
	}
	var buf bytes.Buffer
	if err := g.Save(&buf); err != nil {
		t.Fatal(err)
	}
	loaded, _, err := game.Load(cfg, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if hash(t, loaded) != hash(t, g) {
		t.Fatal("hash changed across save and load")
	}
	for _, choice := range after {
		pick(t, g, choice)
		pick(t, loaded, choice)
	}
	if hash(t, loaded) != hash(t, g) {
		t.Fatalf("hashes differ %d commands after loading", len(after))
	}
}

func TestLoadingAfterRandomPlayContinuesIdentically(t *testing.T) {
	reg := contenttest.Registry(t, newBundle(t))
	for _, tt := range []struct {
		name          string
		seed          uint64
		before, after []int
	}{
		{"seed 0, a failure rapid once found", 0, []int{491, 8, 720}, []int{222, 823, 592}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			continuesIdentically(t, game.Config{Seed: tt.seed, Content: reg}, tt.before, tt.after)
		})
	}
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Uint64().Draw(rt, "seed")
		before := rapid.SliceOfN(rapid.IntRange(0, 1000), 0, 20).Draw(rt, "before")
		after := rapid.SliceOfN(rapid.IntRange(0, 1000), 0, 10).Draw(rt, "after")
		continuesIdentically(rt, game.Config{Seed: seed, Content: reg}, before, after)
	})
}

func TestChangedContentLoadsWithAWarning(t *testing.T) {
	r := newRig(t, 1, newBundle(t))
	for i := range 5 {
		pick(t, r.g, i)
	}
	var buf bytes.Buffer
	if err := r.g.Save(&buf); err != nil {
		t.Fatal(err)
	}
	changed := newBundle(t)
	changed.Timeline.DefaultSpeed++
	other := newRig(t, 1, changed)
	loaded, warnings, err := game.Load(other.cfg, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "content") {
		t.Errorf("warnings = %v, want one about content", warnings)
	}
	if _, err := loaded.View(); err != nil {
		t.Errorf("view of the loaded game: %v", err)
	}
}

func TestHistoryAndLogSurviveSaveAndLoad(t *testing.T) {
	r := newRig(t, 1, newBundle(t))
	for i := range 8 {
		pick(t, r.g, i)
	}
	loaded, _ := reload(t, r.cfg, r.g)
	if len(loaded.g.History()) != 8 {
		t.Errorf("history has %d commands, want 8", len(loaded.g.History()))
	}
	want, _ := r.g.View()
	got, _ := loaded.g.View()
	wantJSON, _ := json.Marshal(want)
	gotJSON, _ := json.Marshal(got)
	if !bytes.Equal(wantJSON, gotJSON) {
		t.Errorf("view changed across save and load:\n%s\n%s", wantJSON, gotJSON)
	}
}

func TestReadRejectsWhatItCannotUse(t *testing.T) {
	header := func(format string, version int) string {
		return fmt.Sprintf(`{"header":{"format":%q,"version":%d}}`, format, version)
	}
	tests := []struct {
		name, doc, want string
	}{
		{"not json", `nope`, "not a JSON object"},
		{"no header", `{}`, "no header"},
		{"another format", header("other", 1), "is not"},
		{"newer version", header(save.Format, save.CurrentVersion+1), "newer than this game understands"},
		{"negative version", header(save.Format, -1), "not valid"},
		{"a version with no migration", header(save.Format, 0), "no migration from version 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := save.Read(strings.NewReader(tt.doc))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one containing %q", err, tt.want)
			}
		})
	}
}

// editHeader rewrites the header object of a document.
func editHeader(t *testing.T, doc map[string]json.RawMessage, edit func(header map[string]json.RawMessage)) {
	t.Helper()
	var header map[string]json.RawMessage
	if err := json.Unmarshal(doc["header"], &header); err != nil {
		t.Fatal(err)
	}
	edit(header)
	raw, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}
	doc["header"] = raw
}

func TestAMigrationRunsOnAnOlderDocument(t *testing.T) {
	r := newRig(t, 1, newBundle(t))
	var buf bytes.Buffer
	if err := r.g.Save(&buf); err != nil {
		t.Fatal(err)
	}
	// Turn the current save into a version 0 one that keeps its seed under an old name.
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	editHeader(t, doc, func(header map[string]json.RawMessage) {
		header["old_seed"] = header["seed"]
		delete(header, "seed")
		header["version"] = json.RawMessage("0")
	})
	old, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	renameSeed := func(doc map[string]json.RawMessage) error {
		editHeader(t, doc, func(header map[string]json.RawMessage) {
			header["seed"] = header["old_seed"]
			delete(header, "old_seed")
		})
		return nil
	}
	got, err := save.ReadWith(
		bytes.NewReader(old),
		save.CurrentVersion,
		map[int]save.Migration{0: renameSeed, 1: func(map[string]json.RawMessage) error { return nil }},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Header.Version != save.CurrentVersion || got.Header.Seed != 1 {
		t.Errorf("header = %+v, want the current version and seed 1", got.Header)
	}
}
