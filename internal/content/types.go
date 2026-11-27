package content

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// Bundle is the whole of content/data as decoded.
type Bundle struct {
	Materials map[string]Material  `json:"materials"`
	Terrains  map[string]Terrain   `json:"terrains"`
	Text      map[string]string    `json:"text"`
	Calendar  Calendar             `json:"calendar"`
	Templates map[string]Template  `json:"templates"`
	Party     PartyConfig          `json:"party"`
	Combat    CombatConfig         `json:"combat"`
	Timeline  TimelineConfig       `json:"timeline"`
	Actions   map[string]ActionDef `json:"actions"`
	World     WorldDef             `json:"world"`
	// Places are the hand-made pieces of the world ("place." IDs).
	Places map[string]PlaceDef `json:"places"`
	// Pieces are the building blocks towns are laid out from ("piece." IDs).
	Pieces map[string]PieceDef `json:"pieces,omitempty"`
	// Geographies are the kinds of generated ground ("geography." IDs).
	Geographies map[string]GeographyDef `json:"geographies,omitempty"`
	Start       Start                   `json:"start"`
	Claims      map[string]Claim        `json:"claims"`
	Origins     map[string]Origin       `json:"origins"`
	Knowledge   KnowledgeConfig         `json:"knowledge"`
	Agents      AgentsConfig            `json:"agents"`
	// Parts are the body part templates ("part." IDs); Bodies are the body plans ("body." IDs).
	Parts  map[string]PartTemplate `json:"parts"`
	Bodies map[string]BodyPlan     `json:"bodies"`
	// Types are the types a form can carry ("type." IDs), Items the things types burn ("item." IDs) and
	// Factions the sides things fight for ("faction." IDs).
	Types    map[string]Type    `json:"types"`
	Items    map[string]Item    `json:"items"`
	Factions map[string]Faction `json:"factions"`
	// Statuses are the conditions a body can be under for a while ("status." IDs).
	Statuses map[string]Status `json:"statuses"`
	// Rules are named formulas any formula may use as rule.<name>; the compiler inlines them.
	Rules map[string]string `json:"rules,omitempty"`
	// Choices are the choices actions may ask for, served from a collection of records.
	Choices map[string]ChoiceDef `json:"choices,omitempty"`
	// Vocabulary closes the sets of tags, grant kinds and list values content may use.
	Vocabulary Vocabulary `json:"vocabulary,omitempty"`
	// Drives are the needs and desires bodies carry ("drive." IDs).
	Drives map[string]Drive `json:"drives,omitempty"`
	// Affordances are what a landmark can give ("affordance." IDs), such as water or the dead.
	Affordances map[string]Affordance `json:"affordances,omitempty"`
}

// Data is what a record carries for rules to read. Code never reads a tag, prop or list key;
// formulas do.
type Data struct {
	// Tags are free words: "innate", "piece", "clay", ...
	Tags []string `json:"tags,omitempty"`
	// Props are named numbers: "weight": 2.
	Props map[string]float64 `json:"props,omitempty"`
	// Lists are named lists, usually of IDs: a list of weaknesses, say.
	Lists map[string][]string `json:"lists,omitempty"`
}

// Vocabulary lets content declare which tags, grant kinds and list values are allowed.
type Vocabulary struct {
	// Tags maps a collection ("items", "types", ...) to the tags its records may carry.
	Tags map[string][]string `json:"tags,omitempty"`
	// Grants lists the grant kinds parts may give. Empty allows any.
	Grants []string `json:"grants,omitempty"`
	// Lists maps "<collection>.<key>" to the collection its values must name: "materials.<list>": "types".
	Lists map[string]Collection `json:"lists,omitempty"`
}

// ChoiceDef is a choice an action can ask for, served from a collection of records.
type ChoiceDef struct {
	// From names a collection: "types", "items", ...
	From Collection `json:"from"`
	// Offer is an Expr bool over OfferEnv; empty offers every record.
	Offer string `json:"offer,omitempty"`
}

// Mods are what a record changes about an action it's chosen for: a type loaded into a form, a
// piece held for a graft. Formulas are Expr over ActionEnv.
type Mods struct {
	// Requires is an Expr bool; false refuses the action, with Refusal as the reason.
	Requires string `json:"requires,omitempty"`
	// Refusal is a text ID; empty means text.actions.requires.
	Refusal api.ContentID `json:"refusal,omitempty"`
	// Windup is ticks added to the wind-up.
	Windup string `json:"windup,omitempty"`
	// Recovery is ticks added to the recovery.
	Recovery string      `json:"recovery,omitempty"`
	OnStart  []EffectRef `json:"on_start,omitempty"`
	OnLand   []EffectRef `json:"on_land,omitempty"`
}

// Status is a condition a body is under until a tick, such as Stand's braced stance. Code knows
// nothing of any one status: formulas read the ones a body has through the statuses variable.
type Status struct {
	Data
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Triggers run for bodies carrying the status.
	Triggers []Trigger `json:"triggers,omitempty"`
}

// Claim is a kind of fact a knower can hold, such as a material claim. Each claim needs the text
// "text.knowledge.<name>" (see api.KnowledgeTextID) that reports it in the log.
type Claim struct {
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Value says what a fact's value of this claim holds; empty means a literal, shown as it is.
	Value ValueKind `json:"value,omitempty"`
}

// ValueKind says what a fact's value holds and how it is shown.
type ValueKind string

// Value kinds of a claim.
const (
	// ValueMaterial is a material ID, shown as the material's name.
	ValueMaterial ValueKind = "material"
	// ValueText is a text ID, shown rendered.
	ValueText ValueKind = "text"
)

// Affordance is something a landmark gives, which an agent plans toward when a goal needs it.
type Affordance struct {
	Data
	// Name is a text ID.
	Name api.ContentID `json:"name"`
}

// Origin is where a fact came from, such as a sweep.
type Origin struct {
	// Name is a text ID.
	Name api.ContentID `json:"name"`
}

// KnowledgeConfig is the numbers of every knower's memory.
type KnowledgeConfig struct {
	// Cap is how many facts a knower outside the controlled party keeps; past it the least
	// confident, then the oldest, is forgotten.
	Cap int `json:"cap"`
}

// AgentsConfig is the numbers and claims of agents' minds, and the numbers every agent shares.
type AgentsConfig struct {
	// Weight is an Expr over WeightEnv: an individual's weight on a drive, drawn once when it spawns
	// around its template's typical weight. Required when any template lists drives.
	Weight string `json:"weight,omitempty"`
	// SightRadius and SightDepth are how far an agent sees, in hexes and in depths.
	SightRadius int `json:"sight_radius"`
	SightDepth  int `json:"sight_depth"`
	// AtClaim is the claim ("claim." ID) an agent records for where it saw a body, and DeadClaim the
	// one for seeing it dead. Empty means that sighting is not recorded.
	AtClaim   api.ContentID `json:"at_claim,omitempty"`
	DeadClaim api.ContentID `json:"dead_claim,omitempty"`
	// SeenOrigin is the origin ("origin." ID) of what an agent sees, and SeenConfidence how sure of
	// it the agent is, from 0 to 1.
	SeenOrigin     api.ContentID `json:"seen_origin,omitempty"`
	SeenConfidence float64       `json:"seen_confidence"`
	// GivesClaim is the claim an agent records for what a landmark it sees gives, its value the
	// affordance IDs joined by commas; where the landmark is goes under AtClaim. Empty records neither.
	GivesClaim api.ContentID `json:"gives_claim,omitempty"`
	// GroundClaim is the claim an agent records for how decayed the ground is at a cell, its value the
	// world geography's decay level there: for the cell it stands in, and for the nearest open cell it
	// sees at each other level. Empty records none.
	GroundClaim api.ContentID `json:"ground_claim,omitempty"`
}

// Material is a substance things are made of.
type Material struct {
	Data
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Triggers run for parts made of the material.
	Triggers []Trigger `json:"triggers,omitempty"`
}

// Type is what a form carries into the part it lands on, such as Cut or Soak. Not every type does
// damage. Code knows nothing of any one type:
// what it needs and what it does are its data and mods.
type Type struct {
	Data
	Mods
	// Name is a text ID.
	Name api.ContentID `json:"name"`
}

// Item is something the inventory carries.
type Item struct {
	Data
	Mods
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Bound means tied to one person, e.g. a name or a keepsake.
	Bound bool `json:"bound"`
	// Part is the part template ("part." ID) an item becomes when grafted on; only such an item has one.
	Part *api.ContentID `json:"part,omitempty"`
	// Triggers run for holders of the item.
	Triggers []Trigger `json:"triggers,omitempty"`
}

// Faction is a side. Hostility is not symmetric unless both sides list each other.
type Faction struct {
	Data
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// HostileTo holds faction IDs.
	HostileTo []api.ContentID `json:"hostile_to"`
}

// StackDef is some of an item a template starts with.
type StackDef struct {
	// Item names the item by its ID.
	Item api.ContentID `json:"item"`
	// Count is at least 1.
	Count int `json:"count"`
}

// SourceDef is what a source holds and how fast it refills.
type SourceDef struct {
	// Item names the item by its ID.
	Item api.ContentID `json:"item"`
	// Max is the most stock the source holds; at least 1.
	Max int `json:"max"`
	// Refill is an Expr with RefillEnv: the stock that comes back at each phase boundary, never
	// negative. The stock never goes above Max.
	Refill string `json:"refill"`
}

// Terrain is what one cell of the world is made of: solid (clay, rock) or open (air, later water).
type Terrain struct {
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Solid blocks movement and can be stood on.
	Solid bool `json:"solid"`
	// Opaque blocks sight.
	Opaque bool `json:"opaque"`
	// FloorGlyph is drawn when you stand on it.
	FloorGlyph string `json:"floor_glyph,omitempty"`
	// WallGlyph is drawn when it bounds open space.
	WallGlyph string `json:"wall_glyph,omitempty"`
	Color     string `json:"color,omitempty"`
	// MoveCost is an Expr with MoveEnv: ticks to walk on top of it; required when solid.
	MoveCost string `json:"move_cost,omitempty"`
}

// Calendar says how ticks group into phases and days. A day is its phases in
// order, and lasts as many ticks as they add up to.
type Calendar struct {
	// Phases are in order; at least one.
	Phases []Phase `json:"phases"`
	// Exports says when the deep is exported and regenerated.
	Exports ExportSchedule `json:"exports"`
}

// ExportSchedule sets the irregular days on which the world exports.
type ExportSchedule struct {
	// FirstDay is the day of the first export, counting from 1.
	FirstDay int `json:"first_day"`
	// MinGap and MaxGap bound the days between one export and the next, both at least 1.
	MinGap int `json:"min_gap"`
	MaxGap int `json:"max_gap"`
}

// Phase is one part of the day, such as dawn.
type Phase struct {
	// ID is like "phase.dawn".
	ID api.ContentID `json:"id"`
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Ticks is how long the phase lasts; at least 1.
	Ticks int64 `json:"ticks"`
}

// Template is a bag of optional components. A nil field means "no such component".
type Template struct {
	Data
	Appearance *comp.Appearance `json:"appearance,omitempty"`
	Party      *PartyDef        `json:"party,omitempty"`
	Actor      *ActorDef        `json:"actor,omitempty"`
	// Actions lists the actions a body of this template knows, as "action." IDs.
	Actions []api.ContentID `json:"actions,omitempty"`
	// Name is the text ID of the entity's name, for display.
	Name *api.ContentID `json:"name,omitempty"`
	// Body is the body plan ("body." ID) the entity is built from.
	Body *api.ContentID `json:"body,omitempty"`
	// Protected gives the entity plot armor: it is downed rather than killed.
	Protected bool `json:"protected,omitempty"`
	// Inventory is the inventory the entity starts with.
	Inventory []StackDef `json:"inventory,omitempty"`
	// Container gives the entity an inventory from the moment it spawns, even an empty one, so it is
	// a container before anyone puts something in it.
	Container bool `json:"container,omitempty"`
	// Source makes the entity a source of one item, which it starts full of.
	Source *SourceDef `json:"source,omitempty"`
	// Triggers react to events for the entities spawned from this template.
	Triggers []Trigger `json:"triggers,omitempty"`
	// Faction is the side ("faction." ID) a body of this template fights for.
	Faction *api.ContentID `json:"faction,omitempty"`
	// Drives maps each drive ID a body of this template carries to its typical weight, which each
	// individual's own weight is drawn around.
	Drives map[string]float64 `json:"drives,omitempty"`
	// Affordances are what an entity of this template gives from the moment it spawns, as
	// "affordance." IDs, which makes it a landmark.
	Affordances []api.ContentID `json:"affordances,omitempty"`
}

// TriggerRole is which subject of an event a trigger's entity must be.
type TriggerRole string

// Roles a trigger's entity may play in the event.
const (
	RoleSource TriggerRole = "source"
	RoleTarget TriggerRole = "target"
	RolePart   TriggerRole = "part"
	RoleAny    TriggerRole = "any"
)

// triggerRoles lists the roles a trigger may name, for bundles built in Go, which skip the schema.
func triggerRoles() []TriggerRole { return []TriggerRole{RoleSource, RoleTarget, RolePart, RoleAny} }

// Trigger says "when this event happens to me, if this holds, do these effects". It lives on a
// template, material, status or item; the entities that carry that thing run it.
type Trigger struct {
	// On is an event kind, e.g. "combat.part_severed".
	On string `json:"on"`
	// If is an Expr bool over EffectEnv; empty means always.
	If string `json:"if"`
	// Do lists effects to apply, with the event's subjects as source, target and part.
	Do   []EffectRef `json:"do"`
	Role TriggerRole `json:"role"`
}

// PartyDef makes a template a party: the member templates it brings with it.
type PartyDef struct {
	// Members are member templates, in party order; the first leads.
	Members []api.ContentID `json:"members"`
}

// PartyConfig holds the numbers every party shares.
type PartyConfig struct {
	// Slots is how many fighters fight at a time.
	Slots int `json:"slots"`
	// SightRadius is in hexes.
	SightRadius int `json:"sight_radius"`
	// SightDepth is how many cells up and down the party sees.
	SightDepth int `json:"sight_depth"`
	// PlanRange is how many steps of walking from the focus the plan view draws reachable floor.
	PlanRange   int `json:"plan_range"`
	MaxStepUp   int `json:"max_step_up"`
	MaxStepDown int `json:"max_step_down"`
	MaxDrop     int `json:"max_drop"`
	// ClimbCost is an Expr with SpeedEnv: ticks added to each step up, on top of the timeline's
	// step_ticks. Climbing costs more than descending on purpose. Empty means no extra cost.
	ClimbCost string `json:"climb_cost,omitempty"`
	// IndividualOnAlert says whether an alert switches the party to individual control, and back
	// when it ends.
	IndividualOnAlert bool `json:"individual_on_alert"`
	// WakeTicks is how long a wiped-out party is out before it wakes at its last rest place.
	WakeTicks int64 `json:"wake_ticks"`
	// FollowAction is the action ("action." ID) a follower steps with; it targets a direction.
	FollowAction api.ContentID `json:"follow_action"`
}

// CombatConfig is the numbers of alert and death.
type CombatConfig struct {
	// CalmTicks is how long no hostile may be in sight before an alert ends.
	CalmTicks int64 `json:"calm_ticks"`
	// StillTicks is how long a creature with no mind of its own keeps still before it looks again.
	StillTicks int64 `json:"still_ticks"`
	// SightRadius and SightDepth are how far a member looks for hostiles, in hexes and in depths.
	SightRadius int `json:"sight_radius"`
	SightDepth  int `json:"sight_depth"`
	// CorpseGlyph and CorpseColor are how a dead body is drawn.
	CorpseGlyph string `json:"corpse_glyph"`
	CorpseColor string `json:"corpse_color"`
	// AIScore is an Expr over ScoreEnv that rates an intent for a hunting creature: the highest wins.
	AIScore string `json:"ai_score"`
	// Aims is an Expr bool over GrantEnv: whether a body can pick its target. One that cannot swings
	// at random among what it could hit.
	Aims string `json:"aims"`
	// AttackTag is the action tag that marks the actions a hunting creature attacks with.
	AttackTag string `json:"attack_tag"`
	// StepAction is the action ("action." ID) a hunting creature closes in with; it targets a direction.
	StepAction api.ContentID `json:"step_action"`
	// MaterialClaim is the claim ("claim." ID) that says what a part is made of, so the bodies view
	// can show "?" until the party has learned it. Empty means no material is ever known.
	MaterialClaim api.ContentID `json:"material_claim"`
}

// LayerDef is one depth of a place. Rows use odd-r offset layout, one character per cell; a space
// leaves the world's fill.
type LayerDef struct {
	Depth int      `json:"depth"`
	Rows  []string `json:"rows"`
}

// WorldDef is the size of the one world and what fills the cells no place covers.
type WorldDef struct {
	// Cols and Rows are hexes across, in odd-r offset, from column 0 and row 0.
	Cols int `json:"cols"`
	Rows int `json:"rows"`
	// MinDepth and MaxDepth are inclusive.
	MinDepth int `json:"min_depth"`
	MaxDepth int `json:"max_depth"`
	// Fill is the terrain of every cell no place covers: solid clay until worldgen generates the deep.
	Fill api.ContentID `json:"fill"`
	// Edge is the terrain beyond the bounds. It must be solid and opaque.
	Edge api.ContentID `json:"edge"`
	// Open is the terrain of a marked cell.
	Open api.ContentID `json:"open"`
	// Geography builds the cells no place covers ("geography." ID). Empty leaves them all Fill.
	Geography api.ContentID `json:"geography,omitempty"`
	// ChunkBudget is the most chunks kept in memory; chunks far from every party member are dropped past
	// it. 0 keeps every chunk.
	ChunkBudget int `json:"chunk_budget,omitempty"`
	// NearChunks is the radius in chunks around each party member whose chunks are never dropped.
	NearChunks int `json:"near_chunks,omitempty"`
	// PocketChunks is the radius in chunks around each party member kept as it was when an export
	// regenerates the deep.
	PocketChunks int `json:"pocket_chunks,omitempty"`
	// PocketExits is the most ways out of the pocket that the regenerated deep is joined to. A world with
	// a geography needs at least 1.
	PocketExits int `json:"pocket_exits,omitempty"`
}

// GeographyDef lists the stages that build chunks of one kind of ground, in order.
type GeographyDef struct {
	Stages []StageDef `json:"stages"`
	// Spawns is what lives on the generated ground: each entry can start a body of its template on any
	// open cell with solid ground below, within its depth band.
	Spawns []GeographySpawn `json:"spawns,omitempty"`
	// Decay changes the solid ground the stages leave, by depth and export count. Empty leaves it as it is.
	Decay *DecayDef `json:"decay,omitempty"`
}

// DecayDef says how decayed the generated ground is at a depth, and what decayed ground is made of.
type DecayDef struct {
	// Level is a formula over DecayEnv giving a whole number: 0 leaves the ground as it is, and level n
	// turns it into Terrains[n-1], or the last terrain once n passes the list.
	Level string `json:"level"`
	// Terrains are solid terrain IDs, from the least decayed to the most.
	Terrains []api.ContentID `json:"terrains"`
}

// GeographySpawn is one kind of creature a geography starts with.
type GeographySpawn struct {
	// Template is the "template." ID of the body to spawn.
	Template api.ContentID `json:"template"`
	// Weight is the template's share among the entries that apply to a cell.
	Weight float64 `json:"weight"`
	// MinDepth and MaxDepth bound the depths the entry applies at, both inclusive.
	MinDepth int `json:"min_depth"`
	MaxDepth int `json:"max_depth"`
	// Density is in [0, 1]: how likely a cell the entry applies to is to start a body, before the entry's
	// weight shares that chance with the other entries that apply there.
	Density float64 `json:"density"`
}

// StageDef names a registered worldgen stage and its numbers. Whether the stage takes those params is the
// stage's to check, against the stage registry.
type StageDef struct {
	// Stage is the registered stage's name, such as "caves".
	Stage  string             `json:"stage"`
	Params map[string]float64 `json:"params,omitempty"`
}

// PlaceDef is a piece of hand-made world, such as a town or a set piece, set at a fixed position.
type PlaceDef struct {
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// At is the world cell of the place's column 0, row 0 and depth 0; layer depths are relative to it.
	At     PlaceAt    `json:"at"`
	Layers []LayerDef `json:"layers"`
	// Legend maps one character to a terrain ID.
	Legend map[string]api.ContentID `json:"legend"`
	// Marks maps one character to a mark name, such as "@": "party_start". A mark name is unique
	// within its place, and a marked cell holds the world's Open terrain.
	Marks map[string]string `json:"marks"`
	// Spawns maps a mark character to a template ("template." ID): a body of it is spawned at each
	// such character at the start, and the character may appear any number of times (at least once).
	// Each character must also be one of Marks, so the layers may use it.
	Spawns map[string]api.ContentID `json:"spawns,omitempty"`
	// Openings lists mark characters that stand where the place opens onto the generated deep: worldgen
	// carves a tunnel from each such cell. The character may appear any number of times (at least once)
	// and must be one of Marks and not one of Spawns.
	Openings []string `json:"openings,omitempty"`
	// Affordances maps a mark character to what each cell holding it gives ("affordance." IDs), such
	// as a graveyard's dead: the cell is a landmark with no body. The character may appear any number of
	// times (at least once) and must be one of Marks and neither a spawn nor an opening.
	Affordances map[string][]api.ContentID `json:"affordances,omitempty"`
}

// PieceDef is a small building block a town is laid out from, such as a house or a well: drawn like a
// place, but with no position of its own. It is stamped into the world at an offset and a rotation.
type PieceDef struct {
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Layers are drawn with column 0, row 0 and depth 0 at the cell the piece is stamped at.
	Layers []LayerDef `json:"layers"`
	// Legend maps one character to a terrain ID.
	Legend map[string]api.ContentID `json:"legend"`
	// Marks maps one character to a mark name, unique within the piece. A marked cell holds the world's
	// Open terrain.
	Marks map[string]string `json:"marks"`
	// Spawns maps a mark character to a template ("template." ID), as a place's do: a body, container or
	// source of it starts at each such character.
	Spawns map[string]api.ContentID `json:"spawns,omitempty"`
	// Affordances maps a mark character to what each cell holding it gives ("affordance." IDs), as a
	// place's do.
	Affordances map[string][]api.ContentID `json:"affordances,omitempty"`
	// Doors lists mark characters that stand where the piece joins a street or another piece. The
	// character may appear any number of times (at least once) and must be one of Marks and neither a
	// spawn nor a landmark.
	Doors []string `json:"doors"`
}

// Layout is how a place or a piece is drawn, apart from where it stands. Openings belong to places and
// Doors to pieces; the other is empty.
type Layout struct {
	Layers      []LayerDef                 `json:"layers"`
	Legend      map[string]api.ContentID   `json:"legend"`
	Marks       map[string]string          `json:"marks"`
	Spawns      map[string]api.ContentID   `json:"spawns,omitempty"`
	Affordances map[string][]api.ContentID `json:"affordances,omitempty"`
	Openings    []string                   `json:"openings,omitempty"`
	Doors       []string                   `json:"doors,omitempty"`
}

// Layout returns the place's drawing. Its maps and slices are shared: don't change them.
func (d PlaceDef) Layout() Layout {
	return Layout{
		Layers: d.Layers, Legend: d.Legend, Marks: d.Marks, Spawns: d.Spawns,
		Affordances: d.Affordances, Openings: d.Openings,
	}
}

// Layout returns the piece's drawing. Its maps and slices are shared: don't change them.
func (d PieceDef) Layout() Layout {
	return Layout{
		Layers: d.Layers, Legend: d.Legend, Marks: d.Marks, Spawns: d.Spawns,
		Affordances: d.Affordances, Doors: d.Doors,
	}
}

// PlaceAt is a world cell in odd-r offset. Row must be even, so the place's odd rows stay odd in the world.
type PlaceAt struct {
	Col   int `json:"col"`
	Row   int `json:"row"`
	Depth int `json:"depth"`
}

// Start is the scenario a new game begins with: a party placed at a mark of a place.
type Start struct {
	Place api.ContentID `json:"place"`
	// The party's leader stands on this mark of the place; the others nearest it.
	Mark string `json:"mark"`
	// Party names the party template to spawn.
	Party api.ContentID `json:"party"`
}

// TimelineConfig holds the numbers of the shared timeline.
type TimelineConfig struct {
	// StepTicks is an Expr with SpeedEnv: base ticks a step keeps its channels busy.
	StepTicks string `json:"step_ticks"`
	// Speed is an Expr over GrantEnv: a body's speed, from the grants of its functioning parts.
	Speed string `json:"speed"`
	// DefaultSpeed is for entities whose speed nothing else provides.
	DefaultSpeed int `json:"default_speed"`
	// WindowTicks is how far ahead the timeline view looks.
	WindowTicks int64 `json:"window_ticks"`
}

// ActorDef puts a template's entities on the timeline.
type ActorDef struct {
	// Controller is "player", "follow", "idle", ...
	Controller string `json:"controller"`
}

// EffectRef is one primitive invocation. Params are Expr sources evaluated with EffectEnv, or, for
// parameters the primitive declares as selectors or text, literal strings.
type EffectRef struct {
	Do     string            `json:"do"`
	Params map[string]string `json:"params"`
}

// Collection names a collection of records: "types", "items", ...
type Collection string

// Collections that hold records.
const (
	CollectionMaterials   Collection = "materials"
	CollectionTerrains    Collection = "terrains"
	CollectionText        Collection = "text"
	CollectionTemplates   Collection = "templates"
	CollectionActions     Collection = "actions"
	CollectionPlaces      Collection = "places"
	CollectionPieces      Collection = "pieces"
	CollectionParts       Collection = "parts"
	CollectionClaims      Collection = "claims"
	CollectionOrigins     Collection = "origins"
	CollectionBodies      Collection = "bodies"
	CollectionTypes       Collection = "types"
	CollectionItems       Collection = "items"
	CollectionFactions    Collection = "factions"
	CollectionStatuses    Collection = "statuses"
	CollectionDrives      Collection = "drives"
	CollectionAffordances Collection = "affordances"
)

// TargetKind is what an action targets.
type TargetKind string

// Target kinds of an action.
const (
	TargetNone      TargetKind = "none"
	TargetSelf      TargetKind = "self"
	TargetDirection TargetKind = "direction"
	TargetBody      TargetKind = "body"
	TargetPart      TargetKind = "part"
	TargetAlly      TargetKind = "ally"
	TargetCell      TargetKind = "cell"
)

// targetKinds lists the target kinds an action may name, for bundles built in Go, which skip the
// schema.
func targetKinds() []TargetKind {
	return []TargetKind{TargetNone, TargetSelf, TargetDirection, TargetBody, TargetPart, TargetAlly, TargetCell}
}

// ActionDef is something a body does over time: a step, a swing, a graft taking both hands. It says
// which channels it occupies, how long it winds up and recovers, what it targets, and which effects
// apply when it starts and when it lands. Formulas are Expr sources over ActionEnv.
type ActionDef struct {
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Key is a short option-ID segment, e.g. "step", "strike".
	Key string `json:"key"`
	// Uses are the channels it occupies; every requirement must be met by free channels.
	Uses   []Requirement `json:"uses"`
	Target TargetKind    `json:"target"`
	// Targets narrows which bodies a body, ally or part action may be aimed at.
	Targets TargetFilter `json:"targets,omitempty"`
	// Range is in hexes, for body/part/ally/cell targets.
	Range int `json:"range,omitempty"`
	// Sight means the target must be visible from the source.
	Sight bool `json:"sight,omitempty"`
	// Choices are extra choices the player makes, by provider name, e.g. "type".
	Choices []string `json:"choices,omitempty"`
	// Requires is an Expr bool over ActionEnv.
	Requires string `json:"requires,omitempty"`
	// Windup is an Expr for ticks before it lands; 0 lands at once.
	Windup string `json:"windup"`
	// Recovery is an Expr for ticks the used channels stay busy after landing.
	Recovery string      `json:"recovery"`
	OnStart  []EffectRef `json:"on_start,omitempty"`
	OnLand   []EffectRef `json:"on_land,omitempty"`
	// Connects is an Expr bool evaluated when the action lands. Empty means "the target is still in
	// range (and in sight, if required)".
	Connects string `json:"connects,omitempty"`
	// Tags are free tags for rules and content: "attack", "movement", ...
	Tags []string `json:"tags,omitempty"`
}

// TargetFilter narrows an action's body targets by the tags on their templates. The zero filter
// allows any living body.
type TargetFilter struct {
	// Tags must all be on the target's template.
	Tags []string `json:"tags,omitempty"`
	// Not must all be absent from the target's template.
	Not []string `json:"not,omitempty"`
	// Corpse allows a dead body as well as a living one.
	Corpse bool `json:"corpse,omitempty"`
}

// PartTemplate is one kind of body part. Material and integrity are copied onto each part spawned.
type PartTemplate struct {
	Data
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	// Slot is "head", "arm", "legs", ...: free text, for display and grafting.
	Slot string `json:"slot"`
	// Material names the part's default material.
	Material  api.ContentID `json:"material"`
	Integrity int           `json:"integrity"`
	Grants    []comp.Grant  `json:"grants"`
	Vital     bool          `json:"vital,omitempty"`
	// CrawlsAs is the template ("template." ID) a body made from this part by spawn_from_part takes
	// as its identity: its name, appearance, the actions it knows and its lists. Without it the body
	// is named for the part template and knows no actions.
	CrawlsAs api.ContentID `json:"crawls_as,omitempty"`
}

// BodyPlan lists the parts a body is built from, in order.
type BodyPlan struct {
	Parts []PlanPart `json:"parts"`
}

// PlanPart is one entry of a body plan.
type PlanPart struct {
	// Part names the part template to use.
	Part api.ContentID `json:"part"`
	// Material overrides the part's own, e.g. a clay arm.
	Material api.ContentID `json:"material,omitempty"`
}

// Requirement is one set of channels an action occupies.
type Requirement struct {
	// Tag is a channel tag: "body", "arm", "legs", "grant:action", ...
	Tag string `json:"tag"`
	// Count is how many such channels; 0 means every one the body has (at least one).
	Count int `json:"count"`
}
