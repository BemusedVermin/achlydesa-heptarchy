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
	Timeline  TimelineConfig       `json:"timeline"`
	Actions   map[string]ActionDef `json:"actions"`
	Maps      map[string]MapDef    `json:"maps"`
	Start     Start                `json:"start"`
	Claims    map[string]Claim     `json:"claims"`
	Sources   map[string]Source    `json:"sources"`
	// Parts are the body part templates ("part." IDs); Bodies are the body plans ("body." IDs).
	Parts  map[string]PartTemplate `json:"parts"`
	Bodies map[string]BodyPlan     `json:"bodies"`
	// Types are the damage types ("type." IDs), Items the things types burn ("item." IDs) and
	// Factions the sides things fight for ("faction." IDs).
	Types    map[string]DamageType `json:"types"`
	Items    map[string]Item       `json:"items"`
	Factions map[string]Faction    `json:"factions"`
}

// Claim is a kind of fact a knower can hold, such as "claim.material". Each claim needs the text
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

// ValueKinds lists the non-empty value kinds a claim may name.
var ValueKinds = []string{ValueMaterial, ValueText}

// Source is where a fact came from, such as "source.sweep".
type Source struct {
	Name api.ContentID `json:"name"` // a text ID
}

// Material is a substance things are made of.
type Material struct {
	Name     api.ContentID   `json:"name"`      // a text ID
	WeakTo   []api.ContentID `json:"weak_to"`   // type IDs
	ImmuneTo []api.ContentID `json:"immune_to"` // type IDs
	Tags     []string        `json:"tags"`      // e.g. "old_matter", "clay"
}

// DamageType is a way of hurting things, such as Cut or Soak. Code knows nothing of any one type:
// only that it may need an item and may have a condition.
type DamageType struct {
	Name      api.ContentID  `json:"name"`                // a text ID
	Innate    bool           `json:"innate"`              // always available, needing no item
	Consumes  *api.ContentID `json:"consumes,omitempty"`  // item ID used up per move
	Condition string         `json:"condition,omitempty"` // Expr bool over TypeEnv; empty means always
}

// Item is something the pack carries.
type Item struct {
	Name  api.ContentID `json:"name"`  // a text ID
	Kind  string        `json:"kind"`  // one of ItemKinds
	Bound bool          `json:"bound"` // tied to one person, e.g. a name or a keepsake
}

// ItemKinds lists the kinds an item may have.
var ItemKinds = []string{"consumable", "piece", "keepsake", "name"}

// Faction is a side. Hostility is not symmetric unless both sides list each other.
type Faction struct {
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
	// Inventory is the pack the entity starts with.
	Inventory []StackDef `json:"inventory,omitempty"`
	// Later: Combatant, Needs, Inventory, Triggers, ...
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
	StepTicks    string `json:"step_ticks"`    // Expr with SpeedEnv: base ticks a step keeps its channels busy
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

// TargetKinds lists the target kinds an action may name.
var TargetKinds = []string{
	TargetNone, TargetSelf, TargetDirection, TargetBody, TargetPart, TargetAlly, TargetCell,
}

// ActionDef is something a body does over time: a step, a swing, a graft taking both hands. It says
// which channels it occupies, how long it winds up and recovers, what it targets, and which effects
// apply when it starts and when it lands. Formulas are Expr sources over ActionEnv.
type ActionDef struct {
	Name     api.ContentID `json:"name"`               // a text ID
	Key      string        `json:"key"`                // short option-ID segment, e.g. "step", "strike"
	Uses     []Requirement `json:"uses"`               // channels it occupies; every requirement must be met by free channels
	Target   string        `json:"target"`             // one of TargetKinds
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

// Grant kinds a part may give its body.
const (
	GrantAction    = "action"    // the part can carry out actions, such as an arm
	GrantTargeting = "targeting" // the part aims the body's attacks, such as a head
	GrantSpeed     = "speed"     // the part adds Value to the body's speed, such as legs
)

// GrantKinds lists the grant kinds a part may name.
var GrantKinds = []string{GrantAction, GrantTargeting, GrantSpeed}

// PartTemplate is one kind of body part. Material and integrity are copied onto each part spawned.
type PartTemplate struct {
	Name      api.ContentID `json:"name"`     // a text ID
	Slot      string        `json:"slot"`     // "head", "arm", "legs", ...: free text, for display and grafting
	Material  api.ContentID `json:"material"` // a material ID
	Integrity int           `json:"integrity"`
	Grants    []comp.Grant  `json:"grants"`
	Vital     bool          `json:"vital,omitempty"`
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
