package harness

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"strconv"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// policyStream is the second PCG word of the policy generator, so it differs from any core stream.
const policyStream = 0x6861726e657373

// Policy picks what a run does next.
type Policy interface {
	// Choose returns the option to take, or false when there is nothing to do.
	Choose(view api.View) (api.Option, bool)
}

// RandomPolicy picks uniformly among the enabled options.
type RandomPolicy struct{ rng *rand.Rand }

// NewRandomPolicy returns a RandomPolicy whose choices depend only on seed.
func NewRandomPolicy(seed uint64) *RandomPolicy {
	return &RandomPolicy{rng: rand.New(rand.NewPCG(seed, policyStream))}
}

// Choose implements [Policy].
func (p *RandomPolicy) Choose(view api.View) (api.Option, bool) {
	var enabled []api.Option
	for _, option := range view.Options {
		if option.Enabled {
			enabled = append(enabled, option)
		}
	}
	if len(enabled) == 0 {
		return api.Option{}, false
	}
	return enabled[p.rng.IntN(len(enabled))], true
}

// Header is the first CSV row.
var Header = []string{
	"seed", "step", "tick", "day", "phase", "map", "q", "r", "depth", "events", "rejected",
}

// RunConfig says what Run plays.
type RunConfig struct {
	Seeds     int
	SeedStart uint64
	Steps     int
	Content   *content.Registry
	// NewPolicy makes the policy for a seed. Nil means a RandomPolicy.
	NewPolicy func(seed uint64) Policy
	// Logger receives one line per seed. Nil discards them.
	Logger *slog.Logger
}

// Run plays cfg.Seeds games and writes one CSV row per step to out.
func Run(cfg RunConfig, out io.Writer) error {
	if cfg.Content == nil {
		return errors.New("harness: no content")
	}
	if cfg.NewPolicy == nil {
		cfg.NewPolicy = func(seed uint64) Policy { return NewRandomPolicy(seed) }
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	writer := csv.NewWriter(out)
	if err := writer.Write(Header); err != nil {
		return fmt.Errorf("harness: %w", err)
	}
	for i := range cfg.Seeds {
		seed := cfg.SeedStart + uint64(i)
		if err := runSeed(cfg, seed, writer); err != nil {
			return fmt.Errorf("harness: seed %d: %w", seed, err)
		}
	}
	writer.Flush()
	return writer.Error()
}

func runSeed(cfg RunConfig, seed uint64, writer *csv.Writer) error {
	g, err := game.New(game.Config{Seed: seed, Content: cfg.Content})
	if err != nil {
		return err
	}
	policy := cfg.NewPolicy(seed)
	for step := 1; step <= cfg.Steps; step++ {
		view, err := g.View()
		if err != nil {
			return err
		}
		option, ok := policy.Choose(view)
		if !ok {
			break
		}
		events, err := g.Apply(option.Command)
		if err != nil {
			return fmt.Errorf("step %d: %w", step, err)
		}
		row, err := stepRow(g, seed, step, events)
		if err != nil {
			return err
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	hash, err := g.Hash()
	if err != nil {
		return err
	}
	cfg.Logger.Info("seed finished", "seed", seed, "hash", fmt.Sprintf("%x", hash))
	return nil
}

func stepRow(g *game.Game, seed uint64, step int, events []api.Event) ([]string, error) {
	view, err := g.View()
	if err != nil {
		return nil, err
	}
	var day, phase, mapID, q, r, depth string
	if view.Clock != nil {
		day, phase = strconv.Itoa(view.Clock.Day), view.Clock.Phase
	}
	if view.Party != nil && len(view.Party.Members) > 0 {
		if pos, ok := g.PositionOf(view.Party.Members[0].UID); ok {
			mapID = strconv.FormatUint(uint64(pos.Map), 10)
			q, r = strconv.Itoa(pos.At.Hex.Q), strconv.Itoa(pos.At.Hex.R)
			depth = strconv.Itoa(pos.At.Depth)
		}
	}
	rejected := 0
	for _, event := range events {
		if _, ok := event.(api.Rejected); ok {
			rejected++
		}
	}
	return []string{
		strconv.FormatUint(seed, 10), strconv.Itoa(step), strconv.FormatInt(view.Tick, 10),
		day, phase, mapID, q, r, depth, strconv.Itoa(len(events)), strconv.Itoa(rejected),
	}, nil
}
