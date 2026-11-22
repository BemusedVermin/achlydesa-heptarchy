package content

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// Bundle is the whole of content/data as decoded. Later tasks add fields.
type Bundle struct {
	Materials map[string]Material `json:"materials"`
	Terrains  map[string]Terrain  `json:"terrains"`
	Text      map[string]string   `json:"text"`
	Calendar  Calendar            `json:"calendar"`
	Templates map[string]Template `json:"templates"`
	Party     PartyConfig         `json:"party"`
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
