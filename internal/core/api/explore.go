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
	ControlFollow     = "follow"     // the leader is played; the rest trail in single file
	ControlIndividual = "individual" // every member's decisions come to the player
	ControlByPlayer   = "player"
	ControlByAlert    = "alert"
)

// Control sets how the party is controlled: a player's choice, which an alert does not override
// when it ends. It is allowed at any time.
type Control struct {
	Party UID    `json:"party"`
	Mode  string `json:"mode"` // ControlFollow or ControlIndividual
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
	Party UID    `json:"party"`
	Mode  string `json:"mode"` // ControlFollow or ControlIndividual
	By    string `json:"by"`   // ControlByPlayer or ControlByAlert
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
	Map   MapID `json:"map"`
	At    Cell  `json:"at"`
	Tick  int64 `json:"tick"`
}

// EventKind implements [Event].
func (PartyWoke) EventKind() string { return PartyWokeKind }

// LogText implements [Loggable].
func (PartyWoke) LogText() Text { return Text{ID: "text.explore.party_woke"} }

// The kinds of a [Moved] event.
const (
	MoveLevel = "level"
	MoveUp    = "up"
	MoveDown  = "down"
	MoveDrop  = "drop"
)

// Moved reports that a body stepped from one cell to another.
type Moved struct {
	Source UID    `json:"source"`
	Map    MapID  `json:"map"`
	From   Cell   `json:"from"`
	To     Cell   `json:"to"`
	Kind   string `json:"kind"` // one of MoveLevel, MoveUp, MoveDown, MoveDrop
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

// MapView is the plan view of the map the party is on: at most one drawn cell per hex, with no
// layers in it. Frontends draw what is here and shade by Below; they never learn a depth.
type MapView struct {
	Map      MapID        `json:"map"`
	Name     string       `json:"name"`
	Cells    []CellView   `json:"cells"`    // sorted by R, then Q
	Entities []EntityView `json:"entities"` // sorted by UID
}

// CellView is one drawn hex.
type CellView struct {
	Hex     Hex    `json:"hex"`
	Glyph   string `json:"glyph"`
	Color   string `json:"color,omitempty"`
	Visible bool   `json:"visible"` // false for a remembered cell
	// Below is the drawn cell's depth minus the party's: positive is lower, negative is higher.
	Below int `json:"below"`
}

// EntityView is one visible entity on the map.
type EntityView struct {
	UID   UID    `json:"uid"`
	Hex   Hex    `json:"hex"`
	Glyph string `json:"glyph"`
	Color string `json:"color,omitempty"`
	Below int    `json:"below"` // as for [CellView]
}

// PartyView is the party the player commands.
type PartyView struct {
	UID UID   `json:"uid"`
	Map MapID `json:"map"`
	// Hex is where to center the map: the current actor if it is a member, else the leader. It is a
	// hex, never a depth.
	Hex     Hex          `json:"hex"`
	Members []MemberView `json:"members"` // in party order
	// Control is ControlFollow or ControlIndividual.
	Control string `json:"control"`
	// Alert is whether a hostile is in sight.
	Alert bool `json:"alert"`
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
