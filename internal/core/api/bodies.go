package api

// BodiesView is what the party can see of every body in sight: their parts, what is known of them,
// which parts are busy, and what is about to land on whom.
type BodiesView struct {
	// Current is the body deciding now, or 0.
	Current UID `json:"current"`
	// Bodies is every body in sight and every member of the party, by UID.
	Bodies []BodyView `json:"bodies"`
	// Threats is every pending action aimed at a member of the party, by landing tick.
	Threats []ThreatView `json:"threats"`
}

// BodyView is one body.
type BodyView struct {
	UID  UID    `json:"uid"`
	Name string `json:"name"`
	// Rendered faction name, or empty when it has none
	Faction string `json:"faction,omitempty"`
	// Member is whether the body belongs to the party.
	Member bool `json:"member"`
	// Hostile is whether the body is an enemy of the party.
	Hostile  bool     `json:"hostile"`
	Statuses []string `json:"statuses,omitempty"`
	// Its parts, by UID
	Channels []ChannelView `json:"channels"`
}

// ChannelView is one part of a body, which acts on its own.
type ChannelView struct {
	UID          UID    `json:"uid"`
	Name         string `json:"name"`
	Slot         string `json:"slot"`
	Integrity    int    `json:"integrity"`
	MaxIntegrity int    `json:"max_integrity"`
	Graft        bool   `json:"graft"`
	// Material is the rendered name of what the part is made of, or "?" when the party does not
	// know. A member's own parts are always known.
	Material string `json:"material"`
	// FreeIn is the ticks until the part is free, 0 when it is free.
	FreeIn int64 `json:"free_in"`
	// Locked is whether a lock is the reason it is busy.
	Locked bool `json:"locked,omitempty"`
	// State says, rendered, whether the part is ready, busy or locked, and whether it is a graft.
	State string `json:"state"`
	// Windup is what the part is winding up, if anything.
	Windup *WindupView `json:"windup,omitempty"`
}

// WindupView is an action in flight that uses a part.
type WindupView struct {
	Action string `json:"action"`
	// The body aimed at, or 0
	Target UID `json:"target"`
	// Ticks until it lands
	Lands int64 `json:"lands"`
	// When says, rendered, the tick it lands.
	When string `json:"when"`
}

// ThreatView is a pending action aimed at a member of the party.
type ThreatView struct {
	Attacker   UID    `json:"attacker"`
	Target     UID    `json:"target"`
	TargetPart UID    `json:"target_part,omitempty"`
	Action     string `json:"action"`
	// Ticks until it lands
	Lands int64 `json:"lands"`
	// The tick it lands
	Tick int64 `json:"tick"`
}
