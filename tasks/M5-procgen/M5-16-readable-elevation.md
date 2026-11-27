---
id: M5-16
title: Make elevation readable on the terminal map
status: done
milestone: M5
kind: feature
depends_on: []
discovered_by: null
size: M
---

# M5-16 · Make elevation readable on the terminal map

## Goal

On the terminal map, a player can tell at a glance how high each cell around them is, which way they can step, climb or drop, and why a step was refused. Today the only cue is a faint shade, and every refused step says "Packed clay. No way through.", even when the cell is open but too high to climb.

## Context

- Filed by the owner on 2026-11-26 after playtesting the generated deep: they couldn't tell when the character was blocked, what the elevation was, or when a wall was really a ledge.
- `docs/game-design.md` → "The hex grid": the plan view is one continuous map, "shaded by how far above or below the party it lies. No layers are ever visible." Depth steps are 1.5 m, and caves "slope gently and drop off ledges".
- `api.CellView.Below` and `api.EntityView.Below` (`internal/core/api/explore.go`) already carry each cell's depth relative to the party. `shade` in `internal/frontend/tui/render.go` darkens 8% per step down and lightens 6% per step up, capped at `maxShade`. Next to the remembered-cell dimming, that is hard to read.
- Step rules: `hexmap.StepRules` (`MaxStepUp`, `MaxStepDown`, `MaxDrop`) and `Map.Step` (`internal/core/hexmap/step.go`), set by `max_step_up` and friends in `content/data/party.cue`. `checkStep` in `internal/core/explore/install.go` returns the same `text.explore.bump` for every kind of refusal.

## How other games do it

Recalled from knowledge of the games, not checked this session:

- **Dwarf Fortress (ASCII):** draws one z-level at a time, with ramps as ▲/▼ and stairs as `<` `>` `X`. Open air shows the level below in darker or bluish colors, so "lower" reads as a hue, not just a brightness.
- **Cataclysm: DDA:** z-levels with `<` `>` for stairs and ledges. An optional 3D field of view draws lower levels faded, and toggleable map overlays put a value on each tile.
- **Tactics games (Final Fantasy Tactics, XCOM, Into the Breach):** the inspected tile shows its height as a number. Move previews highlight only reachable tiles and mark climbs, and XCOM marks an unreachable target and says why.
- **Brogue, DCSS:** no elevation, but a hover path preview shows the route before you commit. That's the same idea: the game shows what a move will do before you make it.

What this game can take from them, keeping its one continuous map:

1. A **hue** for relative height (cool below, warm above), not just brightness, plus a numeric overlay for exact values.
2. A **move readout** that comes from the core: for each of the six directions, whether the step is level, up, down, a drop of n, or refused, and why.
3. **Specific refusals** in the log.

## Deliverables

- `internal/core/api/explore.go`: a per-direction move readout for the acting body (suggestion: `MapView.Moves []MoveView`), with regenerated schemas.
- `internal/core/explore/`: build the readout from `Map.Step` and the rules. `checkStep` returns a distinct reason for each kind of refusal.
- `content/data/text/explore.cue`: the readout's labels and the new refusal texts.
- `internal/frontend/tui/render.go`, `mapscreen.go`, `keys.go`: hue shading, the elevation overlay toggle, and the move readout in the side panel.
- `internal/frontend/tui/` golden render tests under `testdata/`: the overlay on, and the readout.
- `docs/game-design.md` → "The hex grid": describe the hue and the overlay.

## Spec

**Move readout (core; the shape is a suggestion).** For the acting body, one entry per direction in `api.Dir` order: the direction, a kind (`level`, `up`, `down`, `drop`, `blocked`), the relative height of the cell it would end in, and a rendered label. Examples: "E: climb 1 (+5 ticks)", "NE: ledge 2 up, too high", "W: drop 3", "SW: wall". The core decides what is legal (rule 7), and the frontend only draws it.

**Refusals.** `checkStep` tells apart at least: a solid cell (today's bump text), a ledge too high to climb (with its height and the limit), a drop past `MaxDrop`, and an occupied cell (already `text.explore.occupied`). All the wording is content.

**Map shading (TUI).** Keep the continuous map. Encode relative height as a hue shift (suggestion: toward blue below, toward amber above) as well as the existing brightness, so a one-step difference shows and the remembered-cell dimming doesn't drown it. Keep `maxShade` as the cap.

**Elevation overlay (TUI).** One key toggles an overlay that draws each visible open cell as its relative height (`0`, `1` to `9` below, and `-1`… shown as a raised glyph or `+n`, the exact mark is the implementer's call), with solid cells unchanged. This is a view of `Below`, which the view already sends. No new core data is needed.

**Party depth.** The side panel shows the acting body's depth, as a rendered line from the core.

## Acceptance criteria

- [x] The view's move readout lists the six directions with the right kind for a level step, a climb, a step down, a drop and a refused ledge (an explore test on a small fixture map).
- [x] Stepping into a ledge too high to climb logs the ledge refusal, not `text.explore.bump`. Stepping into solid ground still logs the bump.
- [x] With the overlay on, a golden render shows the relative heights of visible open cells. With it off, the render shows the hue shading.
- [x] The side panel shows the move readout and the depth line (a golden render).
- [x] `docs/game-design.md` → "The hex grid" describes the hue and the overlay.
- [x] `go tool task schema` is run and its output committed.
- [x] `go tool task check` passes.

## Out of scope

- Sprites, tiles and a Godot or Ebitengine frontend: later.
- A path preview over several steps: a possible follow-up.
- Showing layers other than the continuous map ("No layers are ever visible").

## Notes

- Core: `hexmap.Map.Refusal` explains a refused step (solid, ledge one past the climb limit, drop past `MaxDrop`); `checkStep` uses it for `text.explore.too_high` and `text.explore.too_far`. A ledge more than one cell past the limit still reads as a wall.
- `api.MapView` gains `Moves`, `Elevation` (named so, since a test forbids "depth" keys) and `CellView.Wall`, which the overlay needs to leave solid cells unchanged.
- The readout labels carry no tick cost ("climb 1", not "+5 ticks"): the cost depends on the body's speed, and the label stays simple. Say if you want it.
- The overlay key is `h`. Above the party draws `+n` using the spacer column before the hex. The move readout hides while the action menu is open, so the menu keeps its room.
- Hue shading: blue and darker below, amber and lighter above; `maxShade` still caps it.
- Tests: hexmap refusal table, explore readout on the descent place, goldens for the overlay and the readout. No explore-level test of the ledge log line (no fixture has a 2-high ledge); the hexmap test covers the classification.
- `docs/game-design.md` "The hex grid" updated.
- Follow-ups: none.
