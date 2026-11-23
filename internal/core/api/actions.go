package api

// Command and event kinds of content-defined actions.
const (
	DoKind            = "actions.do"
	ActionStartedKind = "actions.started"
	ActionMissedKind  = "actions.missed"
)

// Do makes the current player body start one of its actions. The options of the view carry ready
// built Do commands; a frontend sends one of those.
type Do struct {
	Actor      UID               `json:"actor"`
	Action     ContentID         `json:"action"`
	Channels   []UID             `json:"channels,omitempty"` // the limbs to use; empty lets the core pick the lowest UIDs
	Target     UID               `json:"target,omitempty"`   // the body aimed at, or 0
	TargetPart UID               `json:"target_part,omitempty"`
	Dir        *Dir              `json:"dir,omitempty"`
	Choices    map[string]string `json:"choices,omitempty"` // extra choices by provider name, e.g. "type"
}

// CommandKind implements [Command].
func (Do) CommandKind() string { return DoKind }

// ActionStarted reports that a body began an action that lands later.
type ActionStarted struct {
	Actor  UID       `json:"actor"`
	Action ContentID `json:"action"`
	Target UID       `json:"target"`
	Tick   int64     `json:"tick"`
	Lands  int64     `json:"lands"`
	// Autonomous says a grafted part started the action on its own, not its host's controller.
	Autonomous bool `json:"autonomous,omitempty"`
}

// EventKind implements [Event].
func (ActionStarted) EventKind() string { return ActionStartedKind }

// ActionMissed reports that an action landed and did nothing: its target was out of reach, or one of
// its effects could no longer apply.
type ActionMissed struct {
	Actor  UID       `json:"actor"`
	Action ContentID `json:"action"`
	Target UID       `json:"target"`
	Tick   int64     `json:"tick"`
	// Line is the log text, such as "Clay-thing's Strike hits empty air."
	Line Text `json:"line"`
	// Reason says why, when an effect's check gave one.
	Reason Text `json:"reason,omitempty"`
}

// EventKind implements [Event].
func (ActionMissed) EventKind() string { return ActionMissedKind }

// LogText implements [Loggable].
func (e ActionMissed) LogText() Text { return e.Line }
