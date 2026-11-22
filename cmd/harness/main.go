// Command harness plays many seeded games headlessly and replays recorded logs.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

const defaultContent = "content/data"

func main() {
	if err := root().Execute(); err != nil {
		os.Exit(1)
	}
}

func root() *cobra.Command {
	cmd := &cobra.Command{Use: "harness", Short: "Headless seeded runs and replays", SilenceUsage: true}
	cmd.AddCommand(runCommand(), replayCommand())
	return cmd
}

func runCommand() *cobra.Command {
	var seeds, steps int
	var seedStart uint64
	var dir, out string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Play seeded games with a random policy and write CSV",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			reg, err := content.Load(dir)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if out != "" {
				f, err := os.Create(out)
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				w = f
			}
			logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), nil))
			return harness.Run(harness.RunConfig{
				Seeds: seeds, SeedStart: seedStart, Steps: steps, Content: reg, Logger: logger,
			}, w)
		},
	}
	cmd.Flags().IntVar(&seeds, "seeds", 1, "number of seeds to play")
	cmd.Flags().Uint64Var(&seedStart, "seed-start", 1, "first seed")
	cmd.Flags().IntVar(&steps, "steps", 100, "steps per seed")
	cmd.Flags().StringVar(&dir, "content", defaultContent, "content directory")
	cmd.Flags().StringVar(&out, "out", "", "CSV file to write (default stdout)")
	return cmd
}

func replayCommand() *cobra.Command {
	var seed uint64
	var dir, expect string
	cmd := &cobra.Command{
		Use:   "replay FILE",
		Short: "Replay a command log and print the final hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := content.Load(dir)
			if err != nil {
				return err
			}
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			hash, err := harness.Replay(reg, seed, f)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), hash)
			if expect != "" && expect != hash {
				return fmt.Errorf("hash %s does not match expected %s", hash, expect)
			}
			return nil
		},
	}
	cmd.Flags().Uint64Var(&seed, "seed", 1, "the seed the log was recorded with")
	cmd.Flags().StringVar(&dir, "content", defaultContent, "content directory")
	cmd.Flags().StringVar(&expect, "expect", "", "expected final hash in hex; exit 1 on mismatch")
	return cmd
}
