package game

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/combat"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Config says how to start a game.
type Config struct {
	// Seed seeds every random stream. Load ignores it and takes the save's.
	Seed uint64
	// Content is required.
	Content *content.Registry
	// Install, when set, runs after the standard packages are installed and before any map is built
	// or any body spawned, so the start map's bodies may name the controllers and choosers it
	// registers. Tests and the harness use it to register hooks. It runs in New and in Load, so it
	// must not spawn entities.
	Install func(w *sim.World, clk *clock.Clock) error
	// Spawn, when set, runs after the start map's bodies are spawned and adds the entities a
	// scenario needs, such as a test actor. New runs it; Load skips it, because the save holds the
	// entities.
	Spawn func(w *sim.World, clk *clock.Clock) error
}

// Game is one running game. It is not safe for concurrent use.
type Game struct {
	world   *sim.World
	clock   *clock.Clock
	history []api.Command
	// events are those delivered since the last Apply returned.
	events []api.Event
	log    []api.LogLine
	broken error
}

// New builds a game from cfg, installs every package, and plays the timeline forward to the first
// player decision.
func New(cfg Config) (*Game, error) {
	g, err := build(cfg, func(w *sim.World) error {
		if err := explore.Start(w); err != nil {
			return fmt.Errorf("game: start: %w", err)
		}
		return nil
	}, cfg.Spawn)
	if err != nil {
		return nil, err
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

// build is the path New and Load share: an empty world with every package installed, the content
// checked and the scheduler frozen. scenario builds the maps and their first entities, after
// cfg.Install.
func build(cfg Config, scenario func(w *sim.World) error, spawn func(*sim.World, *clock.Clock) error) (*Game, error) {
	if cfg.Content == nil {
		return nil, errors.New("game: no content")
	}
	g := &Game{world: sim.NewWorld(cfg.Seed), clock: clock.New(cfg.Content.Calendar())}
	g.world.Content = cfg.Content
	if err := g.installLog(); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if err := installLabels(g.world); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if err := installAll(g.world, g.clock); err != nil {
		return nil, fmt.Errorf("game: install: %w", err)
	}
	if cfg.Install != nil {
		if err := cfg.Install(g.world, g.clock); err != nil {
			return nil, fmt.Errorf("game: install hook: %w", err)
		}
	}
	if err := scenario(g.world); err != nil {
		return nil, err
	}
	if spawn != nil {
		if err := spawn(g.world, g.clock); err != nil {
			return nil, fmt.Errorf("game: spawn: %w", err)
		}
	}
	if err := validateContent(g.world); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if err := g.world.Systems.Freeze(); err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	return g, nil
}

// validateContent runs the content checks that need the packages installed: effect primitives and
// references, and action choice providers.
func validateContent(w *sim.World) error {
	if err := effects.ValidateContent(w); err != nil {
		return err
	}
	return actions.ValidateContent(w)
}

// SwapContent replaces the game's content with reg, for hot reload in dev builds. It runs the
// content checks build does against the new registry first, and keeps the old content when they
// fail. Call it only between commands, from the goroutine that owns the game.
//
// Entities keep the components they were spawned with, because components hold content IDs and
// values, not pointers into the registry; only later spawns and lookups use the new templates and
// numbers. The swap is not recorded in [Game.History], so a replay that began before it will not
// reproduce the game, and a warning line in the message log says so.
func (g *Game) SwapContent(reg *content.Registry) error {
	if reg == nil {
		return errors.New("game: swap content: no content")
	}
	old := g.world.Content
	g.world.Content = reg
	if err := validateContent(g.world); err != nil {
		g.world.Content = old
		return fmt.Errorf("game: swap content: %w", err)
	}
	g.clock.Reset(reg.Calendar())
	timeline.Reload(g.world)
	g.NoteText(api.Text{ID: "text.game.content_reloaded"})
	return nil
}

// ReloadFailed logs that a content reload failed with err and the old content stays.
func (g *Game) ReloadFailed(err error) {
	g.NoteText(api.Text{ID: "text.game.reload_failed", Args: map[string]string{"error": err.Error()}})
}

// Note adds a line of text straight to the message log, for the frontend to report something that
// is not a game event, such as a failed reload.
func (g *Game) Note(text string) { g.appendLog(text) }

// NoteText is [Game.Note] for text the game's content renders.
func (g *Game) NoteText(text api.Text) { g.appendLog(g.world.Content.Render(text)) }

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
	if combat.PartyWiped(g.world) {
		g.world.Bus.Publish(api.Rejected{Command: cmd.CommandKind(), Reason: api.Text{ID: "text.game.lost"}})
		if _, err := g.world.Bus.Drain(g.world); err != nil {
			return nil, g.fail(err)
		}
		return g.record(cmd), nil
	}
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
	if err := g.wakeIfWiped(); err != nil {
		return nil, g.fail(err)
	}
	return g.record(cmd), nil
}

// record adds cmd to the history and hands over the events delivered since the last Apply.
func (g *Game) record(cmd api.Command) []api.Event {
	g.history = append(g.history, cmd)
	events := g.events
	g.events = nil
	return events
}

// wakeIfWiped gets a wiped-out party up at its last rest place after the configured time, then
// plays on to the first decision. A party with nobody downed stays lost: nothing can wake it.
func (g *Game) wakeIfWiped() error {
	if !combat.PartyWiped(g.world) || !explore.CanWake(g.world) {
		return nil
	}
	if err := g.clock.Advance(g.world, max(g.world.Content.Party().WakeTicks, 0)); err != nil {
		return err
	}
	if err := explore.Wake(g.world); err != nil {
		return err
	}
	if _, err := timeline.Run(g.world, g.clock); err != nil {
		return err
	}
	_, err := g.world.Bus.Drain(g.world)
	return err
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

// View builds a fresh snapshot for a frontend. It fails once an error has left the game unusable.
func (g *Game) View() (api.View, error) {
	if g.broken != nil {
		return api.View{}, g.broken
	}
	return g.world.Views.Build(g.world)
}

// Hash is the SHA-256 of the saved bytes, so it covers everything a save keeps.
func (g *Game) Hash() ([32]byte, error) {
	var buf bytes.Buffer
	if err := g.Save(&buf); err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(buf.Bytes()), nil
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
