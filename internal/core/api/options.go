package api

import (
	"strconv"
	"strings"
)

// The kinds of timeline entry, in TimelineEntry.Kind.
const (
	// A body is about to decide
	EntryReady = "ready"
	// An action in flight is about to land
	EntryLands = "lands"
)

// Option IDs the core builds and a frontend must recognise. An ID is a slash-separated path.
const (
	// Followed by a direction name
	OptionStepPrefix = "step/"
	// Followed by a control mode
	OptionControlPrefix = "control/"
	// Followed by the UID of the member to lead with
	OptionLeadPrefix = "lead/"
	OptionWaitEvent  = "wait/event"
	OptionWaitPhase  = "wait/phase"
)

// StepOption is the ID of the option that steps in the direction, named as in Dir.String.
func StepOption(dir string) string { return OptionStepPrefix + dir }

// ControlOption is the ID of the option that sets the party's control mode.
func ControlOption(mode string) string { return OptionControlPrefix + mode }

// LeadOption is the ID of the option that makes the member with the UID the leader.
func LeadOption(member UID) string { return OptionLeadPrefix + strconv.FormatUint(uint64(member), 10) }

// ParseLeadOption returns the member UID of a lead option ID, and false for any other ID.
func ParseLeadOption(id string) (UID, bool) {
	rest, ok := strings.CutPrefix(id, OptionLeadPrefix)
	if !ok {
		return 0, false
	}
	uid, err := strconv.ParseUint(rest, 10, 64)
	if err != nil {
		return 0, false
	}
	return UID(uid), true
}

// Reject builds the event for a refused command.
func Reject(kind string, reason ContentID) Rejected {
	return Rejected{Command: kind, Reason: Text{ID: reason}}
}
