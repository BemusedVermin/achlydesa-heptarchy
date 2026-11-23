package api

// Command and event kinds of the timeline.
const (
	WaitKind  = "timeline.wait"
	ReadyKind = "timeline.ready"
)

// Ways a [Wait] can end.
const (
	WaitTicks = "ticks" // after Ticks ticks
	WaitEvent = "event" // at the next entry on the timeline that isn't the actor's own
	WaitPhase = "phase" // at the next phase boundary
)

// Wait makes the current player body skip deciding until something happens.
type Wait struct {
	Source UID    `json:"source"`
	Until  string `json:"until"` // WaitTicks, WaitEvent or WaitPhase
	Ticks  int64  `json:"ticks"` // for WaitTicks only; must be above 0
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
	Now     int64           `json:"now"`
	Window  int64           `json:"window"`  // how many ticks past Now the entries reach
	Current UID             `json:"current"` // the body deciding now, or 0
	Entries []TimelineEntry `json:"entries"` // from Now to Now plus the window, in timeline order
}

// TimelineEntry is one thing on the timeline.
type TimelineEntry struct {
	Tick   int64  `json:"tick"`
	Kind   string `json:"kind"` // "ready" or "lands"
	Source UID    `json:"source"`
	Glyph  string `json:"glyph"`
	Color  string `json:"color"`
	Target UID    `json:"target"` // for "lands"
	Label  string `json:"label"`  // rendered, e.g. "The Reaper is ready"
}

// Kinds of the events for pending actions.
const (
	ActionLandedKind   = "timeline.action_landed"
	ActionCanceledKind = "timeline.action_canceled"
)

// ActionLanded reports that a pending action has landed and been resolved.
type ActionLanded struct {
	Kind   string `json:"kind"` // the resolver's name
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
