package sim

import (
	"errors"
	"fmt"
)

// Cadence says when a system runs.
type Cadence uint8

const (
	// PerTurn systems run after every actor's turn (turns runs them).
	PerTurn Cadence = iota
	// PerPhase systems run at each day-phase boundary.
	PerPhase
	// PerDay systems run at each day boundary.
	PerDay
	cadenceCount
)

// System is one unit of per-cadence simulation work.
type System struct {
	// Name is unique, "<package>.<what>", e.g. "agents.think".
	Name    string
	Cadence Cadence
	// After names systems of the same cadence that must run first.
	After []string
	Run   func(w *World) error
}

// Scheduler holds the registered systems and, once frozen, their run order.
type Scheduler struct {
	systems map[string]System
	order   [cadenceCount][]string
	frozen  bool
}

// NewScheduler returns an empty, unfrozen scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{systems: make(map[string]System)}
}

// Register adds a system. It errors on a duplicate name, an invalid cadence,
// a missing Run func, or after Freeze.
func (s *Scheduler) Register(sys System) error {
	switch {
	case s.frozen:
		return fmt.Errorf("sim: register system %q after freeze", sys.Name)
	case sys.Cadence >= cadenceCount:
		return fmt.Errorf("sim: system %q has invalid cadence %d", sys.Name, sys.Cadence)
	case sys.Run == nil:
		return fmt.Errorf("sim: system %q has no Run func", sys.Name)
	}
	if _, dup := s.systems[sys.Name]; dup {
		return fmt.Errorf("sim: duplicate system %q", sys.Name)
	}
	s.systems[sys.Name] = sys
	return nil
}

// Freeze sorts each cadence on After, breaking ties by name. It errors on an
// unknown After name, an After naming a system of another cadence, or a cycle.
func (s *Scheduler) Freeze() error {
	var perCadence [cadenceCount][]orderItem
	for _, sys := range s.systems {
		for _, dep := range sys.After {
			other, ok := s.systems[dep]
			if ok && other.Cadence != sys.Cadence {
				return fmt.Errorf("sim: system %q (cadence %d) runs after %q of another cadence (%d)",
					sys.Name, sys.Cadence, dep, other.Cadence)
			}
		}
		perCadence[sys.Cadence] = append(perCadence[sys.Cadence], orderItem{sys.Name, sys.After})
	}
	var order [cadenceCount][]string
	for cadence := range perCadence {
		sorted, err := sortByAfter("system", perCadence[cadence])
		if err != nil {
			return err
		}
		order[cadence] = sorted
	}
	s.order = order
	s.frozen = true
	return nil
}

// Order returns the system names of cadence c in run order. Empty before Freeze.
func (s *Scheduler) Order(c Cadence) []string {
	if c >= cadenceCount {
		return nil
	}
	return append([]string(nil), s.order[c]...)
}

// Run runs every system of cadence c in order, draining the bus after each,
// and stops at the first error. It errors before Freeze.
func (s *Scheduler) Run(w *World, c Cadence) error {
	if !s.frozen {
		return errors.New("sim: run systems before freeze")
	}
	if c >= cadenceCount {
		return fmt.Errorf("sim: invalid cadence %d", c)
	}
	for _, name := range s.order[c] {
		if err := s.systems[name].Run(w); err != nil {
			return fmt.Errorf("sim: system %q: %w", name, err)
		}
		if _, err := w.Bus.Drain(w); err != nil {
			return fmt.Errorf("sim: after system %q: %w", name, err)
		}
	}
	return nil
}
