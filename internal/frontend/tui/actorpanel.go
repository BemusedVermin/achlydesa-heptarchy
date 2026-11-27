package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// mark is how the map overlay highlights a cell.
type mark uint8

const (
	markNone mark = iota
	markThreat
	markTarget
	markCursor
)

// apply styles already-painted text for the mark.
func (m mark) apply(text string) string {
	switch m {
	case markThreat:
		return lipgloss.NewStyle().Background(lookupColor("red").color()).Render(text)
	case markTarget:
		return lipgloss.NewStyle().Reverse(true).Bold(true).Render(text)
	case markCursor:
		return lipgloss.NewStyle().Reverse(true).Underline(true).Render(text)
	default:
		return text
	}
}

// MapOverlay is what combat draws over the map.
type MapOverlay struct {
	// Threats maps each party member something is aimed at to the ticks until the soonest landing.
	Threats map[api.UID]int64
	// Target is the body whose hex is highlighted while a target or part is being chosen, or 0.
	Target api.UID
	// Cursor is the look cursor, if look mode is on.
	Cursor *api.Hex
	// Heights draws each visible open cell as its height relative to the party.
	Heights bool
}

// OverlayFor builds the overlay for a view: the threats, plus the target and cursor given.
func OverlayFor(bodies *api.BodiesView, target api.UID, cursor *api.Hex) MapOverlay {
	overlay := MapOverlay{Target: target, Cursor: cursor}
	if bodies == nil {
		return overlay
	}
	for _, threat := range bodies.Threats {
		if overlay.Threats == nil {
			overlay.Threats = map[api.UID]int64{}
		}
		if soonest, ok := overlay.Threats[threat.Target]; !ok || threat.Lands < soonest {
			overlay.Threats[threat.Target] = threat.Lands
		}
	}
	return overlay
}

// bodyNames maps every body and part UID of the view to a name.
func bodyNames(bodies *api.BodiesView) map[api.UID]string {
	names := map[api.UID]string{}
	if bodies == nil {
		return names
	}
	for _, body := range bodies.Bodies {
		names[body.UID] = body.Name
		for _, channel := range body.Channels {
			names[channel.UID] = channel.Name
		}
	}
	return names
}

// currentBody is the deciding body when it belongs to the party.
func currentBody(bodies *api.BodiesView) (api.BodyView, bool) {
	if bodies == nil {
		return api.BodyView{}, false
	}
	for _, body := range bodies.Bodies {
		if body.UID == bodies.Current && body.Member {
			return body, true
		}
	}
	return api.BodyView{}, false
}

// RenderBody draws one body: a header, one line per part (slot, name, material, integrity bar and
// state), and each swing being wound up with its target and landing tick. A line wider than width
// wraps rather than lose its end, which holds the numbers.
func RenderBody(body api.BodyView, bodies *api.BodiesView, width int) string {
	names := bodyNames(bodies)
	header := body.Name
	if body.Faction != "" {
		header += " (" + body.Faction + ")"
	}
	if len(body.Statuses) > 0 {
		header += " [" + strings.Join(body.Statuses, ", ") + "]"
	}
	lines := []string{lipgloss.NewStyle().Bold(true).Render(header)}
	var swings []api.WindupView
	for _, c := range body.Channels {
		fields := []string{fmt.Sprintf("%-5s", c.Slot), c.Name, c.Material, bar(c.Integrity, c.MaxIntegrity), c.State}
		lines = append(lines, lipgloss.Wrap(strings.Join(fields, " "), max(1, width), ""))
		if c.Windup != nil && (len(swings) == 0 || swings[len(swings)-1] != *c.Windup) {
			swings = append(swings, *c.Windup)
		}
	}
	for _, s := range swings {
		line := "~ " + s.Action
		if name, ok := names[s.Target]; ok {
			line += " -> " + name
		}
		line += ", " + s.When
		lines = append(lines, lipgloss.Wrap(line, max(1, width), ""))
	}
	return strings.Join(lines, "\n")
}

// bar is a four-cell integrity bar.
func bar(have, most int) string {
	const cells = 4
	filled := 0
	if most > 0 {
		filled = (have*cells + most - 1) / most
	}
	filled = max(0, min(cells, filled))
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", cells-filled) + "]"
}
