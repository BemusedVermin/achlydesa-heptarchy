package sim

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentUseIsSafe(t *testing.T) {
	w := NewWorld(1)
	var wg sync.WaitGroup
	const workers = 8

	uids := make(chan uint64, workers*100)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				uids <- uint64(w.NewUID())
			}
		}()
	}
	for worker := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Subscribe(w.Bus, fmt.Sprint("sub", worker), func(*World, pingEvent) {})
			w.Bus.Publish(pingEvent{N: worker})
			if _, err := w.Bus.Drain(w); err != nil {
				t.Error(err)
			}
			_ = w.Systems.Register(System{Name: fmt.Sprintf("p.s%d", worker), Run: noop})
			_ = w.Systems.Order(PerEntry)
			_ = w.Commands.Kinds()
			_, _ = w.Views.Build(w)
		}()
	}
	wg.Wait()
	close(uids)

	seen := map[uint64]bool{}
	for uid := range uids {
		if seen[uid] {
			t.Fatalf("UID %d handed out twice", uid)
		}
		seen[uid] = true
	}
	if len(seen) != workers*100 {
		t.Fatalf("got %d UIDs, want %d", len(seen), workers*100)
	}
	if _, err := w.Bus.Drain(w); err != nil { // flush anything a skipped concurrent Drain left
		t.Fatal(err)
	}
	if err := w.Systems.Freeze(); err != nil {
		t.Fatal(err)
	}
	var runs sync.WaitGroup
	for range workers {
		runs.Add(1)
		go func() {
			defer runs.Done()
			if err := w.Systems.Run(w, PerEntry); err != nil {
				t.Error(err)
			}
		}()
	}
	runs.Wait()
}
