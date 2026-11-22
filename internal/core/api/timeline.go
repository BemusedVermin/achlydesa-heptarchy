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
	Actor UID    `json:"actor"`
	Until string `json:"until"` // WaitTicks, WaitEvent or WaitPhase
	Ticks int64  `json:"ticks"` // for WaitTicks only; must be above 0
}

// CommandKind implements [Command].
func (Wait) CommandKind() string { return WaitKind }

// Ready reports that a body may decide at Tick.
type Ready struct {
	Actor UID   `json:"actor"`
	Tick  int64 `json:"tick"`
}

// EventKind implements [Event].
func (Ready) EventKind() string { return ReadyKind }

// TimelineView is the shared timeline: who decides now and what comes next.
type TimelineView struct {
	Now     int64           `json:"now"`
	Current UID             `json:"current"` // the body deciding now, or 0
	Entries []TimelineEntry `json:"entries"` // from Now to Now plus the window, in timeline order
}

// TimelineEntry is one thing on the timeline.
type TimelineEntry struct {
	Tick   int64  `json:"tick"`
	Kind   string `json:"kind"` // "ready" now; "lands" from M1-18
	Actor  UID    `json:"actor"`
	Glyph  string `json:"glyph"`
	Color  string `json:"color"`
	Target UID    `json:"target"` // for "lands"
	Label  string `json:"label"`  // rendered, e.g. "The Reaper is ready"
}
