# M7 · Tuning

**Status:** outline. Break it into task files once M6 is done.

**Goal:** the clock numbers and content values come from data: hundreds of seeded headless runs, analyzed outside Go, with changes made in content rather than code.

**Design sources:** `docs/game-design.md` → "Open questions" (Clock numbers). Tech design → "Harness and testing".

## Planned tasks

- **Long runs.** `harness run --seeds 500 --days 365` with smarter policies (a party that hunts seats, one that ignores them).
- **Metric definitions.** Seat stages and passes, deaths per town, migrations, towns emptied, export effects, fights per day.
- **Analysis** in a notebook or spreadsheet outside the repo's Go code.
- **Performance.** pprof on the off-screen tick so 500 seeds over 365 days finishes overnight.
- **Balance passes** on game length, seat speed per stage, kill setback, export frequency, and the combat placeholders from M2.

## Exit criteria

- The clock numbers are decided and recorded in the game design's open questions.
- 500 seeds over 365 days finish overnight on the owner's machine.
