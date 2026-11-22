package api

// View is everything a frontend needs to draw the current state. It holds
// display-ready strings; the core resolves text before building it. Later tasks
// add a section each as a pointer field: Clock (M1-08), Timeline (M1-17), Map and Party (M1-10),
// Combat (M2).
type View struct {
	Tick     int64         `json:"tick"`
	Options  []Option      `json:"options"`
	Log      []LogLine     `json:"log"`
	Clock    *ClockView    `json:"clock,omitempty"`
	Timeline *TimelineView `json:"timeline,omitempty"`
	Map      *MapView      `json:"map,omitempty"`
	Party    *PartyView    `json:"party,omitempty"`
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
}

// LogLine is one line of the player's message log.
type LogLine struct {
	Tick int64  `json:"tick"`
	Text string `json:"text"`
}
