package gametest

import (
	"bytes"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// New builds a game on the sample content, after mutate has edited a fresh bundle (nil leaves it as it
// is). Each configure function may set fields of the config, such as Spawn, before the game starts.
// It returns the game, its world and the config, which Reload and game.Load need.
func New(
	t testing.TB,
	seed uint64,
	mutate func(*content.Bundle),
	configure ...func(*game.Config),
) (*game.Game, *sim.World, game.Config) {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	if mutate != nil {
		mutate(&bundle)
	}
	cfg := game.Config{Seed: seed, Content: contenttest.Registry(t, bundle)}
	for _, f := range configure {
		f(&cfg)
	}
	return Start(t, cfg)
}

// Start builds a game from cfg and returns it with its world, keeping any Install the config has.
func Start(t testing.TB, cfg game.Config) (*game.Game, *sim.World, game.Config) {
	t.Helper()
	var world *sim.World
	cfg.Install = capture(cfg.Install, &world)
	g, err := game.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return g, world, cfg
}

// Reload saves g and loads the bytes into a new game with the same config. It returns the new game,
// its world and the load warnings.
func Reload(t testing.TB, cfg game.Config, g *game.Game) (*game.Game, *sim.World, []string) {
	t.Helper()
	var buf bytes.Buffer
	if err := g.Save(&buf); err != nil {
		t.Fatal(err)
	}
	var world *sim.World
	cfg.Install = capture(cfg.Install, &world)
	loaded, warnings, err := game.Load(cfg, &buf)
	if err != nil {
		t.Fatal(err)
	}
	return loaded, world, warnings
}

func capture(next func(*sim.World, *clock.Clock) error, world **sim.World) func(*sim.World, *clock.Clock) error {
	return func(w *sim.World, c *clock.Clock) error {
		*world = w
		if next != nil {
			return next(w, c)
		}
		return nil
	}
}
