// Package harness plays seeded games headlessly and writes a CSV row per step, plays scripted fights
// to measure them, and replays recorded command logs to check their hashes. Each game is driven by a
// pluggable [Policy]. It talks only to core/game.
package harness
