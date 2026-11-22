package api

// Kinds of the events the timing primitives publish.
const (
	ChannelsLockedKind     = "effects.channels_locked"
	ActionsInterruptedKind = "effects.actions_interrupted"
	ChannelsDelayedKind    = "effects.channels_delayed"
	ChannelsReadiedKind    = "effects.channels_readied"
)

// ChannelsLocked reports that channels were kept busy for longer, and what they were winding up
// was canceled.
type ChannelsLocked struct {
	Channels []UID `json:"channels"` // the channels' entity UIDs, in UID order
	Ticks    int64 `json:"ticks"`
	Tick     int64 `json:"tick"`
}

// EventKind implements [Event].
func (ChannelsLocked) EventKind() string { return ChannelsLockedKind }

// ActionsInterrupted reports that pending actions were canceled without locking anything.
type ActionsInterrupted struct {
	Actors []UID `json:"actors"` // the bodies whose actions were canceled, in UID order
	Tick   int64 `json:"tick"`
}

// EventKind implements [Event].
func (ActionsInterrupted) EventKind() string { return ActionsInterruptedKind }

// ChannelsDelayed reports that channels' busy time was pushed back.
type ChannelsDelayed struct {
	Channels []UID `json:"channels"`
	Ticks    int64 `json:"ticks"`
	Tick     int64 `json:"tick"`
}

// EventKind implements [Event].
func (ChannelsDelayed) EventKind() string { return ChannelsDelayedKind }

// ChannelsReadied reports that busy channels were freed at once.
type ChannelsReadied struct {
	Channels []UID `json:"channels"`
	Tick     int64 `json:"tick"`
}

// EventKind implements [Event].
func (ChannelsReadied) EventKind() string { return ChannelsReadiedKind }
