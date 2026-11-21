package comp

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Pos is where an entity is.
type Pos struct {
	Map api.MapID `json:"map"`
	At  api.Cell  `json:"at"`
}

// Identity is what makes an entity an individual.
type Identity struct {
	UID      api.UID       `json:"uid"`
	Template api.ContentID `json:"template"`
}

// Appearance is how an entity is drawn.
type Appearance struct {
	Glyph  string `json:"glyph"`            // one grapheme, for the terminal
	Color  string `json:"color,omitempty"`  // a color name or #rrggbb
	Sprite string `json:"sprite,omitempty"` // for the graphics frontend later
}
