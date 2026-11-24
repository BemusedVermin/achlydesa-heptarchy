package api

// Command is a request from the player or the harness. Implementations are plain structs.
type Command interface{ CommandKind() string }

// Event is something that happened. Implementations are plain structs.
type Event interface{ EventKind() string }

// Loggable events produce a line in the player's message log.
type Loggable interface{ LogText() Text }

// UID is the stable identity of an individual; 0 means none.
type UID uint64

// ContentID is a namespaced content ID such as "material.clay".
type ContentID string

// MapID identifies a map, hand-made or generated; 0 means none.
type MapID uint32

// Hex is an axial hex coordinate, independent of the hex library.
type Hex struct {
	Q int `json:"q"`
	R int `json:"r"`
}

// Text refers to player-facing text: a content text ID plus named arguments.
type Text struct {
	ID   ContentID         `json:"id"`
	Args map[string]string `json:"args,omitempty"`
}

// RejectedKind is the event kind of [Rejected].
const RejectedKind = "core.rejected"

// Rejected reports a command the core refused.
type Rejected struct {
	Command string `json:"command"`
	Reason  Text   `json:"reason"`
}

// EventKind implements [Event].
func (Rejected) EventKind() string { return RejectedKind }

// LogText implements [Loggable].
func (r Rejected) LogText() Text { return r.Reason }

// Roles an entity can play in a [Subjected] event.
const (
	SubjectSource = "source"
	SubjectTarget = "target"
	SubjectPart   = "part"
)

// Subjected events name the entities they are about, by role ("source", "target", "part"), so
// content triggers can find who they apply to. A role an event has no entity for is left out.
type Subjected interface{ Subjects() map[string]UID }

// Params events carry named values for the formulas of the triggers they fire: the choices of the
// action that caused them, such as "type": "type.burn", or numbers such as the cells fallen.
type Params interface{ Params() map[string]string }
