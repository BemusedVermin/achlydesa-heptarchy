package sim

import (
	"fmt"
	"slices"
	"strings"
)

// orderItem is one named thing with ordering constraints.
type orderItem struct {
	name  string
	after []string
}

// sortByAfter orders items so each comes after everything in its after list.
// Ties break by name ascending (Kahn's algorithm with a sorted ready set).
// kind ("system", "view builder") only words the errors. Errors name the items involved.
func sortByAfter(kind string, items []orderItem) ([]string, error) {
	byName := make(map[string]orderItem, len(items))
	for _, item := range items {
		byName[item.name] = item
	}
	pending := make(map[string]int, len(items))
	dependents := make(map[string][]string, len(items))
	for _, item := range items {
		for _, dep := range item.after {
			if _, ok := byName[dep]; !ok {
				return nil, fmt.Errorf("sim: %s %q runs after unknown %s %q", kind, item.name, kind, dep)
			}
			pending[item.name]++
			dependents[dep] = append(dependents[dep], item.name)
		}
	}
	var ready []string
	for _, item := range items {
		if pending[item.name] == 0 {
			ready = append(ready, item.name)
		}
	}
	order := make([]string, 0, len(items))
	for len(ready) > 0 {
		slices.Sort(ready)
		next := ready[0]
		ready = ready[1:]
		order = append(order, next)
		for _, dependent := range dependents[next] {
			pending[dependent]--
			if pending[dependent] == 0 {
				ready = append(ready, dependent)
			}
		}
	}
	if len(order) < len(items) {
		var stuck []string
		for _, item := range items {
			if pending[item.name] > 0 {
				stuck = append(stuck, item.name)
			}
		}
		slices.Sort(stuck)
		return nil, fmt.Errorf("sim: cycle among %ss: %s", kind, strings.Join(stuck, ", "))
	}
	return order, nil
}
