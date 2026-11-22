package api

// Command and event kinds of the timeline.
const (
	EndKind          = "turns.end"
	WaitKind         = "turns.wait"
	TurnStartedKind  = "turns.turn_started"
	TurnEndedKind    = "turns.turn_ended"
	UpcomingTurnsMax = 8
)

// End ends the current actor's turn.
type End struct {
	Actor UID `json:"actor"`
}

// CommandKind implements [Command].
func (End) CommandKind() string { return EndKind }

// Wait lets Ticks pass before the actor's next turn, and ends the turn.
type Wait struct {
	Actor UID   `json:"actor"`
	Ticks int64 `json:"ticks"`
}

// CommandKind implements [Command].
func (Wait) CommandKind() string { return WaitKind }

// TurnStarted reports that an actor's turn began at Tick.
type TurnStarted struct {
	Actor UID   `json:"actor"`
	Tick  int64 `json:"tick"`
}

// EventKind implements [Event].
func (TurnStarted) EventKind() string { return TurnStartedKind }

// TurnEnded reports that an actor's turn ended.
type TurnEnded struct {
	Actor UID `json:"actor"`
}

// EventKind implements [Event].
func (TurnEnded) EventKind() string { return TurnEndedKind }

// TurnsView is the timeline: whose turn it is and who acts next.
type TurnsView struct {
	Current  UID   `json:"current"`  // 0 when no turn is open
	Upcoming []UID `json:"upcoming"` // the next actors in timeline order, at most UpcomingTurnsMax
}
