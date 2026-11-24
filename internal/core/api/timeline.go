package api

// Command and event kinds of the timeline.
const (
	WaitKind  = "timeline.wait"
	ReadyKind = "timeline.ready"
)

// Ways a [Wait] can end.
const (
	// After Ticks ticks
	WaitTicks = "ticks"
	// At the next entry on the timeline that isn't the actor's own
	WaitEvent = "event"
	// At the next phase boundary
	WaitPhase = "phase"
)

// Wait makes the current player body skip deciding until something happens.
type Wait struct {
	Source UID `json:"source"`
	// WaitTicks, WaitEvent or WaitPhase
	Until string `json:"until"`
	// For WaitTicks only; must be above 0
	Ticks int64 `json:"ticks"`
}

// CommandKind implements [Command].
func (Wait) CommandKind() string { return WaitKind }

// Ready reports that a body may decide at Tick.
type Ready struct {
	Source UID   `json:"source"`
	Tick   int64 `json:"tick"`
}

// EventKind implements [Event].
func (Ready) EventKind() string { return ReadyKind }

// Subjects implements [Subjected]: the body deciding is the source, so its parts' triggers can fire.
func (e Ready) Subjects() map[string]UID { return map[string]UID{SubjectSource: e.Source} }

// TimelineView is the shared timeline: who decides now and what comes next.
type TimelineView struct {
	Now int64 `json:"now"`
	// How many ticks past Now the entries reach
	Window int64 `json:"window"`
	// The body deciding now, or 0
	Current UID `json:"current"`
	// From Now to Now plus the window, in timeline order
	Entries []TimelineEntry `json:"entries"`
}

// TimelineEntry is one thing on the timeline.
type TimelineEntry struct {
	Tick int64 `json:"tick"`
	// EntryReady or EntryLands
	Kind   string `json:"kind"`
	Source UID    `json:"source"`
	Glyph  string `json:"glyph"`
	Color  string `json:"color"`
	// For "lands"
	Target UID `json:"target"`
	// Rendered, e.g. "The Reaper is ready"
	Label string `json:"label"`
}

// Kinds of the events for pending actions.
const (
	ActionLandedKind   = "timeline.action_landed"
	ActionCanceledKind = "timeline.action_canceled"
)

// ActionLanded reports that a pending action has landed and been resolved.
type ActionLanded struct {
	// The resolver's name
	Kind   string `json:"kind"`
	Source UID    `json:"source"`
	Target UID    `json:"target"`
	Tick   int64  `json:"tick"`
}

// EventKind implements [Event].
func (ActionLanded) EventKind() string { return ActionLandedKind }

// ActionCanceled reports that a pending action was removed before it landed.
type ActionCanceled struct {
	Kind   string `json:"kind"`
	Source UID    `json:"source"`
	Target UID    `json:"target"`
	Tick   int64  `json:"tick"`
	Reason Text   `json:"reason"`
}

// EventKind implements [Event].
func (ActionCanceled) EventKind() string { return ActionCanceledKind }
