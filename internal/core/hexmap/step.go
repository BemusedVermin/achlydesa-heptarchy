package hexmap

import (
	"errors"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// StepRules bound how far a body can climb or fall in one step.
type StepRules struct{ MaxStepUp, MaxStepDown, MaxDrop int }

// StepKind says how a step changed depth.
type StepKind uint8

// The kinds of step.
const (
	StepLevel StepKind = iota
	StepUp
	StepDown
	StepDrop
)

// Step is a legal move from one cell.
type Step struct {
	To   Link
	Kind StepKind
	// Cells fallen, for StepDrop
	Fall int
}

// ErrBadDir is returned by [Map.Step] for a direction that is not one of the six.
var ErrBadDir = errors.New("hexmap: invalid direction")

// Step finds where a body at from ends up moving one hex in dir, if anywhere. Bodies are one cell tall.
func (m *Map) Step(from api.Cell, dir api.Dir, rules StepRules) (Step, bool, error) {
	if !dir.Valid() {
		return Step{}, false, ErrBadDir
	}
	link := m.topo.Neighbor(from, dir)
	n := link.Cell
	t, err := m.At(n)
	if err != nil {
		return Step{}, false, err
	}
	if t.Solid {
		return m.stepUp(from, n, rules)
	}
	standable, err := m.Standable(n)
	if err != nil {
		return Step{}, false, err
	}
	if standable {
		return Step{To: link, Kind: StepLevel}, true, nil
	}
	return m.stepDown(n, rules)
}

// stepUp climbs onto a standable cell above the solid cell n, needing headroom above from.
func (m *Map) stepUp(from, n api.Cell, rules StepRules) (Step, bool, error) {
	link, _, err := m.climb(from, n, rules.MaxStepUp)
	if err != nil || link == nil {
		return Step{}, false, err
	}
	return Step{To: *link, Kind: StepUp}, true, nil
}

// climb finds the lowest standable cell above the solid cell n, at most limit cells up, needing
// headroom above from. It returns nil when there is none, and how many cells up it found otherwise.
func (m *Map) climb(from, n api.Cell, limit int) (*Link, int, error) {
	above, head := n, from
	for k := 1; k <= limit; k++ {
		aboveLink, headLink := m.topo.Above(above), m.topo.Above(head)
		above, head = aboveLink.Cell, headLink.Cell
		blocked, err := m.solid(headLink)
		if err != nil || blocked {
			// No headroom, and climbing higher needs even more
			return nil, 0, err
		}
		standable, err := m.Standable(above)
		if err != nil {
			return nil, 0, err
		}
		if standable {
			return &aboveLink, k, nil
		}
	}
	return nil, 0, nil
}

// stepDown finds the first standable cell below the open cell n, at most MaxDrop cells down.
func (m *Map) stepDown(n api.Cell, rules StepRules) (Step, bool, error) {
	cur := n
	for fall := 1; fall <= rules.MaxDrop; fall++ {
		link := m.topo.Below(cur)
		cur = link.Cell
		standable, err := m.Standable(cur)
		if err != nil {
			return Step{}, false, err
		}
		if !standable {
			continue
		}
		if fall <= rules.MaxStepDown {
			return Step{To: link, Kind: StepDown}, true, nil
		}
		return Step{To: link, Kind: StepDrop, Fall: fall}, true, nil
	}
	return Step{}, false, nil
}

// RefusalKind says why [Map.Step] offers no step.
type RefusalKind uint8

// The kinds of refusal.
const (
	// RefusedSolid is a wall, or a ledge with no floor just above the climb limit.
	RefusedSolid RefusalKind = iota
	// RefusedTooHigh is a ledge one cell above what can be climbed.
	RefusedTooHigh
	// RefusedTooFar is open ground with no floor within MaxDrop.
	RefusedTooFar
)

// Refusal is why a step is not offered.
type Refusal struct {
	Kind RefusalKind
	// Height is how many cells up the ledge is, for RefusedTooHigh.
	Height int
}

// Refusal explains why [Map.Step] offers no step from from in dir. Call it only after Step
// reported none.
func (m *Map) Refusal(from api.Cell, dir api.Dir, rules StepRules) (Refusal, error) {
	if !dir.Valid() {
		return Refusal{}, ErrBadDir
	}
	n := m.topo.Neighbor(from, dir).Cell
	t, err := m.At(n)
	if err != nil {
		return Refusal{}, err
	}
	if !t.Solid {
		return Refusal{Kind: RefusedTooFar}, nil
	}
	link, k, err := m.climb(from, n, rules.MaxStepUp+1)
	if err != nil || link == nil {
		return Refusal{Kind: RefusedSolid}, err
	}
	return Refusal{Kind: RefusedTooHigh, Height: k}, nil
}
