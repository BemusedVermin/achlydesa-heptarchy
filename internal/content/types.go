package content

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Bundle is the whole of content/data as decoded. Later tasks add fields.
type Bundle struct {
	Materials map[string]Material `json:"materials"`
	Terrains  map[string]Terrain  `json:"terrains"`
	Text      map[string]string   `json:"text"`
	Calendar  Calendar            `json:"calendar"`
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

// Calendar says how ticks group into phases and days.
type Calendar struct {
	TicksPerPhase int64   `json:"ticks_per_phase"`
	Phases        []Phase `json:"phases"` // in order; at least one
}

// Phase is one part of the day, such as dawn.
type Phase struct {
	ID   api.ContentID `json:"id"`   // "phase.dawn"
	Name api.ContentID `json:"name"` // a text ID
}
