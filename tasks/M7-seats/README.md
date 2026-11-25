# M7 · Seats

**Status:** outline. Break it into task files once M6 is done.

**Goal:** one seat works end to end: its stages, the piece's readings, investigation, killing and passing, and wound scoring choosing the next holder.

**Design sources:** `docs/game-design.md` → "Candidates and seats", "Core loop and time management" (the piece as sensor, Investigation, Kills and the clock), "Simulation" (Seats, the cast model, Authored first, generated after).

## Planned tasks

- **Seat entities.** Seven seats from content, one per peak, each a state machine (Offer, Fever, Shaping, Price, Enshrined) with stage durations as Expr, advanced by a PerDay system.
- **Candidates as components.** `Candidate{Seat, Wish, CurdleMode}` on ordinary person templates; authored candidates only for now.
- **Wounds and passing.** Wound scoring over people's losses; a kill sets the seat's clock back by a content amount and passes it to the most wounded person.
- **Stage effects.** The ascending body changes through stages (parts added or swapped from content); archon prices as composed effects.
- **The piece of Serilmen.** An item whose readings (region and progression, never who) form a view section, read slightly stronger near other pieces.
- **Investigation.** Clues from people, places, bodies and dreams write facts; names and keepsakes become items; a journal screen as a view over the party's facts.
- **Where a seat is.** The piece reports a seat's region. Procgen doesn't divide the world into regions, so M7 defines what that is.

## Exit criteria

- Read the piece, investigate a candidate, find them, kill them, and watch the seat pass and the piece's reading shift as if the seat were wounded.
- A false positive (a grieving person in the same region) is possible and indistinguishable to the piece.

## Questions to settle before breaking it down

- What exactly the companion beat needs from the seat system.
- Act I: is the first god the already-enshrined archon, or a second seat completing? (Open in the game design.)
- What the piece's region is: an area around the holder, a direction and a depth, or something else. It has to be coarse enough to hold a holder and their false positives.
