---
id: M6-13
title: "Thing actions: take, put and consume"
status: todo
milestone: M6
kind: feature
depends_on: [M6-02, M6-11]
discovered_by: null
size: M
---

# M6-13 · Thing actions: take, put and consume

## Goal

Agents and the party can take from a source or container, put into one, and consume an item, all as ordinary content actions on the timeline.

## Context

- `docs/game-design.md` → "How minds work" (Supplies are things in places); "Party" → "Food and water" (drinking spends the same lake water Soak uses).
- M6-02 provides `inventory.Move`; the `consume` effect already exists (`internal/core/combat/install.go`).
- The party's eat and drink (M6-24) must reuse these actions, not copy them.

## Deliverables

- `content/data/actions.cue`: `action.take`, `action.put`, `action.consume`, with target filters by tag (M6-05), a reach of one hex, and short wind-ups.
- `content/data/operators.cue`: `operator.take`, `operator.put`, `operator.consume`, with their preconditions and effects over `has`, `near` and `drive_met`.
- `internal/core/agents/`: the target pickers (the nearest believed source or container with the item), and the claim a take makes on the stock it will empty.
- Tests: a body takes from a source until empty; two agents cannot claim the same last unit; a put into a full container is refused with a reason.

## Spec

- **Firm:** nothing here is agent-only. The actions appear in the player's menu like any other when their targets are in reach.
- **Firm:** consuming applies the item's effects through `satisfy` (M6-03), so food meets hunger by content.
- **Firm:** reasons for a refusal come from content text, shown as for any action.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Hauling as a job: M6-23. The party's use: M6-24.

## Notes

