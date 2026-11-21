---
id: M1-12
title: "save: versioned snapshots and migrations"
status: todo
milestone: M1
kind: seam
depends_on: [M1-16]
discovered_by: null
size: M
---

# M1-12 · save: versioned snapshots and migrations

## Goal

A save captures the whole game in one versioned file, loads into a game that continues exactly as if it never stopped, and older versions upgrade through a migration chain. The world hash becomes the hash of the save bytes, so every later system is covered by the determinism tests for free.

## Context

- Tech design: "Simulation loop and determinism → Saves and replays", "Designing for pivots" (Saves, Identity).
- ark-serde: https://github.com/mlange-42/ark-serde (`go doc github.com/mlange-42/ark-serde`). Check what it requires of the target world before deserializing, such as registered component types.

## Deliverables

- `internal/core/save/save.go`, `migrate.go`, `save_test.go`
- `internal/core/game/game.go`: `Save`, `Load`, and `Hash` switched to save bytes.
- `go.mod`: adds `github.com/mlange-42/ark-serde`.

## Spec

**File:** one JSON document.

```json
{
  "header": { "format": "heptarchy-save", "version": 1, "content_hash": "...", "seed": 1, "tick": 0 },
  "world": { "...ark-serde output..." },
  "rng": { "stream name": "base64 PCG state" },
  "next_uid": 2,
  "map_edits": { "1": [ { "at": { "hex": { "q": 0, "r": 0 }, "depth": 1 }, "terrain": "terrain.air" } ] },
  "history": [ { "kind": "explore.step", "data": { } } ]
}
```

Maps from content aren't saved; they're rebuilt from content on load, and then each map's edits (`hexmap.Map.Edits()`, such as dug tunnels later) are reapplied. Generated maps arrive in M5 and will either be saved or regenerated from their seed and export count; file a question when that milestone is broken down.

**save** (suggestions):

```go
const CurrentVersion = 1
type Header struct { /* fields above */ }
type Snapshot struct {
	Header  Header
	World   json.RawMessage
	RNG     map[string][]byte
	NextUID api.UID
	MapEdits map[api.MapID][]hexmap.Edit
	History []json.RawMessage
}
func Write(w io.Writer, s Snapshot) error
func Read(r io.Reader) (Snapshot, []string, error) // warnings, e.g. a content hash mismatch

// Migration upgrades a raw document from version N to N+1.
type Migration func(doc map[string]json.RawMessage) error
// migrations[i] upgrades version i+1 to i+2. Empty for now.
var migrations []Migration
```

- `Read` rejects other formats and versions newer than `CurrentVersion`, and applies migrations from the file's version up.
- A content hash mismatch is a warning, not an error, because content changes during development.

**game:**

```go
func (g *Game) Save(w io.Writer) error
func Load(cfg Config, r io.Reader) (*Game, []string, error)
```

- `Load` builds a game through the same `Install` path as `New` but skips the start scenario. It then restores the ECS world, the RNG streams, `Now`, the UID counter, the map edits and the history.
- Split `New` into an unexported base builder plus the start scenario so the two share one path.
- `Hash` becomes the SHA-256 of the saved bytes. If ark-serde's output isn't byte-stable for identical worlds, canonicalize it first, and say how in Notes.

## Acceptance criteria

- [ ] Save at tick 0 then load: identical hash.
- [ ] A map edit made in a test survives save and load.
- [ ] A `rapid` property: apply N random commands, save, load, apply M more to both the loaded and the original game; the hashes match.
- [ ] A save whose version is newer than `CurrentVersion` fails with a clear error; a fake version-0-to-1 migration in a test runs when loading a version-0 document.
- [ ] Changing content and loading an old save returns a content-hash warning and still loads.
- [ ] The replay tests from M1-09 and M1-10 still pass with the new `Hash`.
- [ ] `go tool task check` passes.

## Out of scope

- Save slots, autosave and a load screen in the frontend.
- Compression.

## Notes
