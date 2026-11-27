---
id: M6-23
title: "Residents: roles, jobs and a town that can run dry"
status: todo
milestone: M6
kind: content
depends_on: [M6-15, M6-16, M6-18, M6-22]
discovered_by: null
size: M
---

# M6-23 · Residents: roles, jobs and a town that can run dry

## Goal

Towns have people with needs. Haulers carry water into the cistern and larder, residents drink and eat from them, and when the haulers die or the source fails, people go looking for what they need.

## Context

- `docs/game-design.md` → "People with needs" (needs, behaviors, what falls out); "Three tiers of people".
- M6-22 (towns), M6-13 (take, put, consume), M6-15 and M6-16 (planners), M6-18 (hearsay).
- Rule 10: no system makes a town empty or announces it.

## Deliverables

- `content/data/templates.cue`: `template.person.resident`, `.hauler`, `.farmer` (a role is a template with a recipe list and drive weights), using the shared `agents` package.
- `content/data/recipes.cue`: `recipe.haul_water` (go to a believed source, take, go to the believed cistern, put), `recipe.haul_food`, `recipe.work`, `recipe.worship`.
- Settlements (M6-22) list how many of each role to spawn through their pieces' marks.
- Tests (golden, a small generated town): a hauler fills the cistern from the remote source; residents drink from it; kill the haulers and the cistern empties, the residents' thirst rises, and they leave the town toward water they believe in, with no event announcing it.

## Spec

- **Firm:** nothing counts or announces an emptied town. The golden test measures it from state (who stands where).
- **Firm:** important residents (named cast later) use GOAP and generics use HTN, by template (M6-15).
- **Provisional:** first values for the sizes of populations and the cistern are for tuning in M9.

## Acceptance criteria

- [ ] The golden test above passes.
- [ ] `go tool task check` passes.

## Out of scope

- Births, arrivals, new crawlspaces, rumors that drift: M8.
- The named cast: M7.

## Notes

