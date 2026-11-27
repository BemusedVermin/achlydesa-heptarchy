package harness

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// Replay plays the command log read from r on a new game and returns the final hash as hex. The
// codec comes from a fresh game, which knows every command kind.
func Replay(reg *content.Registry, seed uint64, r io.Reader) (string, error) {
	return ReplayTrace(reg, seed, r, nil)
}

// ReplayTrace is Replay that also writes to trace, when it is not nil, each command's index, its
// kind and the message log lines it added.
func ReplayTrace(reg *content.Registry, seed uint64, r io.Reader, trace io.Writer) (string, error) {
	g, err := game.New(game.Config{Seed: seed, Content: reg})
	if err != nil {
		return "", fmt.Errorf("harness: %w", err)
	}
	cmds, err := g.Codec().ReadLog(r)
	if err != nil {
		return "", fmt.Errorf("harness: %w", err)
	}
	var seen int64
	for i, cmd := range cmds {
		if _, err := g.Apply(cmd); err != nil {
			return "", fmt.Errorf("harness: command %d (%s): %w", i, cmd.CommandKind(), err)
		}
		if trace == nil {
			continue
		}
		view, err := g.View()
		if err != nil {
			return "", fmt.Errorf("harness: %w", err)
		}
		if _, err := fmt.Fprintf(trace, "#%d %s\n", i, cmd.CommandKind()); err != nil {
			return "", fmt.Errorf("harness: %w", err)
		}
		for _, line := range view.Log {
			if line.Seq <= seen {
				continue
			}
			seen = line.Seq
			if _, err := fmt.Fprintf(trace, "  t%d %s\n", line.Tick, line.Text); err != nil {
				return "", fmt.Errorf("harness: %w", err)
			}
		}
	}
	hash, err := g.Hash()
	if err != nil {
		return "", fmt.Errorf("harness: %w", err)
	}
	return hex.EncodeToString(hash[:]), nil
}
