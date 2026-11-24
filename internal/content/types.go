package content

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// Bundle is the whole of content/data as decoded. Later tasks add fields.
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
	Maps      map[string]MapDef    `json:"maps"`
	Start     Start                `json:"start"`
	Claims    map[string]Claim     `json:"claims"`
	Origins   map[string]Origin    `json:"origins"`
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
}

// Data is what a record carries for rules to read. Code never reads a tag, prop or list key;
// formulas do.
type Data struct {
	Tags  []string            `json:"tags,omitempty"`  // free words: "innate", "piece", "clay", ...
	Props map[string]float64  `json:"props,omitempty"` // named numbers: "weight": 2
	Lists map[string][]string `json:"lists,omitempty"` // named lists, usually of IDs: a list of weaknesses, say
}

// Vocabulary lets content declare which tags, grant kinds and list values are allowed.
type Vocabulary struct {
	// Tags maps a collection ("items", "types", ...) to the tags its records may carry.
	Tags map[string][]string `json:"tags,omitempty"`
	// Grants lists the grant kinds parts may give. Empty allows any.
	Grants []string `json:"grants,omitempty"`
	// Lists maps "<collection>.<key>" to the collection its values must name: "materials.<list>": "types".
	Lists map[string]string `json:"lists,omitempty"`
}

// ChoiceDef is a choice an action can ask for, served from a collection of records.
type ChoiceDef struct {
	From  string `json:"from"`            // a collection: "types", "items", ...
	Offer string `json:"offer,omitempty"` // Expr bool over OfferEnv; empty offers every record
}

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

// Status is a condition a body is under until a tick, such as Stand's braced stance. Code knows
// nothing of any one status: formulas read the ones a body has through the statuses variable.
type Status struct {
	Data
	Name     api.ContentID `json:"name"`               // a text ID
	Triggers []Trigger     `json:"triggers,omitempty"` // run for bodies carrying the status
}

// Claim is a kind of fact a knower can hold, such as a material claim. Each claim needs the text
// "text.knowledge.<name>" (see api.KnowledgeTextID) that reports it in the log.
type Claim struct {
	Name api.ContentID `json:"name"` // a text ID
	// Value says what a fact's value of this claim holds: "material" (a material ID, shown as its
	// name), "text" (a text ID, shown rendered) or empty (a literal, shown as it is).
	Value string `json:"value,omitempty"`
}

// Value kinds of a claim.
const (
	ValueMaterial = "material"
	ValueText     = "text"
)

// valueKinds lists the non-empty value kinds a claim may name.
func valueKinds() []string { return []string{ValueMaterial, ValueText} }

// Origin is where a fact came from, such as a sweep.
type Origin struct {
	Name api.ContentID `json:"name"` // a text ID
}

// Material is a substance things are made of.
type Material struct {
	Data
	Name api.ContentID `json:"name"` // a text ID
	// Triggers run for parts made of the material.
	Triggers []Trigger `json:"triggers,omitempty"`
}

// Type is what a form carries into the part it lands on, such as Cut or Soak. Not every type does
// damage. Code knows nothing of any one type:
// what it needs and what it does are its data and mods.
type Type struct {
	Data
	Mods
	Name api.ContentID `json:"name"` // a text ID
}

// Item is something the inventory carries.
type Item struct {
	Data
	Mods
	Name  api.ContentID `json:"name"`            // a text ID
	Bound bool          `json:"bound,omitempty"` // tied to one person, e.g. a name or a keepsake
	// Part is the part template ("part." ID) an item becomes when grafted on; only such an item has one.
	Part *api.ContentID `json:"part,omitempty"`
	// Triggers run for holders of the item.
	Triggers []Trigger `json:"triggers,omitempty"`
}

// Faction is a side. Hostility is not symmetric unless both sides list each other.
type Faction struct {
	Data
	Name      api.ContentID   `json:"name"`       // a text ID
	HostileTo []api.ContentID `json:"hostile_to"` // faction IDs
}

// StackDef is some of an item a template starts with.
type StackDef struct {
	Item  api.ContentID `json:"item"`  // an item ID
	Count int           `json:"count"` // at least 1
}

// Terrain is what one cell of the world is made of: solid (clay, rock) or open (air, later water).
type Terrain struct {
	Name       api.ContentID `json:"name"`
	Solid      bool          `json:"solid"`                 // blocks movement; can be stood on
	Opaque     bool          `json:"opaque"`                // blocks sight
	FloorGlyph string        `json:"floor_glyph,omitempty"` // drawn when you stand on it
	WallGlyph  string        `json:"wall_glyph,omitempty"`  // drawn when it bounds open space
	Color      string        `json:"color,omitempty"`
	MoveCost   string        `json:"move_cost,omitempty"` // Expr with MoveEnv: ticks to walk on top of it; required when solid
}

// Calendar says how ticks group into phases and days. A day is its phases in
// order, and lasts as many ticks as they add up to.
type Calendar struct {
	Phases []Phase `json:"phases"` // in order; at least one
}

// Phase is one part of the day, such as dawn.
type Phase struct {
	ID    api.ContentID `json:"id"`    // "phase.dawn"
	Name  api.ContentID `json:"name"`  // a text ID
	Ticks int64         `json:"ticks"` // how long the phase lasts; at least 1
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
	// Triggers react to events for the entities spawned from this template.
	Triggers []Trigger `json:"triggers,omitempty"`
	// Faction is the side ("faction." ID) a body of this template fights for.
	Faction *api.ContentID `json:"faction,omitempty"`
	// Later: Combatant, Needs, ...
}

// Roles a trigger's entity may play in the event.
const (
	RoleSource = "source"
	RoleTarget = "target"
	RolePart   = "part"
	RoleAny    = "any"
)

// triggerRoles lists the roles a trigger may name.
func triggerRoles() []string { return []string{RoleSource, RoleTarget, RolePart, RoleAny} }

// Trigger says "when this event happens to me, if this holds, do these effects". It lives on a
// template, material, status or item; the entities that carry that thing run it.
type Trigger struct {
	On   string      `json:"on"`   // an event kind, e.g. "combat.part_severed"
	If   string      `json:"if"`   // Expr bool over EffectEnv; empty means always
	Do   []EffectRef `json:"do"`   // effects to apply, with the event's subjects as source, target and part
	Role string      `json:"role"` // which subject of the event this entity must be: one of the Role constants
}

// PartyDef makes a template a party: the member templates it brings with it.
type PartyDef struct {
	Members []api.ContentID `json:"members"` // member templates, in party order; the first leads
}

// PartyConfig holds the numbers every party shares.
type PartyConfig struct {
	Slots       int `json:"slots"`        // fighters at a time; 4 to start
	SightRadius int `json:"sight_radius"` // hexes; 6 to start
	// SightDepth is how many cells up and down the party sees; 6 to start.
	SightDepth int `json:"sight_depth"`
	// PlanRange is how many steps of walking from the focus the plan view looks for reachable floor,
	// by the party's step rules; 12 to start. Reachable floor is drawn over the walls and ledges
	// that share its hex.
	PlanRange int `json:"plan_range"`
	// MaxStepUp, MaxStepDown and MaxDrop are the hexmap step rules: 1, 1 and 8 to start.
	MaxStepUp   int `json:"max_step_up"`
	MaxStepDown int `json:"max_step_down"`
	MaxDrop     int `json:"max_drop"`
	// ClimbCost is an Expr with SpeedEnv: ticks added to each step up, on top of the timeline's
	// step_ticks. Climbing costs more than descending on purpose. Empty means no extra cost.
	ClimbCost string `json:"climb_cost,omitempty"`
	// IndividualOnAlert says whether an alert switches the party to individual control, and back
	// when it ends. True to start.
	IndividualOnAlert bool `json:"individual_on_alert"`
	// WakeTicks is how long a wiped-out party is out before it wakes at its last rest place.
	WakeTicks int64 `json:"wake_ticks"`
	// FollowAction is the action ("action." ID) a follower steps with; it targets a direction.
	FollowAction api.ContentID `json:"follow_action"`
}

// CombatConfig is the numbers of alert and death. Placeholders to tune.
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

// MapDef is a hand-made map: a few authored layers embedded in endless fill.
type MapDef struct {
	Name api.ContentID `json:"name"` // a text ID
	// Fill is the terrain of every cell not authored, forever in every direction.
	Fill api.ContentID `json:"fill"`
	// Open is the terrain of a marked cell, usually air.
	Open   api.ContentID `json:"open"`
	Layers []LayerDef    `json:"layers"`
	// Legend maps one character to a terrain ID.
	Legend map[string]api.ContentID `json:"legend"`
	// Marks maps one character to a mark name, such as "@": "party_start".
	Marks map[string]string `json:"marks"`
	// Spawns maps a mark character to a template ("template." ID): a body of it is spawned at each
	// such character at the start, in row order, and the character may appear any number of times
	// (at least once). Each character must also be one of Marks, so the layers may use it.
	Spawns map[string]api.ContentID `json:"spawns,omitempty"`
}

// LayerDef is one depth of a hand-made map. Rows use odd-r offset layout, one character per
// cell; a space leaves the fill.
type LayerDef struct {
	Depth int      `json:"depth"`
	Rows  []string `json:"rows"`
}

// Start is the scenario a new game begins with: a party placed at a mark on a map.
type Start struct {
	Map         api.ContentID `json:"map"`
	Party       api.ContentID `json:"party"`         // a party template ID
	PartyAtMark string        `json:"party_at_mark"` // a mark name of the map
}

// TimelineConfig holds the numbers of the shared timeline.
type TimelineConfig struct {
	StepTicks string `json:"step_ticks"` // Expr with SpeedEnv: base ticks a step keeps its channels busy
	// Speed is an Expr over GrantEnv: a body's speed, from the grants of its functioning parts.
	Speed        string `json:"speed"`
	DefaultSpeed int    `json:"default_speed"` // for entities whose speed nothing else provides
	WindowTicks  int64  `json:"window_ticks"`  // how far ahead the timeline view looks
}

// ActorDef puts a template's entities on the timeline.
type ActorDef struct {
	Controller string `json:"controller"` // "player", "follow", "idle", ...
}

// EffectRef is one primitive invocation. Params are Expr sources evaluated with EffectEnv, or, for
// parameters the primitive declares as selectors or text, literal strings.
type EffectRef struct {
	Do     string            `json:"do"`
	Params map[string]string `json:"params"`
}

// Target kinds of an action.
const (
	TargetNone      = "none"
	TargetSelf      = "self"
	TargetDirection = "direction"
	TargetBody      = "body"
	TargetPart      = "part"
	TargetAlly      = "ally"
	TargetCell      = "cell"
)

// targetKinds lists the target kinds an action may name.
func targetKinds() []string {
	return []string{TargetNone, TargetSelf, TargetDirection, TargetBody, TargetPart, TargetAlly, TargetCell}
}

// ActionDef is something a body does over time: a step, a swing, a graft taking both hands. It says
// which channels it occupies, how long it winds up and recovers, what it targets, and which effects
// apply when it starts and when it lands. Formulas are Expr sources over ActionEnv.
type ActionDef struct {
	Name     api.ContentID `json:"name"`               // a text ID
	Key      string        `json:"key"`                // short option-ID segment, e.g. "step", "strike"
	Uses     []Requirement `json:"uses"`               // channels it occupies; every requirement must be met by free channels
	Target   string        `json:"target"`             // one of the Target constants
	Range    int           `json:"range,omitempty"`    // hexes, for body/part/ally/cell targets
	Sight    bool          `json:"sight,omitempty"`    // the target must be visible from the source
	Choices  []string      `json:"choices,omitempty"`  // extra choices the player makes, by provider name, e.g. "type" (M2)
	Requires string        `json:"requires,omitempty"` // Expr bool over ActionEnv
	Windup   string        `json:"windup"`             // Expr ticks before it lands; 0 lands at once
	Recovery string        `json:"recovery"`           // Expr ticks the used channels stay busy after landing
	OnStart  []EffectRef   `json:"on_start,omitempty"`
	OnLand   []EffectRef   `json:"on_land,omitempty"`
	// Connects is an Expr bool evaluated when the action lands. Empty means "the target is still in
	// range (and in sight, if required)".
	Connects string   `json:"connects,omitempty"`
	Tags     []string `json:"tags,omitempty"` // free tags for rules and content: "attack", "movement", ...
}

// PartTemplate is one kind of body part. Material and integrity are copied onto each part spawned.
type PartTemplate struct {
	Data
	Name      api.ContentID `json:"name"`     // a text ID
	Slot      string        `json:"slot"`     // "head", "arm", "legs", ...: free text, for display and grafting
	Material  api.ContentID `json:"material"` // a material ID
	Integrity int           `json:"integrity"`
	Grants    []comp.Grant  `json:"grants,omitempty"`
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
	Part     api.ContentID `json:"part"`               // a part template ID
	Material api.ContentID `json:"material,omitempty"` // override, e.g. a clay arm
}

// Requirement is one set of channels an action occupies.
type Requirement struct {
	Tag   string `json:"tag"`   // a channel tag: "body", "arm", "legs", "grant:action", ...
	Count int    `json:"count"` // how many such channels; 0 means every one the body has (at least one)
}
