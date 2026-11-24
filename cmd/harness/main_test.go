package main

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

// recordLog plays a short game and writes its log to a file, returning the path and the hash.
func recordLog(t *testing.T) (string, string) {
	t.Helper()
	reg, err := content.Load(contenttest.Dir())
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(game.Config{Seed: 3, Content: reg})
	if err != nil {
		t.Fatal(err)
	}
	if err := harness.Play(g, harness.NewRandomPolicy(3), 20); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := g.Codec().WriteLog(&buf, g.History()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "run.log")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	hash, err := g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	return path, hex.EncodeToString(hash[:])
}

func runReplay(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := root()
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(append([]string{"replay", "--seed", "3", "--content", contenttest.Dir()}, args...))
	err := cmd.Execute()
	return strings.TrimSpace(out.String()), err
}

func TestReplayExpect(t *testing.T) {
	path, hash := recordLog(t)
	if got, err := runReplay(t, path, "--expect", hash); err != nil || got != hash {
		t.Errorf("matching --expect: got %q, err %v; want %q, nil", got, err, hash)
	}
	if _, err := runReplay(t, path, "--expect", "deadbeef"); err == nil {
		t.Error("wrong --expect returned no error, so the process would exit 0")
	}
}
