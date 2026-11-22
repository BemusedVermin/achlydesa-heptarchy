---
paths:
  - "internal/frontend/**"
---

# Working on the frontend

The frontend draws the core's view and sends the commands the view offers. Nothing more (AGENTS.md rules 1, 6 and 7).

- Read game state only through `game.Game`: `View()` gives the `api.View` snapshot, and `Apply` sends commands. From the core, import only `game` and `api`, and never reach into the ECS.
- Send only the commands the view offers. Each `api.Option` has a `Label` to show and a `Command` to send when it's chosen. Don't build commands, work out valid moves, or reorder menus here. If a screen needs a choice the view doesn't offer, the core has to serve it: file a follow-up.
- Strings in the view arrive already rendered, so game text never lives in Go here. Key bindings and their help text are the frontend's own.
- Keep drawing in pure functions that take view types and return strings (`RenderMap`, `RenderTimeline`, ...), and cover them with golden tests under `testdata/`. Regenerate the golden files with `go test ./internal/frontend/tui -update`, and read the diff before committing it.
- Use Lip Gloss for layout and Bubbles for widgets, with one model per screen on the screen stack.
- A graphics frontend will later draw the same view and send the same commands. Anything it would need too belongs in the core, not here.
