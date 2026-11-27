package api

// WakeKind is the kind of the Wake command.
const WakeKind = "game.wake"

// OptionWake is the ID of the option that wakes a wiped party.
const OptionWake = "wake"

// Wake gets a wiped party up at its last rest place. It is allowed only while the party is wiped.
type Wake struct{}

// CommandKind implements [Command].
func (Wake) CommandKind() string { return WakeKind }

// WipeView is the wipe screen: present in the view while the controlled party is wiped out.
type WipeView struct {
	Title string   `json:"title"`
	Lines []string `json:"lines"`
	// CanWake is false for a party with nobody downed: nothing is left to do.
	CanWake bool `json:"can_wake"`
}
