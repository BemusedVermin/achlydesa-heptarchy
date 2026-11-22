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
}

// Material is a substance things are made of.
type Material struct {
	Name api.ContentID `json:"name"` // a text ID
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
	// Later: Body, Combatant, Needs, Inventory, Triggers, ...
}

// PartyDef makes a template a party: the member templates it brings with it.
type PartyDef struct {
	Members []api.ContentID `json:"members"` // member templates, in party order; the first leads
}

// PartyConfig holds the numbers every party shares.
type PartyConfig struct {
	Slots       int `json:"slots"`        // fighters at a time; 4 to start
	SightRadius int `json:"sight_radius"` // hexes; 6 to start
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

// Requirement is one set of channels an action occupies.
type Requirement struct {
	Tag   string `json:"tag"`   // a channel tag: "body", "arm", "legs", "grant:action", ...
	Count int    `json:"count"` // how many such channels; 0 means every one the body has (at least one)
}
