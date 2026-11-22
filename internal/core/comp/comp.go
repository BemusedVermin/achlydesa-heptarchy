package comp

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

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

// Party is what a whole party shares. It sits on the party entity, which has no position.
type Party struct {
	Slots  int     `json:"slots"`  // fighters at a time, copied from content at spawn
	Leader api.UID `json:"leader"` // the member who leads
}

// MemberOf is the Ark relation from a party member to its party entity.
type MemberOf struct {
	ecs.RelationMarker `json:"-"`
}

// Busy marks a channel (a body or one of its parts) busy until a tick. No component means free.
type Busy struct {
	Until int64 `json:"until"`
}

// Actor puts a body on the timeline.
type Actor struct {
	Controller string `json:"controller"` // "player", "follow", "idle", ...
	WaitUntil  int64  `json:"wait_until"` // no decision before this tick
}

// Pending is an action in flight, on an entity of its own: one per action. Whatever the action
// needs beyond this lives in a kind-specific component on the same entity.
type Pending struct {
	Kind    string  `json:"kind"`    // the resolver's name, e.g. "combat.move"
	Actor   api.UID `json:"actor"`   // the body acting
	Part    api.UID `json:"part"`    // the channel it uses (an arm), or the body
	Target  api.UID `json:"target"`  // the body it is aimed at, or 0
	Started int64   `json:"started"` // the tick it was started
	Lands   int64   `json:"lands"`   // the tick it lands
	Seq     uint64  `json:"seq"`     // scheduling order, for ties
}
