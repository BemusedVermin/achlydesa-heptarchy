package harness_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

// BenchmarkStep times one harness step on the sample content: build the view, pick a random
// enabled option, apply it. A game that runs out of options is replaced, off the clock.
func BenchmarkStep(b *testing.B) {
	reg, err := loadRegistry()
	if err != nil {
		b.Fatal(err)
	}
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
	b.ResetTimer()
	for range b.N {
		view, err := g.View()
		if err != nil {
			b.Fatal(err)
		}
		option, ok := policy.Choose(view)
		if !ok {
			b.StopTimer()
			g, policy = newGame()
			b.StartTimer()
			continue
		}
		if _, err := g.Apply(option.Command); err != nil {
			b.Fatal(err)
		}
	}
}

func loadRegistry() (*content.Registry, error) { return content.Load(sampleDir) }
