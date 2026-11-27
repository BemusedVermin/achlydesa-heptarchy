package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// playByPlay is the overlay that shows the lines one command added, one at a time.
type playByPlay struct {
	lines []api.LogLine
	// shown is how many lines are on screen.
	shown int
}

// newPlayByPlay opens on the first of the lines.
func newPlayByPlay(lines []api.LogLine) *playByPlay { return &playByPlay{lines: lines, shown: 1} }

// newLines are the lines of the view added after the sequence number.
func newLines(v api.View, afterSeq int64) []api.LogLine {
	var out []api.LogLine
	for _, line := range v.Log {
		if line.Seq > afterSeq {
			out = append(out, line)
		}
	}
	return out
}

// hasPause reports whether any of the lines is a fight beat.
func hasPause(lines []api.LogLine) bool {
	for _, line := range lines {
		if line.Pause {
			return true
		}
	}
	return false
}

// Update implements [Screen]: advance shows the next line, skip jumps to the end, and a key
// pressed on the last line closes the overlay.
func (s *playByPlay) Update(msg tea.Msg, app *App) (Screen, tea.Cmd) {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch {
	case key.Matches(press, app.keys.skip):
		if s.shown >= len(s.lines) {
			return nil, nil
		}
		s.shown = len(s.lines)
	default:
		if s.shown >= len(s.lines) {
			return nil, nil
		}
		s.shown++
	}
	return s, nil
}

// View implements [Screen].
func (s *playByPlay) View(app *App) string {
	width := max(1, app.Width-2)
	rows := make([]string, 0, s.shown+2)
	for _, line := range s.lines[:s.shown] {
		rows = append(rows, lipgloss.NewStyle().Width(width).Render(line.Text))
	}
	rows = append(rows, "", app.help.ShortHelpView([]key.Binding{app.keys.advance, app.keys.skip}))
	return strings.Join(rows, "\n")
}
