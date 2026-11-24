---
id: M3-02
title: "Needs: records, levels, a per-phase tick and the satisfy effect"
status: done
milestone: M3
kind: seam
depends_on: []
discovered_by: null
size: M
---

# M3-02 · Needs: records, levels, a per-phase tick and the satisfy effect

## Goal

Any body can carry needs from its template. Each need's level rises every day phase by a content formula, effects can lower or raise it, and content-defined thresholds apply effects while a need goes unmet. This is the first code in `internal/core/agents`.

## Context

- The agents model in [README.md](README.md), "Needs".
- `docs/game-design.md` → "The living deep" → "Monsters with needs", "People with needs", and "Party" → "Food and water".
- `docs/go-technical-design.md` → "Simulation loop and determinism" → "Tick order" (step 4, the per-phase systems) and the `comp` row (`Needs`).
- `internal/core/sim/scheduler.go` (`PerPhase`), `internal/core/effects/effects.go` (`Register`, `Apply`), `internal/content/env.go` (`BodyEnv`).

## Deliverables

- `internal/content/types.go`: `NeedDef`, and `needs` on `Template`; registry accessors; schemas regenerated.
- `internal/content/env.go`: `BodyEnv.Needs`.
- `internal/core/comp/comp.go`: `Needs`, registered and saved.
- `internal/core/agents/doc.go`, `install.go`, `needs.go`: the per-phase system and the `satisfy` primitive. `internal/core/game/install.go` installs `agents`.
- `content/data/needs.cue` (new), with `need.water` and `need.food` for tests to use; `content/data/text/agents.cue` for their names.
- `internal/core/agents/needs_test.go`.
- `internal/core/save/save.go`: bump `CurrentVersion`.

## Spec

**Firm:**

```go
// content
type NeedDef struct {
    Name api.ContentID       // text ID
    Rise string              // Expr float over NeedEnv: how much the level rises per phase
    Urge string              // Expr float over NeedEnv: 0 to 1, how pressing the need is at this level
    Thresholds []Threshold   // in ascending At
}
type Threshold struct {
    At float64               // applies while level >= At
    Do []EffectRef           // applied to the body once per phase while it holds
}

// comp
type NeedLevel struct { Need api.ContentID; Level float64 }
type Needs struct { Levels []NeedLevel } // sorted by Need, none twice
```

- A template's `needs` maps need IDs to starting levels (`needs: {"need.water": 0.2}`).
- The `needs` system runs `PerPhase` in UID order over every body with `Needs` that is not dead: it adds `Rise`, clamps to 0..1, then applies each threshold that holds through `effects.Apply` with the body as source and target.
- `satisfy` takes `need` and `amount` (an Expr; negative raises the level). It lowers the target's level, clamped to 0..1, and does nothing to a body without that need.
- `BodyEnv.Needs` is a map from need key (the ID after `need.`) to level, so a formula reads `source.needs.water`.
- Content errors (a threshold naming an unknown effect, a template naming an unknown need) fail at load.

**Suggestion:** `NeedEnv` holds `level`, `phase` (the phase ID) and `source` (`BodyEnv`).

## Acceptance criteria

- [x] A test body with `need.water` at 0.2 and a rise of 0.1 reads 0.3 after one phase boundary and 0.4 after two.
- [x] A content action whose `on_land` is `satisfy` lowers the level, and the level never leaves 0..1.
- [x] A threshold at 0.5 applies its effect each phase while the level is at or above 0.5, and stops once `satisfy` drops it below.
- [x] A save taken with needs set loads back with the same levels and the same `Game.Hash`.
- [x] `go tool task check` passes.

## Out of scope

- Choosing behaviors by need (M3-03).
- The party's food and water and their thresholds (M3-13).
- Off-screen aggregates' needs (M3-10).

## Notes

- **Deviation:** the need names sit beside the records in `content/data/needs.cue`, not in `content/data/text/agents.cue`, because `.claude/rules/content.md` puts a record's name beside the record and keeps `text/` for engine messages. `agents` has no engine messages yet, so `text/agents.cue` doesn't exist.
- **Decisions where the spec is silent (all reversible):**
  - A body's needs tick in ID order, each one reading a fresh `BodyEnv`, so a threshold on one need can affect the rise of the next.
  - Threshold `at` must be from 0 to 1 and strictly ascending, and starting levels must be from 0 to 1. Both fail at load.
  - `satisfy` acts on the effect's target only, and publishes no event. `need` is a text param: a literal need ID is checked when a game starts.
  - The sample `need.water` and `need.food` have placeholder rises (0.05 and 0.02 per phase), `urge: "level"`, no thresholds, and are named "water" and "food". No sample template carries a need, so sample games behave as before. The party's needs belong to M3-13.
  - Bodies with `comp.Dead` are skipped. Downed bodies still tick.
- **Files outside Deliverables:** `internal/content/needs.go` (loader checks and accessors), `actions.go` (threshold effects are effect sites, so `effects.ValidateContent` checks them when a game starts), `records.go`, `registry.go`, `load.go`; `internal/core/effects/env.go` (fills `BodyEnv.Needs`); `internal/core/save/save_test.go` (the migration test builds its no-op chain up to `CurrentVersion` instead of hard-coding two steps); `docs/content-guide.md` (needs in the layout and ID tables, a Needs section, `satisfy` in the primitive list).
- **Follow-ups:** M3-02.F1 (a need dropped by hot-reloaded content stops the game at the next phase boundary).

