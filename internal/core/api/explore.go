package api

import "strconv"

// Event kinds of exploring. Stepping has no command of its own: it is a content action, started
// with [Do].
const (
	MovedKind          = "explore.moved"
	FellKind           = "explore.fell"
	ControlKind        = "explore.control"
	LeadKind           = "explore.lead"
	ControlChangedKind = "explore.control_changed"
	LeaderChangedKind  = "explore.leader_changed"
	PartyWokeKind      = "explore.party_woke"
)

// The modes of party control, and who may have set one.
const (
	// The leader is played; the rest trail in single file
	ControlFollow = "follow"
	// Every member's decisions come to the player
	ControlIndividual = "individual"
	ControlByPlayer   = "player"
	ControlByAlert    = "alert"
)

// Control sets how the party is controlled: a player's choice, which an alert does not override
// when it ends. It is allowed at any time.
type Control struct {
	Party UID    `json:"party"`
	Mode  string `json:"mode"`
}

// CommandKind implements [Command].
func (Control) CommandKind() string { return ControlKind }

// Lead makes a member the leader, at the front of the party order. It is allowed at any time.
type Lead struct {
	Party  UID `json:"party"`
	Member UID `json:"member"`
}

// CommandKind implements [Command].
func (Lead) CommandKind() string { return LeadKind }

// ControlChanged reports that the party's control mode changed.
type ControlChanged struct {
	Party UID `json:"party"`
	// ControlFollow or ControlIndividual
	Mode string `json:"mode"`
	// ControlByPlayer or ControlByAlert
	By string `json:"by"`
}

// EventKind implements [Event].
func (ControlChanged) EventKind() string { return ControlChangedKind }

// LogText implements [Loggable].
func (c ControlChanged) LogText() Text {
	return Text{ID: ContentID("text.explore.control_changed." + c.Mode)}
}

// LeaderChanged reports that another member leads the party.
type LeaderChanged struct {
	Party  UID `json:"party"`
	Leader UID `json:"leader"`
	// Name is the leader's rendered name, for the log.
	Name string `json:"name"`
}

// EventKind implements [Event].
func (LeaderChanged) EventKind() string { return LeaderChangedKind }

// LogText implements [Loggable].
func (l LeaderChanged) LogText() Text {
	return Text{ID: "text.explore.leader_changed", Args: map[string]string{"name": l.Name}}
}

// PartyWoke reports that a party that was wiped out got up again at its last rest place.
type PartyWoke struct {
	Party UID   `json:"party"`
	At    Cell  `json:"at"`
	Tick  int64 `json:"tick"`
}

// EventKind implements [Event].
func (PartyWoke) EventKind() string { return PartyWokeKind }

// LogText implements [Loggable].
func (PartyWoke) LogText() Text { return Text{ID: "text.explore.party_woke"} }

// FightBeat implements [Pausing].
func (PartyWoke) FightBeat() {}

// The kinds of a [Moved] event.
const (
	MoveLevel = "level"
	MoveUp    = "up"
	MoveDown  = "down"
	MoveDrop  = "drop"
)

// Moved reports that a body stepped from one cell to another.
type Moved struct {
	Source UID  `json:"source"`
	From   Cell `json:"from"`
	To     Cell `json:"to"`
	// One of MoveLevel, MoveUp, MoveDown, MoveDrop
	Kind string `json:"kind"`
}

// EventKind implements [Event].
func (Moved) EventKind() string { return MovedKind }

// Fell reports that a body dropped off a ledge.
type Fell struct {
	Body  UID `json:"body"`
	Cells int `json:"cells"`
}

// EventKind implements [Event].
func (Fell) EventKind() string { return FellKind }

// Subjects implements [Subjected]: the faller is the target.
func (f Fell) Subjects() map[string]UID { return map[string]UID{SubjectTarget: f.Body} }

// Params implements [Params]: the cells fallen, for the formulas of the triggers it fires.
func (f Fell) Params() map[string]string { return map[string]string{"cells": strconv.Itoa(f.Cells)} }

// LogText implements [Loggable].
func (f Fell) LogText() Text {
	return Text{ID: "text.explore.fell", Args: map[string]string{"cells": strconv.Itoa(f.Cells)}}
}

// MapView is the plan view of the world around the party: at most one drawn cell per hex, with no
// layers in it. Frontends draw what is here and shade by Below; they never learn a depth.
type MapView struct {
	// Sorted by R, then Q
	Cells []CellView `json:"cells"`
	// Sorted by UID
	Entities []EntityView `json:"entities"`
	// Moves has one entry per direction in Dir order, for the body the view is centered on.
	Moves []MoveView `json:"moves"`
	// Elevation is the rendered line telling how deep that body stands.
	Elevation string `json:"elevation"`
}

// MoveBlocked is the kind of a [MoveView] with no step; the others are the [Moved] kinds.
const MoveBlocked = "blocked"

// MoveView is what a step in one direction would do.
type MoveView struct {
	Dir Dir `json:"dir"`
	// One of MoveLevel, MoveUp, MoveDown, MoveDrop, MoveBlocked
	Kind string `json:"kind"`
	// Below is the destination's depth minus the body's, as for [CellView]; 0 when blocked.
	Below int `json:"below"`
	// Label is the rendered line, including the reason when blocked.
	Label string `json:"label"`
}

// CellView is one drawn hex.
type CellView struct {
	Hex   Hex    `json:"hex"`
	Glyph string `json:"glyph"`
	Color string `json:"color,omitempty"`
	// False for a remembered cell
	Visible bool `json:"visible"`
	// Below is the drawn cell's depth minus the party's: positive is lower, negative is higher.
	Below int `json:"below"`
	// Wall is true for a solid cell drawn with its wall glyph.
	Wall bool `json:"wall,omitempty"`
	// Weight is how strongly to draw the cell: -1 dim, 0 normal, 1 bold.
	Weight int `json:"weight,omitempty"`
	// Unreachable is true for a visible open cell within the plan range that the party cannot walk to.
	Unreachable bool `json:"unreachable,omitempty"`
}

// EntityView is one visible entity on the map.
type EntityView struct {
	UID   UID    `json:"uid"`
	Hex   Hex    `json:"hex"`
	Glyph string `json:"glyph"`
	Color string `json:"color,omitempty"`
	// As for [CellView]
	Below int `json:"below"`
}

// PartyView is the party the player commands.
type PartyView struct {
	UID UID `json:"uid"`
	// Hex is where to center the map: the current actor if it is a member, else the leader. It is a
	// hex, never a depth.
	Hex Hex `json:"hex"`
	// In party order
	Members []MemberView `json:"members"`
	// Control is ControlFollow or ControlIndividual.
	Control string `json:"control"`
	// Alert is whether a hostile is in sight.
	Alert bool `json:"alert"`
	// Status is the rendered lines a status bar shows: the alert while one is on, and the control
	// mode while the player decides for a member.
	Status []string `json:"status,omitempty"`
}

// MemberView is one member of the party.
type MemberView struct {
	UID     UID    `json:"uid"`
	Name    string `json:"name"`
	Glyph   string `json:"glyph"`
	Color   string `json:"color,omitempty"`
	Leader  bool   `json:"leader"`
	Current bool   `json:"current"`
}
