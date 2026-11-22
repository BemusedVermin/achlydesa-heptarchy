package harness

import (
	"fmt"
	"io"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// Replay plays the command log read from r on a new game and returns the final hash as hex. The
// codec comes from a fresh game, which knows every command kind.
func Replay(reg *content.Registry, seed uint64, r io.Reader) (string, error) {
	cfg := game.Config{Seed: seed, Content: reg}
	fresh, err := game.New(cfg)
	if err != nil {
		return "", fmt.Errorf("harness: %w", err)
	}
	var cmds []api.Command
	if cmds, err = fresh.Codec().ReadLog(r); err != nil {
		return "", fmt.Errorf("harness: %w", err)
	}
	g, err := game.Replay(cfg, cmds)
	if err != nil {
		return "", fmt.Errorf("harness: %w", err)
	}
	hash, err := g.Hash()
	if err != nil {
		return "", fmt.Errorf("harness: %w", err)
	}
	return fmt.Sprintf("%x", hash), nil
}
