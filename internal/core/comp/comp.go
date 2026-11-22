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
	Leader api.UID `json:"leader"` // the member who leads; always first in Order
	// Order is the party order: the leader first, then the rest. Each follower follows the member
	// just before it. Spawning fills it from the party template.
	Order []api.UID `json:"order"`
	// Alert is set by the danger check (M2-07) and is always false until then.
	Alert bool `json:"alert"`
	// Control is api.ControlFollow or api.ControlIndividual.
	Control string `json:"control"`
	// ControlBy is who set Control last: api.ControlByPlayer or api.ControlByAlert.
	ControlBy string `json:"control_by"`
	// AlertSeen is the last Alert value the party_control system acted on.
	AlertSeen bool `json:"alert_seen"`
}

// Trail is the cell a party member left on its most recent step, which the member behind it steps
// into.
type Trail struct {
	Last api.Cell `json:"last"`
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

// Landing is what an action from content needs to land, on the same entity as its Pending. It
// names every channel the action occupies, so a lock on any of them cancels it.
type Landing struct {
	Action     api.ContentID     `json:"action"`
	Channels   []api.UID         `json:"channels"`    // the channels it occupies, in UID order
	TargetPart api.UID           `json:"target_part"` // the part aimed at, or 0
	HasDir     bool              `json:"has_dir"`     // whether Dir is set
	Dir        api.Dir           `json:"dir"`
	Choices    map[string]string `json:"choices,omitempty"`
}

// Controlled marks the party the player commands.
type Controlled struct{}

// Seen is what a party remembers of the maps: cells any member has seen. Each slice is sorted by
// depth, then R, then Q, and has no repeats. Use the explore package's helpers rather than reading
// the slices, because the representation will change.
type Seen struct {
	ByMap map[api.MapID][]api.Cell `json:"by_map"`
}

// Register registers every component type of this package with the world, in a fixed order. A world
// restored from a save needs each type registered before it can read the components, and a fixed
// order keeps component IDs the same in every world.
func Register(world *ecs.World) {
	ecs.ComponentID[Pos](world)
	ecs.ComponentID[Identity](world)
	ecs.ComponentID[Appearance](world)
	ecs.ComponentID[Party](world)
	ecs.ComponentID[Trail](world)
	ecs.ComponentID[MemberOf](world)
	ecs.ComponentID[Busy](world)
	ecs.ComponentID[Actor](world)
	ecs.ComponentID[Pending](world)
	ecs.ComponentID[Landing](world)
	ecs.ComponentID[Controlled](world)
	ecs.ComponentID[Seen](world)
}
