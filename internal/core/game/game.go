package game

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Config says how to start a game.
type Config struct {
	Seed    uint64
	Content *content.Registry
	// Extra, when set, runs after the standard packages are installed and before the scheduler
	// freezes. Tests and the harness use it to add what a scenario needs, such as a spawned actor.
	Extra func(w *sim.World, clk *clock.Clock) error
}

// Game is one running game. It is not safe for concurrent use.
type Game struct {
	world   *sim.World
	clock   *clock.Clock
	history []api.Command
	events  []api.Event // delivered since the last Apply returned
	log     []api.LogLine
	broken  error
}

// New builds a game from cfg, installs every package, and plays the timeline forward to the first
// player decision.
func New(cfg Config) (*Game, error) {
	if cfg.Content == nil {
		return nil, errors.New("game: no content")
	}
	g := &Game{world: sim.NewWorld(cfg.Seed), clock: clock.New(cfg.Content.Calendar())}
	g.world.Content = cfg.Content
	if err := g.installLog(); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if err := installAll(g.world, g.clock); err != nil {
		return nil, fmt.Errorf("game: install: %w", err)
	}
	if cfg.Extra != nil {
		if err := cfg.Extra(g.world, g.clock); err != nil {
			return nil, fmt.Errorf("game: extra: %w", err)
		}
	}
	if err := effects.ValidateContent(g.world); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if err := actions.ValidateContent(g.world); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if err := g.world.Systems.Freeze(); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if _, err := g.world.Bus.Drain(g.world); err != nil {
		return nil, fmt.Errorf("game: setup events: %w", err)
	}
	if _, err := timeline.Run(g.world, g.clock); err != nil {
		return nil, fmt.Errorf("game: first decision: %w", err)
	}
	if _, err := g.world.Bus.Drain(g.world); err != nil {
		return nil, fmt.Errorf("game: setup events: %w", err)
	}
	g.events = nil
	return g, nil
}

// Apply runs one command, plays time forward until a player-controlled body decides, and returns
// the events in delivery order. An invalid command is not an error: it comes back as an
// api.Rejected event. An unknown command kind is an error and changes nothing. Any other error,
// from a handler or a system, is returned and leaves the game unusable: every later call fails.
func (g *Game) Apply(cmd api.Command) ([]api.Event, error) {
	if g.broken != nil {
		return nil, g.broken
	}
	if !slices.Contains(g.world.Commands.Kinds(), cmd.CommandKind()) {
		return nil, fmt.Errorf("game: unknown command kind %q", cmd.CommandKind())
	}
	g.events = nil
	// Handlers that take time occupy channels themselves, so the returned cost is ignored.
	if _, err := g.world.Commands.Dispatch(g.world, cmd); err != nil {
		return nil, g.fail(err)
	}
	if _, err := timeline.Run(g.world, g.clock); err != nil {
		return nil, g.fail(err)
	}
	if _, err := g.world.Bus.Drain(g.world); err != nil {
		return nil, g.fail(err)
	}
	g.history = append(g.history, cmd)
	events := g.events
	g.events = nil
	return events, nil
}

func (g *Game) fail(err error) error {
	g.broken = fmt.Errorf("game: unusable after error: %w", err)
	return g.broken
}

// Current returns the UID of the body deciding now, or 0.
func (g *Game) Current() api.UID {
	if body, ok := timeline.Current(g.world); ok {
		return g.world.UIDOf(body)
	}
	return 0
}

// View builds the view of the current state.
func (g *Game) View() (api.View, error) {
	if g.broken != nil {
		return api.View{}, g.broken
	}
	return g.world.Views.Build(g.world)
}

// Hash is the SHA-256 of the view as JSON. M1-12 replaces its input with the save bytes.
func (g *Game) Hash() ([32]byte, error) {
	view, err := g.View()
	if err != nil {
		return [32]byte{}, err
	}
	data, err := json.Marshal(view)
	if err != nil {
		return [32]byte{}, fmt.Errorf("game: hash: %w", err)
	}
	return sha256.Sum256(data), nil
}

// History returns every command applied so far, in order.
func (g *Game) History() []api.Command { return slices.Clone(g.history) }

// Codec returns the codec that knows every command and event kind of this game.
func (g *Game) Codec() *api.Codec { return g.world.Codec }

// Replay starts a game from cfg and applies cmds in order.
func Replay(cfg Config, cmds []api.Command) (*Game, error) {
	g, err := New(cfg)
	if err != nil {
		return nil, err
	}
	for i, cmd := range cmds {
		if _, err := g.Apply(cmd); err != nil {
			return nil, fmt.Errorf("game: replay command %d: %w", i, err)
		}
	}
	return g, nil
}
