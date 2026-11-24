package harness_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

// BenchmarkStep times one harness step on the sample content: build the view, pick a random
// enabled option, apply it. A game that runs out of options is replaced, off the clock.
func BenchmarkStep(b *testing.B) {
	reg := contenttest.Sample(b)
	var seed uint64 = 1
	newGame := func() (*game.Game, *harness.RandomPolicy) {
		g, err := game.New(game.Config{Seed: seed, Content: reg})
		if err != nil {
			b.Fatal(err)
		}
		policy := harness.NewRandomPolicy(seed)
		seed++
		return g, policy
	}
	g, policy := newGame()
	for b.Loop() {
		_, ok, err := harness.Step(g, policy)
		if err != nil {
			b.Fatal(err)
		}
		if !ok {
			b.StopTimer()
			g, policy = newGame()
			b.StartTimer()
		}
	}
}
