---
id: M1-11
title: Terminal frontend on the new API, with a CLI
status: todo
milestone: M1
kind: feature
depends_on: [M1-16]
discovered_by: null
size: M
---

# M1-11 · Terminal frontend on the new API, with a CLI

## Goal

The Bubble Tea frontend draws only the core's view and sends only the commands the view offers. Exploring and fighting share one map screen; a screen stack exists only for overlays such as the journal and dreams later. The prototype is deleted.

## Context

- Tech design: "Frontends → The terminal frontend (now)", "Designing for pivots" (Choices come from the core, Text and looks).
- `cmd/heptarchy/main.go`: the prototype's double-width rendering to port.
- Bubble Tea v2 (`charm.land/bubbletea/v2`), Bubbles v2 (`charm.land/bubbles/v2`), Lip Gloss v2 (`charm.land/lipgloss/v2`), cobra (`github.com/spf13/cobra`).

## Deliverables

- `internal/frontend/tui/app.go`: the root model and screen stack.
- `internal/frontend/tui/mapscreen.go`: the map screen.
- `internal/frontend/tui/render.go`: pure rendering functions.
- `internal/frontend/tui/keys.go`: the keymap.
- `internal/frontend/tui/render_test.go`, `testdata/*.golden`
- `cmd/heptarchy/main.go`: rewritten with cobra.
- Delete `internal/prototype/`.

## Spec

**CLI:** `heptarchy play [--seed N] [--content DIR] [--record FILE]`. Defaults: seed 1, content `content/data`, no recording. `play` is also what a bare `heptarchy` runs. `--record` writes the command history as a replay log (`Codec.WriteLog`) when the game exits. Keep the existing `DEBUG` log-to-file behavior.

**Root model:**

```go
// Screen is one full-screen mode. The root model owns a stack of them.
type Screen interface {
	Update(msg tea.Msg, app *App) (Screen, tea.Cmd) // return nil to pop
	View(app *App) string
}
```

`App` holds the `*game.Game`, the latest `api.View`, the window size and the stack. After any command it re-reads the view. A command that returns an error shows the error and quits; it is a bug, not player input.

**Choosing options:** keys bind to option IDs, never to commands. The keymap (`bubbles/key`):

| Keys | Option ID |
| --- | --- |
| `q`, `e`, `a`, `d`, `z`, `c` | `step.nw`, `step.ne`, `step.w`, `step.e`, `step.sw`, `step.se` |
| `.` | `wait` |
| `space` | `end` |
| `tab` | the next `lead/<uid>` option, cycling |
| `esc`, `ctrl+c` | quit |

When a key's option isn't in the view, nothing happens; when it's disabled, the status line shows its reason. A help line (`bubbles/help`) lists the bindings whose options are present.

**Map screen layout** (Lip Gloss): the map pane on the left. On the right: the clock (day, phase, a bar for ticks in the phase), the party (members in order, the leader marked, the current actor highlighted), the turn order (`TurnsView.Upcoming`, by glyph and name), and the last lines of the log. The help line runs along the bottom. Leave room on the right for M2-10's action panel.

**Rendering** (pure functions, golden-tested):

```go
func RenderMap(m *api.MapView, party *api.PartyView, width, height int) string
```

- Double-width layout: column `2*q + r`, row `r`, shifted so the map is centered on the party when it's larger than the pane. Center on `PartyView.Hex`.
- **Relative depth is shading, never a layer.** A visible cell at `Below` 0 draws in its own color. Each cell lower darkens it a step and each cell higher lightens it a step, smoothly, out to ±8. Use truecolor through Lip Gloss and fall back to fewer steps on 256-color terminals. Nothing on screen shows a depth number or a layer.
- Remembered hexes draw dimmed on top of that; entities draw over terrain with the same shading; the party draws over everything.
- Color names come from `Appearance.Color` and terrain colors in content. Map names to Lip Gloss colors in one table, and fall back to no color.

## Acceptance criteria

- [ ] `go run ./cmd/heptarchy play --seed 1` shows the hall, walks the leader with the six keys while the follower trails, bumps with the log line, `.` waits to the next phase, and `tab` changes the leader.
- [ ] `--record run.jsonl` writes a log that `game.Replay` reproduces to the same hash (test via a helper, not the TTY).
- [ ] Golden tests for `RenderMap`: the start view, after walking away (dimmed memory), halfway down the descent's slope (shading changes smoothly), and a pane smaller than the map. Regenerate with `go test ./internal/frontend/tui -update`.
- [ ] Pressing the key for a disabled move shows its reason in the status line and sends nothing.
- [ ] Nothing in `internal/frontend` imports a core package other than `game`, `api` and `content`.
- [ ] `internal/prototype` is gone.
- [ ] `go tool task check` passes.

## Out of scope

- The action menu and look mode (M2-10). Journal and dream overlays (later milestones).
- Terminal font choice (open question in the tech design).

## Notes
