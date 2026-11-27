package api

// View is what a frontend draws, in display-ready strings; a nil section is absent.
type View struct {
	Tick     int64         `json:"tick"`
	Options  []Option      `json:"options"`
	Log      []LogLine     `json:"log"`
	Clock    *ClockView    `json:"clock,omitempty"`
	Timeline *TimelineView `json:"timeline,omitempty"`
	Map      *MapView      `json:"map,omitempty"`
	Party    *PartyView    `json:"party,omitempty"`
	Bodies   *BodiesView   `json:"bodies,omitempty"`
	Wipe     *WipeView     `json:"wipe,omitempty"`
	Labels   LabelsView    `json:"labels"`
}

// LabelsView is the rendered headings a frontend puts on its panels.
type LabelsView struct {
	// Actions heads the action menu.
	Actions string `json:"actions"`
	// Look heads the look panel when a body is under the cursor, and LookEmpty replaces it when none is.
	Look      string `json:"look"`
	LookEmpty string `json:"look_empty"`
}

// Option is one thing the player can do right now. Frontends render Label and
// send Command when it is chosen; they never build commands themselves.
type Option struct {
	// ID is stable within a view, e.g. "step.nw"; slash-separated for menus,
	// e.g. "act/strike/cut".
	ID      string  `json:"id"`
	Label   string  `json:"label"`
	Command Command `json:"-"`
	Enabled bool    `json:"enabled"`
	// Reason says why the option is disabled.
	Reason string `json:"reason,omitempty"`
	// Path names each slash-separated segment of ID, rendered, so a frontend can label the groups
	// of a menu built from the IDs. It may be shorter than the ID, or empty, for an option that is
	// not grouped.
	Path []string `json:"path,omitempty"`
	// Confirm is the rendered question to ask before choosing the option; empty when it needs none.
	Confirm string `json:"confirm,omitempty"`
}

// LogLine is one line of the player's message log.
type LogLine struct {
	// Seq counts every line the log has ever held, so a frontend can tell which lines are new.
	Seq  int64  `json:"seq"`
	Tick int64  `json:"tick"`
	Text string `json:"text"`
	// Pause marks a fight beat: a frontend stops on it so the player can read it.
	Pause bool `json:"pause,omitempty"`
}
