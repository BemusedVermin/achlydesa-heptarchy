// Command heptarchy is the game's terminal frontend.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/frontend/tui"
)

const defaultContent = "content/data"

func main() {
	if err := root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// playOptions are the flags of the play command.
type playOptions struct {
	seed    uint64
	content string
	record  string
}

func (o *playOptions) addFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64Var(&o.seed, "seed", 1, "world seed")
	cmd.Flags().StringVar(&o.content, "content", defaultContent, "content directory")
	cmd.Flags().StringVar(&o.record, "record", "", "write the command history to this replay log when the game exits")
}

// root is the command tree. A bare "heptarchy" runs play.
func root() *cobra.Command {
	opts := &playOptions{}
	cmd := &cobra.Command{
		Use:          "heptarchy",
		Short:        "Achlydesa: The Dreaming Heptarchy",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE:         func(*cobra.Command, []string) error { return play(opts) },
	}
	opts.addFlags(cmd)

	playOpts := &playOptions{}
	playCmd := &cobra.Command{
		Use:   "play",
		Short: "Play in the terminal",
		Args:  cobra.NoArgs,
		RunE:  func(*cobra.Command, []string) error { return play(playOpts) },
	}
	playOpts.addFlags(playCmd)
	cmd.AddCommand(playCmd)
	return cmd
}

func play(opts *playOptions) error {
	if os.Getenv("DEBUG") != "" {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
	}
	reg, err := content.Load(opts.content)
	if err != nil {
		return err
	}
	g, err := game.New(game.Config{Seed: opts.seed, Content: reg})
	if err != nil {
		return err
	}
	app, err := tui.NewApp(g)
	if err != nil {
		return err
	}
	if _, err := tea.NewProgram(app).Run(); err != nil {
		return err
	}
	if opts.record != "" {
		if err := record(g, opts.record); err != nil {
			return err
		}
	}
	return app.Err()
}

func record(g *game.Game, path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	return tui.WriteReplay(g, f)
}
