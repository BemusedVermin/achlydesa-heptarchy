package api

import "strconv"

// Event kinds of the clock.
const (
	PhaseStartedKind = "clock.phase_started"
	DayStartedKind   = "clock.day_started"
)

// PhaseStarted reports that a new day phase began.
type PhaseStarted struct {
	Day   int       `json:"day"`
	Phase ContentID `json:"phase"`
}

// EventKind implements [Event].
func (PhaseStarted) EventKind() string { return PhaseStartedKind }

// LogText implements [Loggable]. Each phase has its own text, "text.clock.<phase id>".
func (p PhaseStarted) LogText() Text {
	return Text{ID: "text.clock." + p.Phase, Args: map[string]string{"day": strconv.Itoa(p.Day)}}
}

// DayStarted reports that a new day began.
type DayStarted struct {
	Day int `json:"day"`
}

// EventKind implements [Event].
func (DayStarted) EventKind() string { return DayStartedKind }

// LogText implements [Loggable].
func (d DayStarted) LogText() Text {
	return Text{ID: "text.clock.day_started", Args: map[string]string{"day": strconv.Itoa(d.Day)}}
}

// ClockView is the time of day. Day counts from 1.
type ClockView struct {
	Day         int    `json:"day"`
	Phase       string `json:"phase"` // display-ready phase name
	TickInPhase int64  `json:"tick_in_phase"`
	PhaseTicks  int64  `json:"phase_ticks"` // length of the current phase
}
