# Achlydesa: The Dreaming Heptarchy — Game Design Document

Nov 20, 2026 · @Liam McGuire

## Overview

Achlydesa: The Dreaming Heptarchy is a turn-based dungeon crawler with time management about the illusion of choice. People on the clay plain who have lost something are ascending into archons in the deep. A Reaper and an elevated salken hunt the candidates, and learn late that there are always seven seats, so their kills only ever chose who sits.

- **Touchstones:** Fear & Hunger for unsettling lore and limb-level brutality. SMT and Final Fantasy for weakness-driven, turn-manipulating combat, as inspiration only, not a copy. Pathologic for open exploration, time pressure and failure that continues the story.
- **Tone:** surreal, weird and unsettling throughout.
- **Core theme:** the illusion of choice, not literal loss of autonomy. Moment-to-moment choices (combat, routes, supplies, timing) are fully real. The illusion lives in what those choices add up to, and the characters are not content with it.
- **Genre:** a dungeon crawler where combat is the backbone: regular fights, not one-off boss encounters. Menu-based: every action is picked from a menu. Fights happen on the same hex grid as exploring, and a form's range is a distance in hexes, not a range band.
- **Failure:** handled the Pathologic way. The story continues; there is no roguelike permadeath.
- **Ending:** one ending, not a menu of endings. What varies is which seven archons sat.
- **Presentation:** ASCII for the MVP.

## Setting

The game takes place on the clay plain and in the deep beneath its crawlspaces, merged with the archon setting's Demiurge. Existing canon is used as is; new canon may be invented wherever combat or story needs it.

### Existing canon the game leans on

- **The plain:** pastel peach clay, dundies, the salk (surface), crawlspaces linked by soops, ibben (freemen) and salken (slaves), the test at ten, the Osten, the thrukk, Reapers, standing at attention indoors, the lunn, and the Skadders (who preach fewer births).
- **Theology:** soul is divisible and passed down, never newly made. Suffering erodes it, death by hibblejaunt annihilates it, and a life without suffering earns the immortal body of yore. In this game the doctrine's truth stays ambiguous.
- **Hibblejaunts:** sacred, synthetic animals from beyond the rift that fight each other constantly. Never fought by people.
- **Pieces of Serilmen:** buried ancient technology, believed to be the parts of a dismembered god. A slave who finds one is elevated and must leave the fields.
- **The mountains:** the projection of the rift between worlds. Always at the same distance, and they never change.
- **The far side:** the algocracy. Its algorithm was built for an unchanging world, and its people still vote.
- **The Demiurge and archons:** an all-powerful, frightened, ignorant being enshrines archons. Archons believe what they did was just, their forms reflect what they desire, and there is always a way to kill them.

### New canon for this game

| Element | What it is |
| --- | --- |
| The deep | Pre-entropy ground below the crawlspaces. Where wishes are made, and the game's one shared megadungeon, warped region by region by ascensions. |
| Seven peaks | The mountains have seven peaks, one per archon seat. Visible all game; the meaning is revealed late. |
| Exports | The far side's algorithm pushes its decay through the rift onto the plain. It is on the ballot every time and overridden every time. |
| The clay filter | The clay filters exported decay the way it purifies rain. That is why the deep is undecayed, and every export fills the filter further. |
| Export days | Irregular weather: salk dust thickens, old things in the deep die faster, and regions collapse or open overnight. |
| Pre-entropy remnants | Things in the deep that have never died, now dying for the first time. Their severed parts may keep fighting. |
| Moving clay | Clay that has held the dead long enough starts to move. |
| Daylight jars | Daylight sealed in clay jars on the salk. |
| Violet blood | Hibblejaunt blood never dries, and it burns rift-made flesh. |
| Lees | The dregs left in the thrukk after the dead are pressed. |
| Names in the deep | A true name wounds its owner, and speaking it uses it up. |
| Standing has force | In the deep, anyone who stands freezes everyone who crawls. |
| Early ascension | Looks like slow dying: fever, wasting, weeks of not quite going. |
| Serilmen's truth | The pieces are the refugees who died holding the door shut behind the crossing. The defectors (Salumen) opened it. Reassembled, Serilmen can reach the direction nothing on the plain can travel. |
| Graft will | Grafted pieces want to be reassembled. They sometimes act on their own and quietly shape the options offered. |
| Fatherless children | A child with no father's household goes to the fields at birth. |
| Reapers' standing | Reapers don't marry. When a Reaper stands in a room, what happens there is lawful. |

## World and towns

The towns are fixed islands inside the deep, at different depths, with generated dungeon all around them. Depth is the geography: the deeper the town, the older and less decayed its ground, and the stranger it is. The surface is uniform decayed clay, so variety comes from what lies underground.

**One world, no seams.** The salk, the towns and the deep are one continuous world. The party walks from anywhere to anywhere: nothing loads behind a gate, a door or a transition. Towns and regions of the deep are areas of that one world, never maps of their own.

| Depth | Environment | Town | Local material or service |
| --- | --- | --- | --- |
| Surface | The salk: open, flat, the flock as weather | Salken village | Grafting, Violet blood, daylight jars |
| Shallow | Ordinary tunnels, fully decayed | The Reaper's crawlspace | Reaping calls, the Reaper's standing |
| Middle | Underground lakes | Lake crawlspace | Soak (lake water) |
| Middle | Seams of dead crawlspaces stacked over centuries | Coal crawlspace | Burn (coal) |
| Deep | High-pressure zones | Thrukk crawlspace | Lees, Press training |
| Deeper | Barely decayed ground where nothing rots: the dead lie as they fell, and wounds don't fester | The still town | Healing without infection, until the town starts to rot |
| Deepest | The edge of pre-entropy ground | The archon's crawlspace | The one enshrined archon, its worshippers and beneficiaries |

- Only the surface and the shallow soops are fixed. Every route between deeper towns is dungeon, and regenerates on export days.
- Towns keep their layout but not their condition. As exports push decay downward, deep towns change: the still town starts to rot, and old matter retreats past the archon's crawlspace.
- Towns are open to the deep. Nothing scripted keeps monsters out (see The living deep).
- **Crawlspaces are authored and emergent.** The story's crawlspaces are hand-built places. People can also dig and settle new ones, and abandon old ones, when that meets their needs.
- **The world is bounded.** Its size is content. People arrive from beyond it and leave for beyond it at immigration and emigration points.
- **Small at first.** The first world is small and built by hand: the towns, and a short stretch of deep between them. Procgen generates the rest of the deep, and its regions, later.

### The enshrined archon

- One of the seven seats is already enshrined when the game starts, so six are active. Its archon rules the deepest crawlspace.
- It sits in a seat the Reaper passed long ago without knowing.
- Killing it early reopens its seat, which passes like any other. The archon's town is a trap: killing its god hands a finished seat to someone new.

### What to do in town

| Activity | What you do | What it feeds |
| --- | --- | --- |
| Investigate | Talk to people, learn names, collect keepsakes, notice who's feverish | Name and Keepsake types; telling holders from false positives |
| Answer reaping calls | The Reaper is summoned to the slowly dying: some are dying, some are feverish candidates, some are false positives | The core dilemma at town scale, and seat passing |
| Rest and dream | Sleep moves time forward and brings a dream (see Dreams) | Clues and lore in small doses |
| Attend town events | The Osten, the test at ten, funerals at the thrukk, each on its town's calendar | Being in the right place at the right time; candidates surface at these |
| Prepare | Buy, barter and gather the local material | Consumable types |
| Heal and graft | Recover injuries over days; graft in the salken village | The body-as-actions economy |
| Eat and drink | Buy or draw food and water for the descent | How long the party can stay below |
| Requests | Find someone who went below, bring a body back for the thrukk | Quests, and the cast's relationships |

**Healing.** Resting in town restores a damaged part's integrity over days, at a rate set in content. A severed part never grows back: a graft, or reattaching a part the party kept, is the only way to replace one.

Access matters: a salken can only enter ibben towns through the lunn, the way the dead come in, while the Reaper's standing opens doors. Who is in the party decides who will talk.

## The living deep

The deep is generated around the fixed towns and populated by monsters and people with needs, not by scripted events. Emergent play is the priority: when something walks into a town, it's because what it needs is there.

**Outcomes emerge; nothing forces them.** Every creature and person is an individual with a body, needs, an inventory and memory, all the time, on screen or off. A migration, a crawlspace emptying, a recruit walking off: each is agents choosing what meets their needs, never a system or a timer that makes it happen or an event that announces it. Resources are things in places: the water in a cistern is water somebody carried there. A cistern, a larder or a pile on the floor is a container: something with a place and an inventory, filled and emptied by whoever carries things there.

**Needs change.** What an agent needs isn't fixed. Needs appear, fade and change weight with age, losses, grafts, faith, place and whatever has happened to it: a death adds grief, a grudge adds revenge, a child growing up takes on an adult's needs. Behavior drifts over a run because needs do.

**What happened is found, not announced.** The player learns that a crawlspace was abandoned, or that someone left long ago, the way anyone in the world would: from what the place looks like, and from what people remember and say, including roughly when.

### Procgen and exports

- **Fixed:** the towns, the salk, the shallow soops, and set pieces such as the oldest layer where the reveal happens.
- **Generated:** everything else, grown right up to the edges of the fixed places, so nothing shows where a town ends and the generated deep begins.
- **Reseeding:** each export regenerates the generated deep from a seed derived from the world seed and the export count, so any run can be replayed.
- **Seats** keep their stage and progress; their regions land somewhere new.
- **The map goes stale.** What the player charted no longer exists.
- **The deep changes character over the game.** Each export fills the clay filter, so the upper deep grows more decayed and old matter retreats downward.
- **Guarantees:** every town connects to the deep, every active seat is reachable, and materials appear where they belong (coal near dead crawlspaces, moving clay near the thrukk, old matter deep).

### The hex grid

Exploring and fighting both happen on the hex grid, on the same clock; actions are picked from menus.

- **Why hexes suit the deep:** six even directions make organic caves and tunnels rather than corridors and rooms, and distances are uniform for creature movement.
- **ASCII handles it:** each hex is one glyph, two columns wide, with alternate rows offset by one column. Terminal cells are about twice as tall as they are wide, so the hexes come out roughly even.
- **Human scale:** one hex is 1.5 m across and one depth step is 1.5 m, so a person fills a hex. Everything is built from cells at that scale: a house is walls, rooms and a door; a well is a shaft; a lake is water; a town runs hundreds of hexes. No cell stands for a whole structure.
- **Headroom:** every body fills one cell, standing or crawling. There's no separate headroom, and low passages are the norm.
- **Depth is continuous and seamless:** the deep is a 3D world of hex columns, layered as finely as Minecraft's blocks, so soops cross over and under each other, slope gently and drop off ledges. The plan view shows what the party can see as one continuous map, shaded by how far above or below the party it lies. No layers are ever visible.

```
 . . # # .
. @ . . # .
 . . ~ ~ .
. . . ~ . .
```

- **Later graphics:** Godot's tile maps support hexagonal tiles natively, so moving to Godot doesn't change the map model.
- **In the engine:** the hex topology (coordinates, neighbors, distance, line of sight, pathfinding) lives in `Core`; the frontend only decides how to draw it.

### An export with the party below

- The party's current chamber is kept as a pocket; everything beyond it regenerates around them.
- A route to at least one town is guaranteed, but not necessarily the one they came from.
- What the party carried is all they have until they reach a town.

### Monsters with needs

Every creature has needs and behaviors that meet them. Needs and behaviors are data, composed from engine primitives such as move toward, consume, carry, guard and flee. There is no incursion system.

Sketches of what that produces:

| Creature | Needs | What emerges |
| --- | --- | --- |
| Moving clay | To hold the dead | Gathers at the thrukk and wherever fights left bodies, including the party's own kills |
| Pre-entropy remnants | Undecayed ground | Retreat downward as exports spread decay; cluster near the still town |
| Wish-born creatures | Whatever their wish's price feeds on | Go where the price's targets are, which is usually a town |
| Soul-feeders | Soul | Hunt the living, and the slowly dying |
| The archon's creatures | To protect the archon and its worshippers | Attack outsiders, spare believers |
| The flock | To fight over females | Moves across the salk, killing whoever is in the fields |

What falls out of it:

- **Deaths in towns become candidates.** They add losses to the cast model, which shapes who takes the next passed seat.
- **Mauled people mean more reaping calls,** and more chances to mistake a feverish candidate for a wounded neighbor.
- **Exports displace whole populations,** so the days after one are chaotic.
- **Families fight each other,** so the deep has its own ecology.
- **Players can use needs:** bait moving clay with bodies, or lure a soul-feeder away. A lure that kills a candidate is still a kill, and the seat passes.
- **Nothing scripted protects towns,** so a town can empty out if things go badly.

### People with needs

Residents run on the same needs engine as monsters, plus a social layer. One agent model serves both, so the code is shared.

- **Needs:** water, food, rest, safety, belonging, faith and duty, and grief after a loss.
- **Behaviors:** work (hauling water, mining coal, farming the fields, pressing the dead), trade, worship, attend town events, call the Reaper for the slowly dying, flee, migrate, dig and settle a new crawlspace, and go below.
- **Roles and customs act as constraints:** ibben and salken, jobs, the lunn, standing at attention, and the test at ten sorting children.

What falls out of it:

- **Candidates leave a trail.** When a seat passes, the offer finds the most wounded person and draws them below. Noticing who slips out of town at night is investigation.
- **Rumors travel and drift.** News moves only when people carry it, between towns and through the deep, and it distorts with each retelling. Drifted rumors are where many false positives come from.
- **Quests come from unmet needs.** A resident who can't meet a need asks the party: find a missing son, bring water, carry a body to the thrukk. There is no quest list. After the MVP, a request becomes a story like any other (see the Story tab).
- **Towns can empty.** If water fails or monsters keep coming, people leave for other crawlspaces, or dig new ones, and a town can be abandoned. Nothing counts it as emptied; nobody lives there any more.
- **Traders cross the deep.** When a trader dies on the way, the next town runs short.
- **Plot armor is allowed.** Authored characters can be flagged as protected in their data; everyone else can die, and the seat simulation adapts.

### Three tiers of people

| Tier | Who | Simulation | Death |
| --- | --- | --- | --- |
| Party | The two protagonists and the fixed roster | Authored, never generated | Protected where plot armor is set |
| Named cast | The authored core cast and generated candidates | Full needs agents with relationships, rumors and wounds | Permanent; plot armor optional per character |
| Generic | Guards, merchants, recruits and similar | Very basic needs and wants | Permanent; nobody respawns |

**A party wipe is not game over.** When every member of the party is dead or downed, the deuteragonists wake at the town they last rested in, with some time consumed. The party mends and wakes after `party.wake_ticks`, where it last rested (the opening cell until towns exist); the dead stay dead, and a party with nobody downed stays lost. What the enemies did meanwhile, and any other cost, are still open.

**New people come from births and arrivals,** never timers. Children grow up and take the test at ten. People arrive from beyond the world, and leave for it, at its immigration and emigration points, when that meets their needs. An abandoned crawlspace stays empty until someone has a reason to live there.

### Grudges and memory

Survivors remember, and their grudges come from the memory and needs systems already in the game.

- **Encounters are recorded.** Every agent that survives the party remembers who maimed it, who fled and who left it behind.
- **Bodies remember.** A creature that lost an arm comes back with whatever replaced it. One that lost an arm to Cut might return with a clay arm, weak to Soak instead.
- **Revenge is a need.** A grudge becomes a need in the same engine, so a nemesis hunts the party the way moving clay hunts the dead.
- **Deserters persist.** A recruit who fled keeps living, with a grudge or gratitude, possibly carrying a piece of Serilmen.
- **The deepest nemesis is a seat.** An abandoned recruit is a wounded person, and wounded people take passed seats. The archon who comes for the party can be someone it left behind.
- **Stories drift.** Rumors about the party and each nemesis grow in the retelling, and change how towns receive the party.

### Keeping it affordable

- **Everyone is always an individual.** There are no counts standing in for creatures or people, anywhere.
- **Off screen, less often and coarser.** Far from the party, agents decide less often and travel a coarse graph of the world instead of cell by cell. Near the party, they move and fight cell by cell.
- **The world is bounded,** so the population and the pathing stay finite.
- **The harness measures; the game doesn't announce.** The headless harness counts outcomes from the world's state (deaths, where people are, crawlspaces nobody lives in), so degenerate outcomes such as every monster camping one town show up in tuning rather than in play.

## Story

A Reaper and a cast-out salken team up to kill people who are becoming gods, and end the game by closing the door those gods come through. This story is scaffolding and will be reworked; the opening in particular is provisional.

### Premise

People on the clay plain who have lost something go down into the deep and come back changed, a little more each day. A Reaper and an elevated salken are thrown together by the first case. They set out to kill the candidates before they finish becoming gods.

### Protagonists

Each has exactly what the other lacks, which is why they stay together.

|  | The Reaper | The elevated salken |
| --- | --- | --- |
| Role | The plain's mercy: quick deaths for the slowly dying, so suffering doesn't erode their souls | The slave doctrine's god-challenger: elevated by finding a piece of Serilmen, and required to leave the fields |
| Wants | To know whether his mercy was murder | For their elevation to mean something |
| Can | Kill lawfully, knows everyone, gets into every crawlspace | Sense what is ascending through the piece, and cut what nothing else can |
| Can't | Tell a candidate from someone truly dying, or hurt what a candidate becomes | Go anywhere, or get anyone to talk to them |
| Hidden truth | Early ascension looks like slow dying, so his reaps this season have been passing seats | The piece steered every step of their "choice" to leave the fields |

Neither protagonist can ever become a candidate.

### Opening (provisional)

The salken's piece drags them across the salk to a crawlspace, and they enter through the lunn, the way the dead are carried in. The ibben assume they are dying and call the Reaper.

He arrives to reap the salken. The piece turns hard toward his other patient, a woman due to be reaped tomorrow, and the salken says plainly that she isn't dying. Her husband has already promised to marry her sister once she's gone, and the sister is in the room. If the player spares her, she goes into the deep that night, and the husband declares her dead and marries the sister anyway. If the player reaps her, he marries the sister within the week, starts wasting within days and goes down himself. Either way the duo follow, and the choice only changed who.

### Act I: The first god

1. The opening above, ending with the first descent into the deep.
2. The fixed companions join, including one who lost the woman he loved at an Osten and is looking for her. Nobody knows yet that she's a candidate.
3. A seat further along than anyone knew completes on schedule. A god walks out onto the salk and some crawlspaces start to worship it. It came from a seat the Reaper passed.

### Act II: The hunt

1. The loop: read the piece, investigate, crawl, kill. Each kill sets a seat back, so it seems to work, and new candidates seem like fresh discoveries.
2. Export days begin and the deep shifts overnight. Scraps of lore hint at a far side.
3. The grafts start pulling, and the party grows stranger.
4. Midpoint: the companion's beloved, the bride taken at the Osten, is a candidate. Killing her passes the seat to the companion; sparing her leaves the companion a false positive.
5. The reveal, in the oldest layer of the deep: seven peaks, seven seats. Every kill only handed a seat to someone new.

### Act III: The seven peaks

1. The goal turns from prevention to buying time and choosing who sits.
2. Serilmen's truth is learned, and with it why the grafts pull.
3. The far side's lore: the export on the ballot, overridden every time, through the same door.
4. Climax: the party, by now more Serilmen than human, goes where nothing on the plain can go and closes the door. The seated archons defend their seats, so the final fights are against whichever seven the player let sit.

### Ending

- The mountains are gone; for the first time, the horizon is empty.
- No new seats and no new archons. Archons die when they are killed; any not killed in the climax are still out there.
- No more exports. The clay holds and the deep lasts longer. The far side keeps its own decay, which is karma for exporting it.
- The protagonists did what every candidate did: reached for a power to fix the world while someone unseen pays.
- The final lore implies Serilmen was always going to finish closing the door. The characters believe they chose; the player is left unsure.
- Final image: the flock, violet with each other's blood, under an empty sky.

### After the MVP

How the story actually runs (scenes unlocked by world state, the director, casting, strain and threads), the draft story beats, and emergent stories are post-MVP. They live in the Story tab (`docs/story.md`).

## Candidates and seats

There are always seven seats, one per peak, and they always fill. Killing a candidate never prevents an archon: it sets that seat's clock back some and hands the seat to someone new. The player learns this only in a late lore reveal.

### Seats

- The ascension belongs to the seat, not the person. A new holder picks up at the seat's stage, minus the setback from the kill.
- The seat passes to whoever the kill wounded most, so the player's kills author the next tragedy.
- A seat moves through five stages, each emitting world events: Offer, Fever, Shaping, Price, Enshrined.

&#91;embedded content: seat lifecycle · 5 stages, kills hand the seat on\]

Kills can land at any stage before a seat is enshrined, and none of them close it.

One seat is already enshrined when the game starts, so six are active (see World and towns). Killing an enshrined archon reopens its seat, which passes like any other.

### Candidates

- Candidates are characters within the story, never strangers on a list.
- Each chose ascension with good intentions, out of a tragic past. Wanting the immortal body is what the plain's faith teaches, so by its standards they have done nothing wrong.
- Ascension is one-way: no candidate can be turned back or talked down, and the player cannot influence what they become.
- The dilemma: killing early means killing a sympathetic person for something they haven't done yet, possibly by mistake. Waiting means harm lands first.

### How wishes curdle

- **From the wound, not the wording.** What they become comes from what they lacked, so two people with the same wish become different archons. Hearing a wish tells you almost nothing.
- **The wish works.** No literal-genie loophole. The horror is the price, which lands somewhere the wisher never looked.
- **Someone is better off.** Every archon has beneficiaries, so killing one takes something real from real people.
- **Never "kill everything that fits."** Curdles that reduce to exterminating a category are predictable and are out.

### False positives

Some people grieve, pray and show the signs but are never ascending. Killing one is killing someone who was never going to become anything.

### The companion beat

One fixed companion is bound to a candidate they love. Kill that candidate and the seat passes to the companion. Spare them and the companion stays a false positive. Either way the seat fills: the choice is who, never whether.

### Sample candidates

| Candidate | Wish | What the wish does | Who benefits | Who pays |
| --- | --- | --- | --- | --- |
| An orphan | No child without a parent | He becomes the parent no one can lose. Children in his crawlspace stop growing, so none reaches the test at ten and no new salken are made | The salken, who worship him | No one grows up, and real parents slowly stop recognizing their children |
| A woman accused of being a vampire | No mother loses her soul to a child | Mothers keep their souls, but soul is never newly made, so it is taken from everyone nearby, a little at a time | Mothers | Everyone, slightly. She becomes what she was falsely accused of being |
| One who watched a loved one die slowly | No one suffers in dying | Dying stops hurting. People stop fearing the salk and walk calmly out to the flock | Everyone, if the doctrine is false | If the doctrine is true, people are annihilated gladly. The Reaper's work becomes pointless |

Four more starting candidates are still to be written.

## Core loop and time management

Each day the player decides which descent they can afford, given what the piece reports, what supplies they carry and the state of the party's bodies. Days pass whether or not they act, and every seat keeps advancing.

### The loop

1. **Read the piece.** It reports each active seat's region and progression.
2. **Investigate.** Learn who the holder might be and which parts and weaknesses they have.
3. **Prepare.** Fetch the consumable materials the target's weaknesses call for.
4. **Descend and crawl.** Fight through the deep to the holder's region.
5. **Kill or spare.** A kill sets the seat back and passes it on.
6. **Return, heal and graft.** Recovery costs days.

### The piece of Serilmen as sensor

- It reports **region** (direction of the pull) and **progression** (strength of the pull), never who. A region can hold a candidate and several grieving false positives, and the piece can't tell them apart.
- A passed seat reads like a wounded one: the pull weakens and may shift region. Before the reveal, players read that as "we hurt it and it fled."
- It reads slightly stronger near other pieces of Serilmen, so its guidance is honest but tilted toward reassembly.

### Investigation

Investigation turns time into knowledge of a holder's parts and weaknesses. There are three sources:

- **People** in the crawlspaces: the holder's name, keepsakes, and what they lost.
- **Places:** the holder's region of the deep, where remains show which materials its creatures are made of.
- **The body:** a Sweep during a fight, or watching them from hiding, as the archon setting's "observe your prey" rule requires.

### Time pressures

- **Seats advance daily,** and their regions deepen whether or not the player visits.
- **Depth costs time:** deeper descents take longer, and the party has to get back.
- **Healing and grafting take days.** Pushing on while maimed saves time but costs actions in every fight.
- **Food and water run out below.** What the party carries sets how long a descent can last.
- **Supplies are fetched from places:** coal from dead crawlspaces, salk dust and Violet blood from the surface, lake water from underground lakes, daylight jars from the salk, lees from the thrukk, keepsakes from people.
- **The flock is weather.** The hibblejaunts are agents with needs like everything else, and where they roam decides which surface routes are open. Anyone they catch is annihilated, if the doctrine is right.
- **Export days** land irregularly, on a schedule set by a vote and an algorithm in another world. They reshape the deep overnight.

### Kills and the clock

Before the reveal, kills seem to hold the line, because each one sets a seat back. After the reveal, kills do two things: buy time to finish reassembling Serilmen, and decide who sits.

### Dreams

Sleeping in town moves time forward and brings a dream. Dreams carry lore without dumping it:

1. **One image per sleep,** a few seconds long and skippable. No narration and no explanation.
2. **Show, don't tell.** A concrete, measurable detail that something found later proves, with nothing saying they're linked. For example, a dream shows a door seven hands wide; much later, the oldest layer holds a sealed doorway exactly that size.
3. **Clues, not exposition.** Some dreams hold something usable: a name, where a keepsake lies, what a part is made of.
4. **Unreliable.** Dreams are never confirmed true, which keeps the soul doctrine and the world-is-a-dream belief ambiguous.
5. **Driven by state.** Grafted characters dream Serilmen's memories, more vividly the more grafts they carry. A recently passed seat shows up as a face the player hasn't met yet.
6. **Never spoil the reveal.** Dreams can foreshadow, but the oldest layer is where it all comes together.

## Combat

Every combatant is a body made of parts, and parts give actions. Hitting a part's weakness severs it, and severing parts is how turns are manipulated. The idea comes from Serilmen being cut to pieces, combined with Fear & Hunger's limb targeting.

### Fighting on the map

Fights happen on the same hex map and on the same clock as exploring, like a roguelike. There's no separate combat screen.

- **One timeline.** Every body, in the party or not, acts on one shared clock of ticks.
- **Parts act on their own time.** A step keeps the body busy for a time its legs set, and a swing keeps one arm busy. A body decides whenever one of its parts is free, so more arms means more swings in the air at once, and a swing can start while the legs recover from a step.
- **Some actions need more than one part.** A Sweep takes both arms and a Graft takes two hands, and some effects hit several parts or stop everything a body is doing.
- **Lead and follow.** The player moves a leader and the rest trail in single file, or controls each member directly, and can switch between the two at any time. A hostile coming into sight switches to individual control automatically, and the party goes back to following when the danger passes, unless the player chose the mode.
- **Actions come from menus.** The player positions a member on the grid, then picks the action, type and target from a menu.
- **Position matters.** An action reaches only what's in sight and in range. In a soop one cell wide only the front member can fight, and ledges and slopes give height.
- **The timeline strip.** Like the turn-order bars in many RPGs, a strip shows the upcoming ticks: an icon on each tick where something happens, +N when several things share one, and selecting a tick lists who acts and what lands.
- **Fleeing is walking away.** A body with slower legs falls behind.

### Timing

- **Wind-up and recovery.** Every action takes time to land and time to recover from. The form, the type (lake water is heavy), the part's material and grafts all change both.
- **Dodging.** A swing is aimed when it starts. If its target is out of reach or out of sight when it lands, it misses, so stepping away in time is the answer to a slow, heavy blow.
- **Hit-and-run.** Against something slower, swing, let it land, and step away while its own swing is still winding up. Against something faster, read its wind-ups on the timeline and choose your moment.
- **Staggering.** Some hits cancel the swings a body has in progress. Cutting off the arm that's winding up cancels that swing for good.

### Lethality

Combat is meant to be lethal and short. If you go in without a plan, you lose.

- **Few hits.** A part comes off in one to three ordinary hits, and in one on its weakness.
- **Blind parties lose.** Without knowing what a body is made of and carrying the types that sever it, the party wears parts down while the enemy takes theirs off.
- **The plan is the core loop.** Learn materials, fetch the right consumables, pick the ground (a soop mouth, a ledge), and time the approach.
- **Winning still costs.** Injuries persist and healing takes days, so even a won fight spends time on the seat clock.

### Bodies and timing

- **Arms (or arm-equivalents)** each act on their own time: each can be winding up or recovering from its own swing. The **head** handles targeting. The **legs** decide how fast the body steps.
- **Strange bodies have strange parts.** An archon's form reflects what it wants, so its parts are organs of its wish.
- **Each part is made of a material** with its own weakness. Hit the weakness and the part comes off in one blow; hit it any other way and you wear it down.
- **Severing manipulates timing:** cut an arm mid-swing and that swing never lands; cut the legs and every step takes longer; cut the head and it attacks at random.
- **The same rules apply to the party.** A lost arm means one fewer swing in the air at once, and injuries persist between fights.
- **Pre-entropy things don't know how to die.** A part severed from one sometimes crawls off and fights as its own combatant, so cutting can multiply enemies. A crawling part is not one of "everyone who crawls": Stand does not freeze it. A crawling part fights with whatever forms its own grants allow, like any body, and swings with the Cut type.

### Actions = form × type

An action is a character's **form** loaded with a **type**. Cut, Press, Relic and Name are innate. Every other type is a consumable material that has to be fetched.

### Types

| Type | Source | Severs | Catch |
| --- | --- | --- | --- |
| Cut | Any edge (innate) | Flesh, hair, sinew | Baseline |
| Press | The thrukk grip (Reaper, innate) | Anything hollow: skulls, shells, clay vessels | Strongest on a pinned part |
| Relic | Pieces of Serilmen (innate to grafts) | Native flesh and clay | Does nothing to old matter, which is from the same era |
| Name | Names the party has learned (innate) | Anything that was once someone | Works only on its owner, and speaking it in the deep uses it up |
| Burn | Coal from dead crawlspaces | Flesh, hair | Hardens clay, making clay parts tougher |
| Rot | Salk dust | Old matter from before the transition | Also eats the party's Serilmen grafts |
| Soak | Underground lake water | Clay, including moving clay | Heavy to haul, which slows travel |
| Light | Daylight jars | Eyes and senses of things that live in the dark | Gone once opened |
| Violet | Hibblejaunt blood | Rift-made flesh: wish-born things, archons | Collected on the salk where the flock fought |
| Keepsake | Belongings from a candidate's past | Wish-carrying parts of that candidate and anything their region breeds | Has to be found or taken from people who knew them |
| Lees | Dregs from the thrukk | Things that feed on soul | Whether it is really soul is up to the doctrine |

Press, Violet and Lees are phrased so the mechanic works whether or not the doctrine and the rift origin are true.

Name and Keepsake are never refused. Used on a body they don't belong to, the swing lands and does nothing, as an immunity does, so the menu never tells the player whose name or keepsake they hold.

### Forms

| Form | Effect |
| --- | --- |
| Strike | Hits one part hard, harder on a weakness. It never severs |
| Sweep | Hits every part of one body lightly, and reveals what each part is made of |
| Sever | The only form that cuts a part off. Cuts the part off outright on a weakness; otherwise it only wears the part, less than a Strike does |
| Pin | Locks a part so it can't act for a while, canceling any swing it was winding up. No damage, which is the safe answer to old things whose severed parts crawl off |
| Stand | Everyone crawling, on both sides, is pushed back on the timeline, and the stander's parts are all ready at once. While standing they take double damage, and allies are delayed too |
| Give | A character spends one arm's time so an ally can act at once |
| Graft | Attaches a piece of Serilmen, or puts a severed part back (a reattach, which is not a graft), mid-fight. Graftable characters only |

Example actions: lake water with Sweep douses every clay part on a body. A daylight jar with Sweep blinds every eye. A candidate's name with Pin stops the part that still remembers being him.

### Grafts

- Pieces of Serilmen are the main progression: they replace what the deep takes.
- Grafted parts sometimes take their action on their own, choosing their own target.
- They also quietly shape what the player is offered: which routes look reachable, which targets the menu suggests.
- Over a full run the party gradually becomes Serilmen, which is heresy against Salumen.

## Party

One party with two protagonists and a fixed roster of about six characters, four of whom fight at a time, with recruits filling those same slots. The two separate routes were collapsed into one party to cut the story writing.

Party characters are fixed, authored characters; none are generated.

### Protagonists in combat

- **The Reaper:** Press, Pin and Stand. He is the authority indoors.
- **The elevated salken:** Relic, Sever and Graft. They are the first character to carry a graft.

Neither protagonist is ever eligible for apotheosis.

### The fixed roster

The roster is fixed rather than freely recruited, and it has two kinds of companion:

- **Graftable humans.** They grow stranger as pieces of Serilmen replace what the deep takes.
- **Deep recruits.** They can't be grafted; each has one unique effect instead. Two sketches:
  - A severed hand of something ancient that followed the party out. It can latch onto another character for a fight, giving them an extra arm and so an extra action.
  - A remnant learning to die. It can't be killed, but every "death" costs it a part for good, so it shrinks over the course of the game.

One companion is bound to a candidate they love (see the companion beat under Candidates and seats). Because a companion's parts are the party's actions, losing them to a seat costs the party in combat too.

### Food and water

Party members need food and water, on the same needs engine as everyone else. Hunger and thirst rise with time and are met by eating rations and drinking lake water from the party's inventory, so drinking spends the same lake water that Soak uses. A member who goes without is slowed first, then loses integrity from vital parts; the thresholds are content. Starvation can kill, but slowly.

### Recruits

The protagonists can hire generic people to fight beside the fixed roster. Recruits are where the deep's brutality lands without costing the story its characters.

- **Who:** each town offers different recruits by caste and trade: guards and miners in ibben towns, field hands in the salken village, exiles on the salk.
- **Bodies:** the same part-based bodies as everyone else, so a lost arm is a lost action.
- **Needs and wants:** basic ones such as pay, food, safety and faith. Unmet needs make them complain and eventually walk off; terror makes them flee mid-fight. Both are ordinary behaviors, not a desertion system.
- **Caste friction (proposal):** ibben recruits may refuse to follow a salken, and some balk at sacrilege such as collecting Violet blood.
- **Death is permanent,** with no plot armor. The hiring pool refills only as people are born, grow up and arrive.

Recruits take one of the four fight slots; a fifth slot is something to playtest. Recruits can be grafted, which makes them the cheapest way to learn what a piece does, and the graft's pull works on them too. A grafted recruit who deserts takes the piece with them.

**Price:** there is no fixed wage or currency. Each NPC names their own price, in what they want, scaled by what they bring. Merchants work the same way.

**Paying them.** What a recruit wants is a list of item kinds from content (food, lees, Violet blood and so on). Hiring takes a bundle of those items from the party's inventory, scaled by what the recruit brings. Each day after that they want upkeep from the same list. Pay is one of their needs: as it goes unmet, complaining and then walking off score higher than staying, through the same behaviors as anyone else. Later, personality decides what a recruit does as pay goes unmet: grumble, then walk off or rob the party, and loyalty can keep one around; for now it grumbles, then walks off. A deserter lives on as an individual and is neutral to the party: whether it holds a grudge is left to the grudges and memory system (M7), one individual at a time.

### Exotic recruits

Rarer recruits from the deep can't be grafted, but each brings a special ability, and their wants are as strange as they are. Sketches:

| Recruit | Wants | Special ability |
| --- | --- | --- |
| Tamed moving clay | The dead | Takes a severed enemy part as its own, gaining that part's action |
| A leashed soul-feeder | To feed | Grows stronger near the slowly dying, on either side |
| A walking corpse from the still town | Stillness | Cannot bleed or rot, but never heals; it leaves once the still town starts to rot |

## Simulation

Apotheosis is simulated, not scripted: seven seat state machines advance on the day clock, and a small model of the cast decides who takes a seat when one passes. Hand-written candidates come first; generated ones take over when the authored pool runs out.

### Seats

- Each seat holds its stage, its progress within the stage, its current holder and its region of the deep.
- Each stage change emits an apotheosis event: the region shifts, creatures spawn, rumors spread, NPCs react. Those events tell the player what is happening, alongside the piece.
- A kill pushes progress back by a tuned amount, passes the seat to a new holder, and may move it to a different region.

### The cast model

Every resident of the crawlspaces carries a tiny model:

```
Person:  relationships  (who they love, and how much)
         losses         (simulated: deaths, reaps, exiles, exports, archon prices)
         lack           (what they never had; drives HOW a wish curdles)
         authored_wish  (optional)
```

When a seat passes, everyone is scored by how badly the kill wounded them: relationship strength to the dead holder, plus accumulated losses. The top scorer takes the seat.

Residents are full agents (see People with needs); these fields are the part of their state the seat simulation reads. When a seat passes, the offer finds the top scorer and draws them into the deep.

### Authored first, generated after

The seven starting candidates, the companion's beloved and a core cast of about 15 to 20 people have hand-written wishes. Anyone without one gets a generated wish through this pipeline:

**Wound → Wish → Curdle mode → Price and beneficiaries → Form**

- The curdle mode is chosen from the person's `lack`, not from the wording of the wish, so two people with the same wish come out differently.
- Curdle modes include stasis, diffuse cost, transfer and uncertain harm. The generator avoids repeating whichever mode it used recently.
- An archon's price runs as a simulation rule applied to the cast (a soul drain spread across everyone, children frozen in place). Beneficiaries emerge as the NPCs whose state improves, rather than being scripted.
- Forms are assembled from a library of parts by material. Combat already treats bodies as sets of parts, so generated archons are affordable, and in ASCII they are simply listed as parts.

### Deferred

Simulating a candidate's behavior once they reach a given stage is out of scope for now and kept for later.

## Technical plan

The MVP is the full game engine with almost no content: all logic is code, all content is data. That keeps content design separate from logic design, so either can be tested and tuned without touching the other.

### The engine/data boundary

Forms and archon prices have behavior, so the engine defines a small vocabulary of effect primitives and the data composes them. New types, forms and archons almost never need new code; a genuinely new primitive is a deliberate engine change.

**Code (the engine):**

- Combat resolution: parts grant actions, sever rules, Pin and Stand semantics, graft autonomy.
- Effect primitives: `damage`, `sever`, `lock`, `delay`, `ready`, `interrupt`, `add_status`, `attach_part`, `raise_max_integrity`, `heal`, `set_rest_place`, `heal`, `set_rest_place`, `reveal`, `consume`, `act_alone`, `spawn_from_part` and `step`.
- The seat state machine and stage events, the piece sensor, and investigation.
- The cast model: wound scoring, seat passing and the wish generator.
- Schedulers: the day clock and time costs, exports and the flock.

**Data (the content):**

- Materials, forms (compositions of primitives with parameters), part templates, enemy templates and regions.
- Authored candidates: wound, wish, curdle mode, price as composed effects, and form.
- Cast members, the curdle-mode library and clue definitions.
- Tuning constants: seat speed, kill setback, export rate.

Content is validated before play: records on load (every referenced material exists), effects when a game starts (every primitive is known), so errors show up as test failures rather than mid-fight crashes. Dev builds hot-reload data so a fight can be tuned without restarting.

### Architecture

The engine is a pure library with no UI types, so the frontend can be swapped later without touching it. A headless harness runs thousands of seeded days to tune the clock with data instead of guesses.

&#91;embedded content: architecture · one core, content in, four consumers\]

Only the frontends know about screens, so a later move to pixel art or Godot leaves the engine untouched.

### Language and frontend (open)

| Option | For | Against |
| --- | --- | --- |
| C# core, [SadConsole](https://www.nuget.org/packages/SadConsole/) for ASCII | Mainstream game-dev language; the core drops straight into Godot if the final game moves there; SadConsole draws its own tile-based window, so fonts, colors and input are under the game's control | Not the daily work language. Godot C# exports to desktop, Android and iOS, but [not yet to the web](https://github.com/godotengine/godot-proposals/discussions/13076) ([platform docs](https://docs.godotengine.org/es/4.3/tutorials/scripting/c_sharp/index.html)) |
| Go core, [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the terminal | The daily work language, so the fastest start; great testing for the harness; [Ebitengine](https://ebitengine.org/) offers a pure-Go path to pixel graphics later | Niche for games: fewer tutorials, libraries and people to ask, no editor. [graphics.gd](https://github.com/quaadgras/graphics.gd) bridges to Godot but is a community project |

Decision: Go, with Bubble Tea for the MVP and Ebitengine later if tiles and sprites are wanted. The layout and libraries are in the Go technical design tab.

### MVP definition

- Every system implemented: seats, cast and seat passing, the piece, investigation, combat with all forms and primitives, grafts, the day clock, the flock and exports.
- Minimal content: a handful of materials, forms and enemy templates, one crawlspace, one region of the deep, the two protagonists and one companion.
- The test: read the piece, investigate, crawl, fight, kill, and watch the seat pass. If that loop is fun with one seat, seven will work.
- Not in the MVP: the story system (scenes, casting, strain, threads, the director). It's post-MVP; see the Story tab (`docs/story.md`).

## Open questions and rejected ideas

The language choice and the clock's numbers block implementation; the rest is content work that can follow the MVP.

### Open

- [ ] Language and frontend: decided, Go with Bubble Tea, Ebitengine later.
- [ ] Clock numbers: game length in days, seat speed per stage, kill setback, export frequency. To be tuned in the headless harness.
- [ ] Four more starting candidates.
- [ ] The fixed roster: about six companions, including the one bound to a candidate.
- [ ] The deep: its regions, enemy families by material, and how export days reshape it.
- [ ] The reveal: exactly what is found in the oldest layer of the deep.
- [ ] Story rework, starting with the opening.
- [ ] Whether the slave doctrine's one-on-one challenge against a hibblejaunt appears in play.
- [ ] Whether the interface itself is the Demiurge's: honest about what you can do, quietly shaped in what it offers.
- [ ] The bond's ending: whether the piece steered it too, or it's the one thing the protagonists keep.

* [ ] Party death: can a fixed party member die for good, or only be maimed?
* [ ] Act I: is the "first god" the already-enshrined archon, or a second seat completing?

- [ ] Playtest a fifth fight slot.

### Rejected

| Idea | Why |
| --- | --- |
| "The Settling": entropy reaching one crawlspace on a countdown | Obviously ripped from Pathologic |
| Copying SMT's Press Turn | Too close to SMT; weaknesses and turn manipulation are fine as inspiration |
| Argument or cross-examination combat | Not entertaining |
| Combat only against candidates and archons | A dungeon crawler needs regular fights |
| Wishes that reduce to "kill everything that fits" | Predictable; the player just kills every suspect |
| Turning candidates back or talking them down | Ascension is one-way |
| The player shaping what a candidate becomes | Removed by design |
| The Reaper being forced to reap | The theme is the illusion of choice, not literal loss of control |
| An ending where the archons win | Players would feel cheated after trying |
| Two separate protagonist routes | Collapsed into one party to cut story writing |
| A protagonist eligible for apotheosis | Neither protagonist can ever ascend |
| Opening with a man who won't die | Didn't bring the two protagonists together |
| A separate dungeon built from each candidate's psyche | Too close to Persona 5's palaces; one shared deep instead |
