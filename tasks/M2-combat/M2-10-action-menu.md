---
id: M2-10
title: Action menu and look mode on the map screen
status: todo
milestone: M2
kind: feature
depends_on: [M2-08]
discovered_by: null
size: M
---

# M2-10 · Action menu and look mode on the map screen

## Goal

Fighting happens on the map screen. A panel shows the current member's body, a drill-down menu built from the core's options picks the move, targets light up on the map, and a look mode inspects any visible body's parts. There's no separate combat screen.

## Context

- Tech design: "Frontends → The terminal frontend".
- Game design: "Combat → Fighting on the map".
- M1-11 the map screen and its panes; M2-08 `BodiesView` and the `act/` option IDs.

## Deliverables

- `internal/frontend/tui/actorpanel.go`, `menu.go`, `look.go`, tests with goldens.
- `internal/frontend/tui/mapscreen.go`: the panels and the target highlight.
- `internal/frontend/tui/keys.go`: the new bindings.

## Spec

- **Alert:** while the party is on alert, the status line says so, and the turn-order panel marks hostile bodies.
- **Actor panel:** when the current actor is a member, the right pane shows its body. That's one line per part (slot, name, material, an integrity bar, "pinned" when locked, "graft" for grafts) and the steps and actions it has left.
- **Action menu:** `enter` opens a drill-down built by splitting `act/` option IDs on `/`. The levels are form, then type, then target, then part, and a level with a single entry is skipped. Each level is a `bubbles/list`. `esc` goes up a level, `enter` chooses, and choosing a leaf sends that option's command. Step keys still work while the menu is closed.
- **Target highlight:** while a target or part is being chosen, the target's hex is highlighted on the map, found through its `EntityView` by UID.
- **Look mode:** `l` puts a cursor on the map, moved with the step keys. The panel shows the `BodiesView` entry under the cursor, with materials or "?". `esc` leaves.
- **No game logic:** the frontend only groups option IDs and reads the view. It never checks legality, reach, costs or targets itself.
- Render functions are pure and golden-tested, like `RenderMap`.

## Acceptance criteria

- [ ] In `heptarchy play`, walking into sight of the hall's creature shows the alert. Members then step and attack through the menu without leaving the map, and lead and follow resume once it's dead and out of mind.
- [ ] The menu drills form → type → target → part, skips single-choice levels, and backs out with `esc`.
- [ ] Look mode shows an enemy's parts with "?" before a Sweep and materials after.
- [ ] Golden tests: the actor panel at the start of a turn, after a Sweep, after one of the actor's arms is severed, each menu level, and look mode on an enemy.
- [ ] `go tool task check` passes.

## Out of scope

- Animations, mouse support, color themes.

## Notes
