package content

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testDebounce = 40 * time.Millisecond

	// burstDebounce is long enough that a stalled runner can't split the burst.
	burstDebounce = 500 * time.Millisecond
)

type reload struct {
	reg *Registry
	err error
}

func TestWatchOneCallbackPerBurst(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "maps")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	calls := make(chan reload, 16)
	stop, err := watch(dir, burstDebounce, func(reg *Registry, err error) { calls <- reload{reg, err} })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	for range 5 {
		if err := os.WriteFile(filepath.Join(sub, "a.cue"), []byte("package data\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case got := <-calls:
		if got.err == nil {
			t.Fatal("a package with no content should fail to load, got no error")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("no callback after the burst")
	}
	select {
	case <-calls:
		t.Fatal("a burst produced more than one callback")
	case <-time.After(2 * burstDebounce):
	}
}

func TestWatchIgnoresOtherFiles(t *testing.T) {
	dir := t.TempDir()
	calls := make(chan reload, 4)
	stop, err := watch(dir, testDebounce, func(reg *Registry, err error) { calls <- reload{reg, err} })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-calls:
		t.Fatal("a non-CUE file triggered a reload")
	case <-time.After(3 * testDebounce):
	}
}

func TestWatchMissingDir(t *testing.T) {
	if _, err := Watch(filepath.Join(t.TempDir(), "nope"), func(*Registry, error) {}); err == nil {
		t.Fatal("want an error for a missing directory")
	}
}
