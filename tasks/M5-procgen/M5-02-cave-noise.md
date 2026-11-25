---
id: M5-02
title: Cave noise stage
status: todo
milestone: M5
kind: feature
depends_on: [M5-01]
discovered_by: null
size: M
---

# M5-02 · Cave noise stage

## Goal

A `caves` stage carves the solid fill with 3D noise into wide chambers, long winding tunnels and thin crawl soops, so the deep winds over and under itself in every direction, the way Minecraft's caves do.

## Context

- `docs/game-design.md` → "The hex grid" (human scale: a hex and a depth step are 1.5 m; low passages are the norm; depth is continuous), "Procgen and exports".
- `docs/go-technical-design.md` → "Procgen stages"; Library choices (opensimplex-go).
- M5-01's `Stage`, `Context` and registry; the brief's "Planned tasks" → 3D noise caves.

## Deliverables

- `go.mod`: `github.com/ojrac/opensimplex-go` (named in the tech design; Notes say why).
- `internal/core/procgen/caves.go`: the `caves` stage and its registration.
- `content/data/geography.cue`: `geography.deep` lists `caves` with its numbers.
- `internal/core/procgen/caves_test.go`: a golden of one chunk for a fixed seed, and a property that the result does not depend on chunk visiting order.

## Spec

- **Three carvers, summed by union** (suggestion, from Minecraft): *cheese* thresholds one low-frequency 3D noise into big chambers; *spaghetti* carves where two noises are both near zero, giving long tunnels that wander in every direction; *noodle* does the same at a higher frequency and a tighter threshold for thin crawl soops. Chamber sizes are tens of hexes across, tunnels a few hexes wide, noodles one hex.
- **Noise lives in world space** (hex centers converted to a plane position plus depth), so a cave crosses chunk borders without a seam. Seeded by the world seed and export count; cells never depend on neighbor chunk data.
- **Numbers are content** (frequencies, thresholds, a depth falloff so the surface layer stays solid and the deepest layer can be rarer). Params are named so a tuner can read them. Firm.
- Carved cells become the world's `Open` terrain; the rest stays fill. A carved cell never lies outside the bounds.
- No guarantee of connectivity here: joining is M5-05, and the guarantees are M5-07.

## Acceptance criteria

- [ ] Golden and order-independence tests pass.
- [ ] `play --seed 1` shows tunnels and chambers in the plan view away from the places, shaded by depth.
- [ ] `go tool task check` passes.

## Out of scope

- Sinks: M5-03. Connecting caves to places: M5-05. Regions and materials: M5-04.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
