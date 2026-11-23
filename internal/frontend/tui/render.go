package tui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// maxShade is how many cells of relative depth change the shading; deeper or higher cells draw as
// at this distance.
const maxShade = 8

type rgb struct{ r, g, b float64 }

// neutral is the color of anything whose content names no color, or one this table doesn't know.
var neutral = rgb{154, 154, 154}

// colorTable maps content color names to RGB. It is the only place a color name is interpreted.
var colorTable = map[string]rgb{
	"black":   {40, 40, 40},
	"red":     {220, 70, 70},
	"green":   {80, 200, 100},
	"yellow":  {230, 200, 80},
	"blue":    {90, 130, 230},
	"magenta": {200, 90, 200},
	"cyan":    {80, 200, 210},
	"white":   {235, 235, 235},
	"gray":    {154, 154, 154},
	"grey":    {154, 154, 154},
	"orange":  {230, 140, 60},
	"brown":   {160, 110, 70},
}

func lookupColor(name string) rgb {
	if c, ok := colorTable[strings.ToLower(name)]; ok {
		return c
	}
	return neutral
}

// shade darkens a color one step for each cell lower than the party and lightens it one step for
// each cell higher, out to maxShade, and dims it further when it is only remembered.
func shade(c rgb, below int, remembered bool) rgb {
	below = max(-maxShade, min(maxShade, below))
	switch {
	case below > 0:
		f := 1 - 0.08*float64(below)
		c = rgb{c.r * f, c.g * f, c.b * f}
	case below < 0:
		t := 0.06 * float64(-below)
		c = rgb{c.r + (255-c.r)*t, c.g + (255-c.g)*t, c.b + (255-c.b)*t}
	}
	if remembered {
		c = rgb{c.r * 0.45, c.g * 0.45, c.b * 0.45}
	}
	return c
}

func (c rgb) color() color.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", uint8(c.r+0.5), uint8(c.g+0.5), uint8(c.b+0.5)))
}

// paint draws text in a color, shaded.
func paint(text string, c rgb, below int, remembered, bold bool) string {
	style := lipgloss.NewStyle().Foreground(shade(c, below, remembered).color()).Bold(bold)
	return style.Render(text)
}

// glyphOf is the first character of a content glyph, or a space.
func glyphOf(s string) string {
	for _, r := range s {
		return string(r)
	}
	return " "
}

// drawn is one grid position of the map.
type drawn struct {
	glyph      string
	color      rgb
	below      int
	remembered bool
	bold       bool
	mark       mark
}

// RenderMap draws the plan view in double-width layout: hex (q, r) is at column 2q+r, row r. The
// result is exactly height lines of width columns. When the map is larger than the pane it is
// scrolled to center on the party. Depth shows only as shading relative to the party.
func RenderMap(m *api.MapView, party *api.PartyView, width, height int) string {
	return RenderMapOverlay(m, party, MapOverlay{}, width, height)
}

// RenderMapOverlay is [RenderMap] with the combat overlay drawn on it.
func RenderMapOverlay(m *api.MapView, party *api.PartyView, overlay MapOverlay, width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}
	grid := map[[2]int]drawn{}
	place := func(h api.Hex, d drawn) { grid[[2]int{2*h.Q + h.R, h.R}] = d }
	minCol, maxCol, minRow, maxRow := 0, 0, 0, 0
	first := true
	extend := func(h api.Hex) {
		col, row := 2*h.Q+h.R, h.R
		if first {
			minCol, maxCol, minRow, maxRow = col, col, row, row
			first = false
			return
		}
		minCol, maxCol = min(minCol, col), max(maxCol, col)
		minRow, maxRow = min(minRow, row), max(maxRow, row)
	}
	if m != nil {
		for _, cell := range m.Cells {
			extend(cell.Hex)
			place(
				cell.Hex,
				drawn{
					glyph:      glyphOf(cell.Glyph),
					color:      lookupColor(cell.Color),
					below:      cell.Below,
					remembered: !cell.Visible,
				},
			)
		}
		members := map[api.UID]bool{}
		current := map[api.UID]bool{}
		if party != nil {
			for _, member := range party.Members {
				members[member.UID] = true
				current[member.UID] = member.Current
			}
		}
		// Entities draw over terrain; the party draws over everything.
		for _, pass := range []bool{false, true} {
			for _, e := range m.Entities {
				if members[e.UID] != pass {
					continue
				}
				extend(e.Hex)
				ink := rgb{235, 235, 235}
				if e.Color != "" {
					ink = lookupColor(e.Color)
				}
				d := drawn{glyph: glyphOf(e.Glyph), color: ink, below: e.Below, bold: current[e.UID]}
				if e.UID == overlay.Target && overlay.Target != 0 {
					d.mark = markTarget
				}
				if ticks, ok := overlay.Threats[e.UID]; ok {
					if d.mark == markNone {
						d.mark = markThreat
					}
					grid[[2]int{2*e.Hex.Q + e.Hex.R + 1, e.Hex.R}] = drawn{
						glyph: plural(int(min(ticks, 10))), color: lookupColor("red"), bold: true,
					}
				}
				place(e.Hex, d)
			}
		}
	}
	if overlay.Cursor != nil {
		extend(*overlay.Cursor)
		key := [2]int{2*overlay.Cursor.Q + overlay.Cursor.R, overlay.Cursor.R}
		d, ok := grid[key]
		if !ok {
			d = drawn{glyph: " ", color: neutral}
		}
		d.mark = markCursor
		grid[key] = d
	}
	originCol, originRow := minCol, minRow
	if maxCol-minCol+1 > width || maxRow-minRow+1 > height {
		centerCol, centerRow := (minCol+maxCol)/2, (minRow+maxRow)/2
		if party != nil {
			centerCol, centerRow = 2*party.Hex.Q+party.Hex.R, party.Hex.R
		}
		if maxCol-minCol+1 > width {
			originCol = max(minCol, min(maxCol-width+1, centerCol-width/2))
		}
		if maxRow-minRow+1 > height {
			originRow = max(minRow, min(maxRow-height+1, centerRow-height/2))
		}
	}
	lines := make([]string, height)
	for y := range height {
		var line strings.Builder
		for x := range width {
			if d, ok := grid[[2]int{originCol + x, originRow + y}]; ok {
				line.WriteString(d.mark.apply(paint(d.glyph, d.color, d.below, d.remembered, d.bold)))
			} else {
				line.WriteByte(' ')
			}
		}
		lines[y] = line.String()
	}
	return strings.Join(lines, "\n")
}

// TimelineRows is how many lines the timeline strip takes.
const TimelineRows = 3

// timelineCell is how many columns one tick takes in the strip.
const timelineCell = 2

// TimelineTicks is how many ticks fit in a strip of the given width.
func TimelineTicks(t *api.TimelineView, width int) int {
	if t == nil {
		return 0
	}
	return int(min(t.Window+1, int64(width/timelineCell)))
}

// EntriesAt lists the timeline entries that fall on the tick that is offset ticks after Now.
func EntriesAt(t *api.TimelineView, offset int) []api.TimelineEntry {
	if t == nil {
		return nil
	}
	var out []api.TimelineEntry
	for _, e := range t.Entries {
		if e.Tick == t.Now+int64(offset) {
			out = append(out, e)
		}
	}
	return out
}

// RenderTimeline draws the strip of upcoming ticks: one two-column cell per tick from Now, with
// the current tick at the left edge. The first line holds each tick's first icon (the actor's glyph
// for a decision, "!" for something landing, in the actor's color), the second "+N" when more
// share the tick, and the third a ruler. selected is the offset of the chosen tick, or -1.
func RenderTimeline(t *api.TimelineView, width, selected int) string {
	return RenderTimelineMarked(t, nil, width, selected)
}

// RenderTimelineMarked is [RenderTimeline] with combat marks: a hostile body's entry is
// underlined, and a tick where something lands on a party member shows a red "!" instead.
func RenderTimelineMarked(t *api.TimelineView, bodies *api.BodiesView, width, selected int) string {
	hostile, members := map[api.UID]bool{}, map[api.UID]bool{}
	if bodies != nil {
		for _, b := range bodies.Bodies {
			hostile[b.UID], members[b.UID] = b.Hostile, b.Member
		}
	}
	ticks := TimelineTicks(t, width)
	icons, counts, ruler := make([]string, 0, ticks), make([]string, 0, ticks), make([]string, 0, ticks)
	for offset := range ticks {
		entries := EntriesAt(t, offset)
		icon, count := "  ", "  "
		if len(entries) > 0 {
			first := entries[0]
			mark := glyphOf(first.Glyph)
			if first.Kind == "lands" {
				mark = "!"
			}
			icon = paint(mark, lookupColor(first.Color), 0, false, false) + " "
			if hostile[first.Actor] {
				icon = lipgloss.NewStyle().Underline(true).Render(icon[:len(icon)-1]) + " "
			}
			for _, e := range entries {
				if e.Kind == "lands" && members[e.Target] && !members[e.Actor] {
					icon = lipgloss.NewStyle().Bold(true).Foreground(lookupColor("white").color()).
						Background(lookupColor("red").color()).Render("!") + " "
					break
				}
			}
			if len(entries) > 1 {
				count = fmt.Sprintf("+%s", plural(len(entries)-1))
			}
		}
		mark := ". "
		switch {
		case offset == selected:
			mark = "^ "
		case offset%10 == 0:
			mark = "| "
		}
		if offset == selected {
			icon, count = lipgloss.NewStyle().
				Reverse(true).
				Render(icon),
				lipgloss.NewStyle().
					Reverse(true).
					Render(count)
		}
		icons, counts, ruler = append(icons, icon), append(counts, count), append(ruler, mark)
	}
	return strings.Join([]string{strings.Join(icons, ""), strings.Join(counts, ""), strings.Join(ruler, "")}, "\n")
}

// plural is a one-character count: 1-9, or "*" for more.
func plural(n int) string {
	if n > 9 {
		return "*"
	}
	return fmt.Sprint(n)
}

// RenderClock draws the day, the phase and a bar for the ticks elapsed in the phase.
func RenderClock(c *api.ClockView, width int) string {
	if c == nil {
		return ""
	}
	barWidth := max(4, min(24, width-2))
	filled := 0
	if c.PhaseTicks > 0 {
		filled = int(int64(barWidth) * c.TickInPhase / c.PhaseTicks)
	}
	filled = max(0, min(barWidth, filled))
	bar := "[" + strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled) + "]"
	return fmt.Sprintf("Day %d · %s\n%s", c.Day, c.Phase, bar)
}

// RenderParty lists the members in order: the leader marked with "*", the current actor in bold
// and reverse.
func RenderParty(p *api.PartyView) string {
	if p == nil {
		return ""
	}
	lines := make([]string, 0, len(p.Members))
	for _, member := range p.Members {
		mark := " "
		if member.Leader {
			mark = "*"
		}
		line := mark + " " + paint(
			glyphOf(member.Glyph),
			lookupColor(member.Color),
			0,
			false,
			false,
		) + " " + member.Name
		if member.Current {
			line = lipgloss.NewStyle().Bold(true).Reverse(true).Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// RenderLog draws the last lines of the message log that fit in height lines of width columns.
func RenderLog(lines []api.LogLine, width, height int) string {
	if height < 1 {
		return ""
	}
	start := max(0, len(lines)-height)
	out := make([]string, 0, height)
	for _, line := range lines[start:] {
		out = append(out, lipgloss.NewStyle().MaxWidth(max(1, width)).Render(line.Text))
	}
	return strings.Join(out, "\n")
}
