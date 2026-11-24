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

// Step finds where a body at from ends up moving one hex in dir, if anywhere. Bodies are one cell tall. A step
// through a link to another map is taken as given.
func (m *Map) Step(from api.Cell, dir api.Dir, rules StepRules) (Step, bool, error) {
	if !dir.Valid() {
		return Step{}, false, ErrBadDir
	}
	link := m.topo.Neighbor(from, dir)
	n, local := m.local(link)
	if !local {
		return Step{To: link, Kind: StepLevel}, true, nil
	}
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
	above, head := n, from
	for k := 1; k <= rules.MaxStepUp; k++ {
		aboveLink, headLink := m.topo.Above(above), m.topo.Above(head)
		var ok bool
		if above, ok = m.local(aboveLink); !ok {
			return Step{}, false, nil
		}
		if head, ok = m.local(headLink); !ok {
			return Step{}, false, nil
		}
		blocked, err := m.solid(headLink)
		if err != nil || blocked {
			// No headroom, and climbing higher needs even more
			return Step{}, false, err
		}
		standable, err := m.Standable(above)
		if err != nil {
			return Step{}, false, err
		}
		if standable {
			return Step{To: aboveLink, Kind: StepUp}, true, nil
		}
	}
	return Step{}, false, nil
}

// stepDown finds the first standable cell below the open cell n, at most MaxDrop cells down.
func (m *Map) stepDown(n api.Cell, rules StepRules) (Step, bool, error) {
	cur := n
	for fall := 1; fall <= rules.MaxDrop; fall++ {
		link := m.topo.Below(cur)
		var ok bool
		if cur, ok = m.local(link); !ok {
			return Step{}, false, nil
		}
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
