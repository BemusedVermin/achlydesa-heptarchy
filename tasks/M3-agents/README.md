# M3 · Agents

**Status:** outline. Break it into task files with `tasks/TEMPLATE.md` once M2 is done, so the specs reflect what M1 and M2 taught.

**Goal:** monsters and people act on their needs, on screen as full agents and off screen as region and town aggregates; generic NPCs are replaced on timers; the party can hire recruits.

**Design sources:** `docs/game-design.md` → "The living deep" (Monsters with needs, People with needs, Three tiers of people, Keeping it affordable) and "Party → Recruits", "Exotic recruits".

## Planned tasks

- **Needs and behaviors as content.** Needs with Expr curves; a behavior primitive registry (move toward, flee, consume, carry, guard, wander, rest, work, trade, follow a schedule, engage); considerations scored with Expr (utility scoring).
- **Need-driven controllers.** Full agents near the party act on their own turns through `turns` controllers that pick behaviors by need, path with `hexmap.Path`, and replace M2's `hunt` controller for creatures that have needs.
- **Monster families.** The design's sketches as content: moving clay gathering at bodies, remnants retreating from decay, soul-feeders hunting the living, the archon's creatures guarding believers.
- **Level of detail.** Region and town aggregates per family; per-phase drift; promotion when the party arrives and demotion when it leaves, keeping anyone with `Individual` (named, holding a grudge, a deserter) as a full entity.
- **People.** Residents with needs (water, food, rest, safety, belonging, faith and duty, grief), roles and customs as constraints, towns as maps with schedules.
- **Generic NPCs.** Very basic needs; replacement timers while a town is inhabited.
- **Recruits.** Hiring for a price named in what the recruit wants; recruits fill party slots; grumbling, desertion and flight mid-fight; graftable and exotic recruits.
- **An agents chooser for combat,** registered by name to replace M2's score chooser for creatures with needs.
- **Harness columns:** deaths, migrations, towns emptied.

## Exit criteria

- Moving clay walks to a corpse the party left behind.
- A resident fetches water on a schedule, and a town short of water starts to lose people.
- A recruit hired in town fights, then deserts when their wants go unmet.
- Off-screen populations drift and promote into full agents when the party arrives.
- 100 harness seeds show no degenerate camping (one family sitting in one town).

## Questions to settle before breaking it down

- Healing: does rest restore integrity only, or can severed parts regrow? The design leans on grafts to replace what the deep takes.
- Does the party itself need food and water?
- Recruit prices: how "what they want" is represented, and how the party pays it.
