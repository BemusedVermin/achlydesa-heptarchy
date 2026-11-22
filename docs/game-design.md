# Achlydesa: The Dreaming Heptarchy — Game Design Document

Nov 20, 2026 · @Liam McGuire

## Overview

Achlydesa: The Dreaming Heptarchy is a turn-based dungeon crawler with time management about the illusion of choice. People on the clay plain who have lost something are ascending into archons in the deep. A Reaper and an elevated salken hunt the candidates, and learn late that there are always seven seats, so their kills only ever chose who sits.

- **Touchstones:** Fear & Hunger for unsettling lore and limb-level brutality. SMT and Final Fantasy for weakness-driven, turn-manipulating combat, as inspiration only, not a copy. Pathologic for open exploration, time pressure and failure that continues the story.
- **Tone:** surreal, weird and unsettling throughout.
- **Core theme:** the illusion of choice, not literal loss of autonomy. Moment-to-moment choices (combat, routes, supplies, timing) are fully real. The illusion lives in what those choices add up to, and the characters are not content with it.
- **Genre:** a dungeon crawler where combat is the backbone: regular fights, not one-off boss encounters. Menu-based, with no grid and no range bands.
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
| Requests | Find someone who went below, bring a body back for the thrukk | Quests, and the cast's relationships |

Access matters: a salken can only enter ibben towns through the lunn, the way the dead come in, while the Reaper's standing opens doors. Who is in the party decides who will talk.

## The living deep

The deep is generated around the fixed towns and populated by monsters and people with needs, not by scripted events. Emergent play is the priority: when something walks into a town, it's because what it needs is there.

### Procgen and exports

- **Fixed:** the towns, the salk, the shallow soops, and set pieces such as the oldest layer where the reveal happens.
- **Generated:** everything else, stitched to fixed gates at every town so there is no seam between town and dungeon.
- **Reseeding:** each export regenerates the generated deep from a seed derived from the world seed and the export count, so any run can be replayed.
- **Seats** keep their stage and progress; their regions land somewhere new.
- **The map goes stale.** What the player charted no longer exists.
- **The deep changes character over the game.** Each export fills the clay filter, so the upper deep grows more decayed and old matter retreats downward.
- **Guarantees:** every town connects to the deep, every active seat is reachable, and materials appear where they belong (coal near dead crawlspaces, moving clay near the thrukk, old matter deep).

### The hex grid

Exploring and fighting both happen on the hex grid, on the same clock; moves are picked from menus.

- **Why hexes suit the deep:** six even directions make organic caves and tunnels rather than corridors and rooms, and distances are uniform for creature movement.
- **ASCII handles it:** each hex is one glyph, two columns wide, with alternate rows offset by one column. Terminal cells are about twice as tall as they are wide, so the hexes come out roughly even.
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
- **Behaviors:** work (hauling water, mining coal, farming the fields, pressing the dead), trade, worship, attend town events, call the Reaper for the slowly dying, flee, migrate, and go below.
- **Roles and customs act as constraints:** ibben and salken, jobs, the lunn, standing at attention, and the test at ten sorting children.

What falls out of it:

- **Candidates leave a trail.** When a seat passes, the offer finds the most wounded person and draws them below. Noticing who slips out of town at night is investigation.
- **Rumors travel and drift.** News moves only when people carry it, between towns and through the deep, and it distorts with each retelling. Drifted rumors are where many false positives come from.
- **Quests come from unmet needs.** A resident who can't meet a need asks the party: find a missing son, bring water, carry a body to the thrukk. There is no quest list: a request is a story like any other (see Story → Emergent stories).
- **Towns can empty.** If water fails or monsters keep coming, people migrate to other towns, and a town can be abandoned.
- **Traders cross the deep.** When a trader dies on the way, the next town runs short.
- **Plot armor is allowed.** Authored characters can be flagged as protected in their data; everyone else can die, and the seat simulation adapts.

### Three tiers of people

| Tier | Who | Simulation | Death |
| --- | --- | --- | --- |
| Party | The two protagonists and the fixed roster | Authored, never generated | Protected where plot armor is set |
| Named cast | The authored core cast and generated candidates | Full needs agents with relationships, rumors and wounds | Permanent; plot armor optional per character |
| Generic | Guards, merchants, recruits and similar | Very basic needs and wants | Replaced on a timer while the town is still inhabited |

Generic NPCs keep towns working without much simulation cost. An abandoned town stops replacing them, so emptied towns stay empty.

### Grudges and memory

Survivors remember, and their grudges come from the memory and needs systems already in the game.

- **Encounters are recorded.** Every agent that survives the party remembers who maimed it, who fled and who left it behind.
- **Bodies remember.** A creature that lost an arm comes back with whatever replaced it. One that lost an arm to Cut might return with a clay arm, weak to Soak instead.
- **Revenge is a need.** A grudge becomes a need in the same engine, so a nemesis hunts the party the way moving clay hunts the dead.
- **Deserters persist.** A recruit who fled keeps living, with a grudge or gratitude, possibly carrying a piece of Serilmen.
- **The deepest nemesis is a seat.** An abandoned recruit is a wounded person, and wounded people take passed seats. The archon who comes for the party can be someone it left behind.
- **Stories drift.** Rumors about the party and each nemesis grow in the retelling, and change how towns receive the party.

### Keeping it affordable

- **Off-screen:** creatures and residents are simulated coarsely at the region or town level; their needs tick and they move between regions.
- **On-screen:** near the party, they become full agents with real movement and fights.
- **Harness logging:** the headless harness logs emergent events (deaths, migrations, towns emptied), so degenerate outcomes such as every monster camping one town show up in tuning rather than in play.

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

### How the story runs

The story is a set of scenes that world state unlocks, not a sequence of quest steps. The simulation decides who, where and when; scenes give those moments words, faces and choices.

- **Scenes** are short authored units: who is in them (roles), when they can play (a condition over world state), where they play (a slot: resting in town, arriving somewhere, after a kill, visiting a household, or a specific event), their tone (heavy or soft), their lines, the player's choices, and what they change.
- **The spine** is a handful of scenes keyed to seat stages, kills, places reached and facts learned: the opening, the first god, the midpoint, the reveal, the last seven and the door. The goal never changes; the route is the player's.
- **Arcs** are chains of scenes for each authored character and for the protagonists' bond, keyed to what they've lived through, how close they are, and what the deep has done to their bodies.
- **Set pieces.** The biggest moments are staged: before the scene, the director brings the cast together and pauses the clock. The key acts stay playable: the reap in the opening and the kill at the midpoint are the player's own actions, through the normal menus.
- **The director** picks one scene per slot: spine scenes first, then open threads, then stories found in the simulation, then everyday scenes. It alternates heavy and soft, so a hard scene is followed by room to recover, and it nudges the simulation when the spine needs it, for example in who a seat's offer finds. In the fiction, the director is Serilmen.
- **Surviving the simulation.** Anyone can die, so spine scenes cast roles, not names ("the first god" is whichever seat completes first), and every spine scene has a fallback or casts only characters with plot armor. The harness checks that the spine completes on every seed.

### Story beats (draft)

**These specific stories are subject to change.** Every beat below is a working draft, not a commitment: any of them may be rewritten, replaced or cut as the story is reworked. What's settled is the shape around them: scenes driven by world state, heavy and soft beats alternating, and the theme they serve.

Heavy and soft beats alternate. Soft beats build the warmth the heavy ones break, and they give the player room to recover; without them the world is only grim. Every character here is an adult unless stated otherwise, and none of it is graphic: the weight is in what people are made to agree to.

**The main story**

- **The first kill.** Whoever went below is still themselves: lucid, and grateful to feel well for the first time in years. They ask the Reaper why he's here. Afterward the salken says the piece has gone quiet.
- **The first god.** The Reaper has been dreaming of a woman's face and tells the salken so. When the god walks out onto the salk, he says her name: she's the widow of a man he reaped last spring. Her worshippers tell him she always called the Reaper kind.
- **The reveal.** The salken works it out first: "We never killed one. We passed them all." The Reaper turns on them, because their piece led him to every one of those people. It's the worst fight they have.
- **The last seven.** Each archon the party let sit speaks to them before the fight. Several are the women below.
- **The door.** Before they part, the salken admits the piece pulled them toward him on the first day. He says he doesn't care why they met.

**The bride of the Osten (the companion beat)**

- **The Osten.** The winner takes a young woman from the losing crawlspace as his slave wife. She was engaged to one of the losers, and her family doesn't fight it, because fighting means war. The man she was engaged to joins the party later as the companion.
- **The household.** She crawls whenever her husband stands, and he stands all the time. He beats her where the household can see; the town treats it as his right, and so does she when asked.
- **"Tell them I'm dying."** She asks the Reaper to declare her slowly dying so he can give her a quick death, the only way out the doctrine allows. Lying makes him a murderer by his own doctrine; refusing leaves her where she is; taking her onto the salk makes her a field slave under the flock; killing the husband gets her sent to the fields anyway.
- **The midpoint.** A seat finds her, the most wounded person in reach. Kill her and the seat passes to the companion, who has now lost her twice, and the Reaper recognizes his fever. Spare her and she sits, and the companion fights on knowing what she'll become.

**The woman called a vampire**

- **The sanction.** She recovered from the despair after her first child, so her crawlspace says she stole soul. She's pregnant again, and the elders and the Skadders rule that a vampire's child would be born of stolen soul and must be ended before she's cast onto the salk. They need the Reaper to stand in the room so it's lawful. He can refuse; then it happens with him on his knees, unlawful, and nobody stops it.
- **The shelter.** On the surface she shelters the party from the flock and tells them about both her children. She asks the salken whether slaves are allowed to grieve.
- **Her wish.** A seat finds her. Mothers in her old crawlspace stop losing soul to their children, and it's taken from everyone nearby instead. Her accusers sicken first and beg the party to kill her. She asks whether she's what they said she was now.

**The pregnant woman**

- **No father's household.** The father of her child is someone the party killed. To keep the child out of the fields, she must marry whoever will claim it, and the only man who will is one she fears. The Reaper can't claim it, and in the salken village the child would be born a slave. She decides in front of the party and asks them to leave before the wedding. A later visit shows the life she chose.

**The fields**

- **A mercy for a slave.** A field slave asks the Reaper to reap her before the flock comes, because death by hibblejaunt annihilates the soul. Reapers serve the ibben, and the other slaves object: someone has to work her row in the open. The salken worked that row once.
- **The elevated one.** The field slaves expect the salken to challenge a hibblejaunt, as the doctrine says the elevated can. When they don't, their old field partner, whose child the flock took while the salken was gone, calls them a false god.
- **The orphan's crawlspace.** Its children haven't grown in a year, so none reaches the test at ten. The mothers know who the candidate is and hide him. If he's killed, the test resumes and a girl the party knows is sent to the fields; the salken walks her up to the salk and tells her the truth when she asks.

**The spent mother**

- **No soul to save.** A woman lost in despair after childbirth is held to have a spent soul, so the Reaper is forbidden to reap her, however she suffers. Her husband begs him anyway. If he reaps her, the town says he wasted a reaping on something soulless; if she recovers, she'll be called a vampire.

**The Reaper**

- **The archon who thanks him.** The enshrined archon greets him by name and thanks him for the mercy he gave its mother. He walks out without a word, and the salken finds him later sitting in the thrukk.
- **"You were supposed to come."** The candidate whose wish ends painful dying is a man whose wife died slowly while the Reaper was at another call. He tells the Reaper he doesn't need him anymore, and soon no one will.
- **The remnant.** A thing that has never died asks the Reaper to teach it how to die well. At the end it asks him to reap it, and he does. He tells the salken it's the first one he was sure about.

**The two of them**

- **Crawling past.** In a town that hates salken, the ibben spit on the salken as they come through the lunn, and the Reaper walks past without stopping it, because he needs those people to talk. That night the salken asks whether he would have stopped it for an ibben.
- **Standing.** Later, grafted, the salken stands indoors, and the room has to stand for a slave. The next day the party learns the town took it out on a salken servant.
- **Cut it out.** When the salken learns the piece steered their choice to leave the fields, they ask the Reaper to cut it out of them. He refuses, and they don't speak until the next town.

**People from the simulation**

- **The deserter.** A recruit the party left behind is a candidate. His first words when they find him are "You left me."

**Soft beats**

- **Crawling lessons.** The Reaper teaches the salken ibben customs: when to crawl, when to stand at attention. They keep standing at the wrong moment and making the whole room stand. It pays off when the salken stands on purpose.
- **The first dundie.** The salken farmed dundies all their life and never ate one cooked; the Reaper has never eaten one raw. They trade, and both are disappointed.
- **"You came back for me."** The first time one carries the other out of the deep, the one carried says it that night, and the other doesn't know how to answer. It comes before Crawling past, so that one stings.
- **The bride's laugh.** The companion tells how he got engaged: he lost a dundie-eating contest on purpose because she was keeping score. On a visit to her household, she hears he's with the party and laughs, the only time in her thread.
- **The remnant's firsts.** Its first scar, its first time tired, its first sleep. It wakes from its first dream and asks, very seriously, whether that happens every night.
- **The hand.** The severed hand behaves like a stray: it sleeps on whoever is warmest, steals food, and holds the Reaper's hand while he sleeps. Everyone pretends not to notice.
- **"Is the world a dream?"** On a night's rest the party argues over the exiles' belief that this world is a dream you wake from at death. The remnant, which has never died, can't decide which way to vote.
- **A wedding people want.** Two people marry because they want to. A Reaper at a wedding is bad luck, so he waits outside while the salken learns to dance on all fours. It sits near the forced marriages, so the player sees what those women were denied.
- **The false alarm.** A reaping call turns out to be an old woman with a bad cold who called by mistake. She's mortified and makes him tea, and he stays an hour.
- **The boy who passed.** A boy passes the test at ten and his family throws a party. The salken goes, and is glad for him.
- **The recruit who made it.** A recruit who joined to earn lake water for his mother survives and brings it home, and the party is invited to dinner.
- **The still town.** Wounds don't fester there, and for the first time the party sleeps without keeping watch, until the town starts to rot.
- **The partner's seed.** After calling them a false god, the salken's old field partner finds the party and gives the salken a dundie seed from their row: "Come back alive."

### Emergent stories

Emergent stories use the same scenes as authored ones; the difference is casting. An authored scene has its cast fixed, and an emergent scene is cast from whoever in the simulation fits. The writing carries the weight in both, dark or light; the simulation supplies who, why and what came before.

- **Characters** carry traits (setting-specific ones such as devout, dutiful and proud, plus generic ones such as kind, cruel and fearful), beliefs, relationships of a kind (loves, owns, owes, fears, blames), the cast model's losses and lack, hooks (specific facts to hang a story on, such as "lost his betrothed at the Osten"), and strain. Traits and hooks are content. Traits change only through a scene that shows the change.
- **The chronicle** records the events that matter, each with its causes: who killed whom, who was taken, who was left behind, who was reaped and who grieved. Lines can quote it: "the man who took her at the Osten".
- **Casting keeps people in character.**
  - **Fit or silence.** Each role has requirements and a fit score over traits and hooks. If nobody fits well enough, the scene doesn't play: no scene is better than a miscast one.
  - **Against type only as a break.** A scene that makes someone act against their nature has to declare it, and it opens a thread about it.
  - **Characters decide in character.** When someone other than the player chooses inside a scene, each option is weighted by their traits. A kind man almost never picks the cruel option; when fear, duty or custom pushes him to, it costs him strain and opens a thread.
- **Strain** measures how far the world has forced someone from who they are, and on the plain that's constant. Acting against your traits, by choice or under custom, adds it; party members gain it too when the player makes them act against their nature. Past a threshold, a turn scene plays and their traits shift. Strain counts toward the wounds that decide who a seat finds.
- **Threads** make sure what happens is followed up. A break, a death that leaves people grieving, a promise, a betrayal or a grudge opens a thread. The director prefers scenes that move open threads, and a thread closes when a scene resolves it; if the person is gone, it closes with a scene where someone remembers.
- **Story patterns** watch the chronicle for shapes: a recruit the party abandoned takes a seat; someone reaped leaves a spouse who goes below; a maimed creature comes back. A partial match can foreshadow (a rumor about the deserter); a complete match casts and plays its scene.
- **Scenes without the party** resolve through the characters' own choices, with no text. They enter the chronicle and reach the player later as rumors, drifting as they travel.
- **Authored characters go through the same system,** so an emergent scene that casts the Reaper speaks in his voice, and authored characters can rule themselves out of roles (the Reaper is never anyone's lover in an emergent scene). Authored scenes read emergent threads, and emergent scenes read authored facts.
- **Requests are stories too.** An unmet need opens a thread, and a scene puts the ask to the party.

An example: an Osten happens in a crawlspace the party has never visited, unscripted. If the winner is cruel, the household scenes cast him as he is and open a thread about the woman he took; weeks later the generic "ask to be reaped" scene casts her, and its lines name him and that day. If the winner is kind, that scene never casts him, because he doesn't fit. A different one plays: custom forces a kind man to keep a slave wife, he gains strain, and his thread is about what he does with it. Either way it's the same theme, and nobody acts out of character.

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
- **Supplies are fetched from places:** coal from dead crawlspaces, salk dust and Violet blood from the surface, lake water from underground lakes, daylight jars from the salk, lees from the thrukk, keepsakes from people.
- **The flock is weather.** It moves across the salk on its own schedule and decides which surface routes are open. Anyone it catches is annihilated, if the doctrine is right.
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
- **Moves come from menus.** The player positions a member on the grid, then picks the action, type and target from a menu.
- **Position matters.** A move reaches only what's in sight and in range. In a soop one cell wide only the front member can fight, and ledges and slopes give height.
- **The timeline strip.** Like the turn-order bars in many RPGs, a strip shows the upcoming ticks: an icon on each tick where something happens, +N when several things share one, and selecting a tick lists who acts and what lands.
- **Fleeing is walking away.** A body with slower legs falls behind.

### Timing

- **Wind-up and recovery.** Every move takes time to land and time to recover from. The form, the type (lake water is heavy), the part's material and grafts all change both.
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
- **Pre-entropy things don't know how to die.** A part severed from one sometimes crawls off and fights as its own combatant, so cutting can multiply enemies.

### Moves = form × type

A move is a character's **form** loaded with a **type**. Cut, Press, Relic and Name are innate. Every other type is a consumable material that has to be fetched.

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

### Forms

| Form | Effect |
| --- | --- |
| Strike | Hits one part at full force |
| Sweep | Hits every part of one body lightly, and reveals what each part is made of |
| Sever | Weak in general, but removes the part outright on a weakness |
| Pin | Locks a part so it can't act for a while, canceling any swing it was winding up. No damage, which is the safe answer to old things whose severed parts crawl off |
| Stand | Everyone crawling, on both sides, is pushed back on the timeline, and the stander's parts are all ready at once. While standing they take double damage, and allies are delayed too |
| Give | A character spends one arm's time so an ally can act at once |
| Graft | Attaches a piece of Serilmen, or reattaches a severed part, mid-fight. Graftable characters only |

Example moves: lake water with Sweep douses every clay part on a body. A daylight jar with Sweep blinds every eye. A candidate's name with Pin stops the part that still remembers being him.

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

### Recruits

The protagonists can hire generic people to fight beside the fixed roster. Recruits are where the deep's brutality lands without costing the story its characters.

- **Who:** each town offers different recruits by caste and trade: guards and miners in ibben towns, field hands in the salken village, exiles on the salk.
- **Bodies:** the same part-based bodies as everyone else, so a lost arm is a lost action.
- **Needs and wants:** basic ones such as pay, food, safety and faith. Unmet needs mean grumbling, then desertion; terror means flight mid-fight.
- **Caste friction (proposal):** ibben recruits may refuse to follow a salken, and some balk at sacrilege such as collecting Violet blood.
- **Death is permanent,** with no plot armor. Towns replace their generic people on a timer while still inhabited, so the hiring pool refills.

Recruits take one of the four fight slots; a fifth slot is something to playtest. Recruits can be grafted, which makes them the cheapest way to learn what a piece does, and the graft's pull works on them too. A grafted recruit who deserts takes the piece with them.

**Price:** there is no fixed wage or currency. Each NPC names their own price, in what they want, scaled by what they bring. Merchants work the same way.

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
         traits         (who they are; content-defined, see Emergent stories)
         strain         (how far they've been forced to act against their traits)
         authored_wish  (optional)
```

When a seat passes, everyone is scored by how badly the kill wounded them: relationship strength to the dead holder, plus accumulated losses and strain. The top scorer takes the seat.

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
- Effect primitives such as `damage`, `sever_check`, `lock_part`, `freeze_others`, `transfer_action`, `attach_part`, `drain_stat`, `freeze_trait`.
- The seat state machine and stage events, the piece sensor, and investigation.
- The cast model: wound scoring, seat passing and the wish generator.
- Schedulers: the day clock and time costs, exports and the flock.

**Data (the content):**

- Materials, forms (compositions of primitives with parameters), part templates, enemy templates and regions.
- Authored candidates: wound, wish, curdle mode, price as composed effects, and form.
- Cast members, the curdle-mode library and clue definitions.
- Tuning constants: seat speed, kill setback, export rate.

Content is validated on load (every referenced material exists, every primitive is known), so errors show up as test failures rather than mid-fight crashes. Dev builds hot-reload data so a fight can be tuned without restarting.

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
