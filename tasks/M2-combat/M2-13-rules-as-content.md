---
id: M2-13
title: "Rules as content: records, rules and choices in CUE"
status: todo
milestone: M2
kind: seam
depends_on: [M2-05, M2-12]
discovered_by: null
size: L
---

# M2-13 · Rules as content: records, rules and choices in CUE

## Goal

Every combat rule moves out of Go and into CUE: what makes a part weak or immune, which types a body may load, what a loaded type does to a form, and what a kind of item or grant means. Code keeps the mechanisms (primitives, selectors, timing, the action machinery) and stops naming content records, tags, list keys or rules; its own message text is the only content it names. After this, the seven forms (M2-06) are pure content, and combat can be reworked without touching Go.

## Context

- AGENTS.md rule 4, "Content is data".
- Tech design: "Designing for pivots" (the Actions and Rules seams), "Content pipeline".
- Game design: "Combat → Moves = form × type", "Combat → Types".
- This milestone's README, "The combat model" (Rules are content).
- What's in Go today that this task moves into content:
  - `internal/core/effects/effects.go`, `partEnv`: `weak` and `immune` are computed in Go from `Material.WeakTo` and `ImmuneTo` and the choice named `"type"`.
  - `internal/content/types.go`: `DamageType{Innate, Consumes, Condition}` with `TypeEnv`; the closed lists `ItemKinds` and `GrantKinds`.
  - `internal/core/anatomy`: a body's speed is the sum of its `speed` grants, computed in Go.
  - `internal/core/combat/effects.go`: `reveal` accepts only `claim.material` and defaults its source to `source.sweep`; `damage` publishes `Choices["type"]`, and evaluates `amount` once for every part it hits.
  - `internal/core/explore/follow.go`: followers start `action.step` by ID.
  - `internal/core/actions`: the only choice providers are Go functions. M2-06's old spec put the type provider, validator and modifier in Go as well; this task replaces them.

## Deliverables

- `internal/content/types.go`: `Data` on records; `Mods` on types and items; `ChoiceDef`, `Vocabulary`, and `Bundle.Rules`, `Choices` and `Vocabulary`; `PartyConfig.FollowAction`; `TimelineConfig.Speed`. Remove `Material.WeakTo` and `ImmuneTo`, `DamageType.Innate`, `Consumes` and `Condition`, `Item.Kind`, `ItemKinds`, `GrantKinds` and the `Grant*` constants. Regenerate the schema.
- `internal/content/env.go`: `Record`, `Side`, `OfferEnv`, `GrantEnv`. `EffectEnv` and `ActionEnv` grow; `TypeEnv` and `EffectEnv.Weak`, `Immune` and `Type` go.
- `internal/content/compile.go` (new): the one place formulas compile, with rules inlined.
- `internal/content/records.go` (new): looking up any record by ID.
- `internal/content/load.go`, `types_check.go`, `actions.go`, `triggers.go`: vocabulary, rule, choice and mods checks; every formula through `Compile`. Remove `TypeAllowed`.
- `internal/core/effects`: records in the formula environment, text parameters that are formulas, per-part evaluation, and the formula cache per registry.
- `internal/core/actions`: choices from content, chosen records' mods, records in `ActionEnv`.
- `internal/core/inventory`: `HolderOf` (moved from `combat`) and `Usable`.
- `internal/core/combat/effects.go`: `damage` carries every choice; `damage`, `sever` and `raise_max_integrity` evaluate per part; `reveal` is generic.
- `internal/core/anatomy`, `internal/core/timeline`: speed from a formula.
- `internal/core/explore/follow.go`: the follow action from content.
- `internal/core/api/combat.go`: `PartDamaged.Type` becomes `Choices`.
- `.golangci.yml`: forbid `expr.Compile` and `expr.Eval` outside `internal/content/compile.go`.
- `content/data`: new `rules.cue`, `vocabulary.cue` and `choices.cue`. Migrate `materials.cue`, `types.cue`, `items.cue`, `bodies.cue`, `party.cue` and `timeline.cue` with behavior unchanged.
- Tests in every package touched.

## Spec

### 1. Generic data on records

```go
// Data is what a record carries for rules to read. Code never reads a tag, prop or list key;
// formulas do.
type Data struct {
	Tags  []string            `json:"tags,omitempty"`  // free words: "innate", "piece", "clay", ...
	Props map[string]float64  `json:"props,omitempty"` // named numbers: "weight": 2
	Lists map[string][]string `json:"lists,omitempty"` // named lists, usually of IDs: "weak_to": ["type.soak"]
}
```

Embed `Data` in `Material`, `DamageType`, `Item`, `PartTemplate`, `Status`, `Faction` and `Template`. Suggestion: embed it with no json tag, so its fields are promoted the way `encoding/json` promotes them. If `cue get go` or `Value.Decode` mishandles the embedding, repeat the three fields on each record instead, and say so in Notes. `ActionDef.Tags` stays as it is.

What moves:

| Was (Go) | Becomes (content) |
| --- | --- |
| `Material.WeakTo`, `ImmuneTo` | `lists.weak_to`, `lists.immune_to` |
| `DamageType.Innate` | the tag `innate` |
| `DamageType.Consumes` | `lists.consumes` (one item ID), plus a `consume` effect in the type's `on_start` (section 6) |
| `DamageType.Condition` over `TypeEnv` | the type's `requires` over `ActionEnv` (section 6) |
| `Item.Kind`, `ItemKinds` | tags; `vocabulary.tags.items` closes the set |
| `GrantKinds` | `vocabulary.grants` |
| The rule that a weakness can't also be an immunity | a CUE constraint in `materials.cue` (suggestion: a hidden field that lists the overlap, unified with `[]`) |

Nothing reads `Innate`, `Consumes` or `Condition` at run time yet (`TypeAllowed` has no caller outside tests), so moving them changes only content and the loader. Only an item with `part` set can be grafted, so the check that only a `piece` has a part goes away.

**Vocabulary.** Content closes the sets it wants closed; anything it doesn't declare stays open.

```go
// Vocabulary lets content declare which tags, grant kinds and list values are allowed.
type Vocabulary struct {
	// Tags maps a collection ("items", "types", ...) to the tags its records may carry.
	Tags map[string][]string `json:"tags,omitempty"`
	// Grants lists the grant kinds parts may give. Empty allows any.
	Grants []string `json:"grants,omitempty"`
	// Lists maps "<collection>.<key>" to the collection its values must name: "materials.weak_to": "types".
	Lists map[string]string `json:"lists,omitempty"`
}
```

Collection names are the `Bundle`'s json keys. The loader reports a tag its collection doesn't declare, an undeclared grant kind, and a list value that isn't an ID in its declared collection. Start `vocabulary.cue` with today's item kinds as item tags, today's three grant kinds, and `materials.weak_to`, `materials.immune_to` and `types.consumes`.

**Grants.** Every grant has a value; content writes it (an arm is `{kind: "action", value: 1}`), and the loader rejects a grant whose value is 0. `anatomy.Grants` sums the values for every kind. `TimelineConfig.Speed` is an Expr over `GrantEnv`; it starts as `"grant.speed"`. A body's speed is that formula; an entity with no body keeps `default_speed`.

```go
// GrantEnv is what the speed formula sees: the totals of a body's functioning parts' grants, by kind.
type GrantEnv struct {
	Grant map[string]int `expr:"grant"`
}
```

### 2. Records in formulas

```go
// Record is what a formula sees of one content record. Tags, props and lists it doesn't have read
// as empty: no tags, 0, an empty list.
type Record struct {
	ID    string              `expr:"id"`
	Tags  []string            `expr:"tags"`
	Props map[string]float64  `expr:"props"`
	Lists map[string][]string `expr:"lists"`
}

// Side is what a formula sees of one body: its template, one of its parts, and that part's material.
// A record is empty (ID "") when there's none.
type Side struct {
	Template Record `expr:"template"`
	Part     Record `expr:"part"`     // the part's template
	Material Record `expr:"material"` // the part's own material, which a body plan may override
}
```

`(*Registry).Record(id api.ContentID) (Record, bool)` finds a record by its ID's prefix (`material.` in materials, `type.` in types, and so on, as `checkIDs` already enforces). For an ID it doesn't know, it returns `Record{ID: id}` and false. Build the records once per registry.

**The effect environment** gains:

- `Source Side` (`source`): the acting body. Its part is the first channel the action occupies that's a part, such as the striking arm.
- `Target Side` (`target`): the target body and the target part.
- `Chosen map[string]Record` (`chosen`): every choice the action or event carries, as the record its value names. A value that names no record, such as a direction, is `Record{ID: value}`.
- `Pack map[string]int` (`pack`): see below.

It loses `Weak`, `Immune` and `Type`; `rule.weak`, `rule.immune` and `chosen.type.id` replace them.

**The action environment** gains `Source`, `Target`, `Chosen` and `Pack` with the same meanings. `Choice` stays.

**The pack.** `inventory.Usable(w, body, target ecs.Entity) map[string]int` counts, per item, the stacks in the pack of `inventory.HolderOf(w, body)` that can be used on the target: unbound stacks and stacks bound to the target. With no target, every stack counts. `HolderOf` moves from `combat` to `inventory` unchanged (the party when the body is a member, otherwise the body).

**Events.** `PartDamaged.Type` becomes `Choices map[string]string`: every choice of the action that dealt the damage. `PartDamaged.Choices()` returns it, so trigger formulas see it as `chosen`. The clay trigger in `materials.cue` becomes `chosen.type.id == "type.burn"`.

### 3. Rules

`Bundle.Rules map[string]string` (`rules`) holds named formulas. Any formula in content can say `rule.<name>`. The compiler inlines the rule's source in its place before type checking, so a rule is checked in the environment of each formula that uses it and takes its type from that use. Suggestion: an `expr.Patch` visitor that replaces the member node `rule.<name>` with the rule's parsed tree. Expr runs patch visitors before its final type check.

- Rules may use other rules.
- Load errors: a formula naming an unknown rule; rules that use each other in a cycle; a rule name that isn't lower-case words. Each error names the path of the formula.
- A rule nothing uses is a warning (`Registry.Warnings`).

**One compiler.** `func (r *Registry) Compile(src string, env any, options ...expr.Option) (*vm.Program, error)` is the only way to compile a formula. Move costs, step ticks, climb cost, speed, action formulas, mods, choice offers, trigger conditions and effect parameters all go through it. Add a forbidigo rule against `expr.Compile` and `expr.Eval` everywhere else.

**Starting rules** (`rules.cue`), which keep today's behavior:

```cue
rules: {
	weak:   "chosen.type.id in target.material.lists.weak_to"
	immune: "chosen.type.id in target.material.lists.immune_to"
}
```

**The effects cache.** `effects` caches compiled parameters by source text. With rules inlined, the same text compiles to something different after a hot reload, so key the cache by the registry's `Hash()` as well. The roll check ("a primitive with a Check cannot read roll") must look at the formula after inlining, so a rule that reads `roll` counts.

### 4. Effect parameters

**Text parameters that are formulas.** A text parameter whose value starts with `=` is an Expr over `EffectEnv` that returns a string; the rest of the value, trimmed, is the formula. Any other value stays a literal. `Params.String` evaluates it. It compiles at load like any other formula.

- `VerifyContent` functions check literal values only.
- A formula's result is checked when it's applied: an unknown ID is an error naming the parameter.
- Selector parameters (`KindBodies`, `KindChannels`) stay literal.

**Per-part evaluation.** A primitive that acts on a channel selector evaluates its other parameters once per selected part, with that part as the target part (`target.part`, `target.material`) and its body as the target. Parts go in UID order, as now. This applies to `damage`, `sever`, `raise_max_integrity` and `reveal`. A per-part formula that reads `roll` draws once per part.

Suggestion: `Params` gains `For(c timeline.Channel) Params`, which returns the same parameters over a copy of the context whose target is the channel's body and whose target part is the channel's entity when it's a part.

Without this, a Sweep can't hit each part by its own weakness: `amount` is evaluated once, with no target part.

### 5. Choices from content

```go
// ChoiceDef is a choice an action can ask for, served from a collection of records.
type ChoiceDef struct {
	From  string `json:"from"`            // a collection: "types", "items", ...
	Offer string `json:"offer,omitempty"` // Expr bool over OfferEnv; empty offers every record
}

// OfferEnv is what a choice's offer formula sees.
type OfferEnv struct {
	Record Record         `expr:"record"` // the candidate
	Source Side           `expr:"source"` // the actor; only its template is filled in
	Pack   map[string]int `expr:"pack"`   // the actor's whole pack: there's no target yet, so bound stacks count
}
```

`Bundle.Choices map[string]ChoiceDef` (`choices`) is keyed by choice name.

- A choice name is served by the Go provider registered under it (only `direction` today), or else by the content choice with that name. Values come in ID order.
- `actions.ValidateContent` reports an action asking for a choice neither serves, and a content choice that shadows a Go provider.
- The loader checks that `from` names a record collection and that `offer` compiles.
- `actions` reads content choices when it evaluates, so hot reload picks up changes with nothing to re-register.

Starting content (`choices.cue`), keeping today's meaning:

```cue
choices: type: {
	from:  "types"
	offer: "\"innate\" in record.tags || any(record.lists.consumes, {pack[#] > 0})"
}
```

### 6. Chosen records change the action

```go
// Mods are what a record changes about an action it's chosen for: a type loaded into a form, a
// piece held for a graft. Formulas are Expr over ActionEnv.
type Mods struct {
	Requires string        `json:"requires,omitempty"` // false refuses the action, with Refusal as the reason
	Refusal  api.ContentID `json:"refusal,omitempty"`  // a text ID; empty means text.actions.requires
	Windup   string        `json:"windup,omitempty"`   // ticks added to the wind-up
	Recovery string        `json:"recovery,omitempty"` // ticks added to the recovery
	OnStart  []EffectRef   `json:"on_start,omitempty"`
	OnLand   []EffectRef   `json:"on_land,omitempty"`
}
```

Embed `Mods` in `DamageType` and `Item`. When an intent makes choices, go through them in the order of the action's `choices`. For each value that names a record carrying `Mods`:

- add its wind-up and recovery ticks;
- append its start and landing effects after the action's own and before any Go modifier hook's;
- check its `requires`. False makes the plan refused with a reason, like a validator, so the option shows as disabled with why.

A landing rebuilds the plan, so it runs the mods again to get its landing effects. `EffectSites` lists the mods' effects so they're checked at load, and the loader checks that `refusal` names existing text.

`AddValidator` and `AddModifier` stay as Go seams; nothing in the repo registers them.

**Migrating `types.cue`.** Each consumable type gets `lists: consumes: [<item>]` and `on_start: [{do: "consume", params: {item: <item>}}]`. Suggestion: a CUE definition such as `#Consumable` so each item is written once. The Name and Keepsake conditions become `requires` over the pack: `pack["item.keepsake"] > 0` for Keepsake. Name needs a name item; add `item.name` (bound, tagged `name`) and `requires: "pack[\"item.name\"] > 0"`. Neither has a caller yet, so this keeps today's behavior.

### 7. The last hard-coded content

- **`reveal`** becomes `reveal(claim, value, source, parts?)`. For each part `parts` selects (a channel selector; default `target:all`), the knower learns a fact about that part. Its value is `value` evaluated with that part as the target part (usually `"= target.material.id"`). The check that only `claim.material` may be revealed goes, and so does the `source.sweep` default: `source` is required. `VerifyContent` checks that the claim and the source exist.
- **`damage`** publishes every choice in the context as `PartDamaged.Choices`.
- **Following.** `PartyConfig.FollowAction` (`follow_action`) names the action a follower steps with. The loader checks that it exists and targets a direction. `party.cue` sets it to `action.step`.
- Remove the constants `defaultRevealFrom`, `materialClaim` and `stepAction`.

## Acceptance criteria

- [ ] Content is migrated with behavior unchanged: the existing tests pass with only fixture updates (`weak`, `immune` and `type` in test formulas become `rule.weak`, `rule.immune` and `chosen.type.id`).
- [ ] Changing a fixture's `weak` rule to `chosen.type.id in target.material.lists.weak_to || "brittle" in target.part.tags` changes which parts a formula sees as weak, with no Go change.
- [ ] A formula naming an unknown rule, and two rules that use each other, fail at load with errors naming the formula's path.
- [ ] A rule that reads `roll`, used in a parameter of a primitive that has a Check, is rejected at load.
- [ ] Hot reload: after `SwapContent` with a changed rule, a formula already evaluated before the swap computes the new result.
- [ ] Missing data reads as empty: with no type chosen, `chosen.type.props.nope == 0` and `"x" in target.material.lists.nope` is false, with no error.
- [ ] The `type` choice comes from `choices.cue`: a body whose party carries one lake water is offered Soak, and isn't once it's spent. Innate types are always offered. Values are in ID order.
- [ ] Mods: a test type with `windup: "2"`, `recovery: "1"`, a `consume` on start and an effect on landing changes a test action's timing by exactly that, spends the item at the start, and applies its landing effect after the action's own. A false `requires` shows the option disabled, with the type's refusal text.
- [ ] Text formulas: `consume(item: "= chosen.type.lists.consumes[0]")` spends lake water when Soak is chosen. A formula that names an unknown item fails when applied, with an error naming the parameter.
- [ ] Per part: `damage(parts: "target:all", amount: "rule.weak ? 2 : 1")` with Soak on a body with one clay arm takes 2 from the clay arm and 1 from every other part.
- [ ] `reveal` with `value: "= target.material.id"` teaches each part's own material, including a body plan's override (a clay arm on a flesh body).
- [ ] Vocabulary: an item with an undeclared tag, a part with an undeclared grant kind, and a material whose `weak_to` names a non-type each fail at load. With nothing declared, all three load.
- [ ] Changing `timeline.speed` changes a body's step ticks with no Go change.
- [ ] A follower steps with whatever action `party.follow_action` names.
- [ ] Outside `_test.go` files and generated schemas, nothing under `internal/` names `weak_to`, `immune_to`, `claim.material`, `source.sweep`, `action.step` or `Choices["type"]` (check with `grep -rn`).
- [ ] `go tool task check` passes.

## Out of scope

- The forms, Soak's weight, Press's stagger, Stand's status and the lethal numbers (M2-06).
- Factions on templates (M2-07).
- Cell-targeted actions (M1-20.F2).
- Exposing records to formulas other than the ones listed (move costs, step ticks): add them when content needs them.

## Notes
