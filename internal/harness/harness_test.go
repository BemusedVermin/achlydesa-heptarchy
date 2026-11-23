package harness_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

const sampleDir = "../../content/data"

func load(t *testing.T) *content.Registry {
	t.Helper()
	bundle, err := content.LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	// The sample crawler hunts and kills random walkers, which would end runs early. These tests
	// count rows, so they keep it standing.
	crawler := bundle.Templates["template.creature.crawler"]
	crawler.Actor = &content.ActorDef{Controller: "rest"}
	crawler.Actions = nil
	crawler.Lists = nil
	bundle.Templates["template.creature.crawler"] = crawler
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func runCSV(t *testing.T, reg *content.Registry, seeds, steps int) []byte {
	t.Helper()
	var buf bytes.Buffer
	cfg := harness.RunConfig{Seeds: seeds, SeedStart: 1, Steps: steps, Content: reg}
	if err := harness.Run(cfg, &buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestRunWritesOneRowPerStep(t *testing.T) {
	data := runCSV(t, load(t), 3, 200)
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(rows), 1+3*200; got != want {
		t.Fatalf("rows = %d, want %d (header plus 600)", got, want)
	}
	if strings.Join(rows[0], ",") != strings.Join(harness.Header, ",") {
		t.Errorf("header = %v", rows[0])
	}
}

func TestRunIsByteIdentical(t *testing.T) {
	reg := load(t)
	if a, b := runCSV(t, reg, 2, 100), runCSV(t, reg, 2, 100); !bytes.Equal(a, b) {
		t.Fatal("two runs with the same arguments differ")
	}
}

func TestReplayPrintsTheGamesHash(t *testing.T) {
	reg := load(t)
	g, err := game.New(game.Config{Seed: 7, Content: reg})
	if err != nil {
		t.Fatal(err)
	}
	policy := harness.NewRandomPolicy(7)
	for range 50 {
		view, err := g.View()
		if err != nil {
			t.Fatal(err)
		}
		option, ok := policy.Choose(view)
		if !ok {
			break
		}
		if _, err := g.Apply(option.Command); err != nil {
			t.Fatal(err)
		}
	}
	var log bytes.Buffer
	if err := g.Codec().WriteLog(&log, g.History()); err != nil {
		t.Fatal(err)
	}
	hash, err := g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%x", hash)
	got, err := harness.Replay(reg, 7, &log)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("replay hash = %s, want %s", got, want)
	}
}
