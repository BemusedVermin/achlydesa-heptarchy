package harness_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func TestReplayTraceListsCommandsAndKeepsTheHash(t *testing.T) {
	reg := contenttest.Sample(t)
	g, err := game.New(game.Config{Seed: 1, Content: reg})
	if err != nil {
		t.Fatal(err)
	}
	policy := harness.NewRandomPolicy(1)
	for range 12 {
		if _, ok, err := harness.Step(g, policy); err != nil || !ok {
			t.Fatalf("step: ok=%v err=%v", ok, err)
		}
	}
	var log bytes.Buffer
	if err := g.Codec().WriteLog(&log, g.History()); err != nil {
		t.Fatal(err)
	}
	plain, err := harness.Replay(reg, 1, bytes.NewReader(log.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	var trace bytes.Buffer
	traced, err := harness.ReplayTrace(reg, 1, bytes.NewReader(log.Bytes()), &trace)
	if err != nil {
		t.Fatal(err)
	}
	if traced != plain {
		t.Errorf("trace hash %s, want %s", traced, plain)
	}
	got := trace.String()
	path := filepath.Join("testdata", "trace.golden")
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
	if got != string(want) {
		t.Errorf("trace differs from the golden file; run go test ./internal/harness -update\n got:\n%s", got)
	}
}
