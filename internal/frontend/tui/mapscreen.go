package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// sidePanelWidth is the right-hand column: the clock, the party and the log. M2-10's action panel
// goes in the same column.
const sidePanelWidth = 34

// mapScreen is the one screen of play: the timeline strip on top, the map on the left, the side
// panel on the right, then the status and help lines.
type mapScreen struct {
	selected int // offset of the selected timeline tick
	// lastVertical is the step option the last w or s sent, or empty after any other key.
	lastVertical string
}

func newMapScreen() *mapScreen { return &mapScreen{} }

// Update implements [Screen].
func (s *mapScreen) Update(msg tea.Msg, app *App) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return s, s.key(msg, app)
	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		if mouse.Button == tea.MouseLeft && mouse.Y < TimelineRows {
			s.selectTick(app, mouse.X/timelineCell)
			s.open(app)
		}
	}
	return s, nil
}

func (s *mapScreen) key(msg tea.KeyPressMsg, app *App) tea.Cmd {
	up, down := key.Matches(msg, app.keys.up), key.Matches(msg, app.keys.down)
	if up || down {
		return s.vertical(app, up)
	}
	s.lastVertical = ""
	switch {
	case key.Matches(msg, app.keys.quit):
		return tea.Quit
	case key.Matches(msg, app.keys.prevTick):
		s.selectTick(app, s.selected-1)
	case key.Matches(msg, app.keys.nextTick):
		s.selectTick(app, s.selected+1)
	case key.Matches(msg, app.keys.selectTic):
		s.open(app)
	default:
		if id, ok := app.keys.OptionFor(msg, app.State); ok {
			cmd := app.Choose(id)
			s.selectTick(app, s.selected)
			return cmd
		}
	}
	return nil
}

// vertical zigzags the party up or down the screen, alternating two diagonals.
func (s *mapScreen) vertical(app *App, up bool) tea.Cmd {
	id, ok := VerticalOption(app.State, up, s.lastVertical)
	if !ok {
		return nil
	}
	enabled := false
	for _, o := range app.State.Options {
		if o.ID == id {
			enabled = o.Enabled
		}
	}
	cmd := app.Choose(id)
	if enabled {
		s.lastVertical = id
	}
	s.selectTick(app, s.selected)
	return cmd
}

// selectTick moves the selection, keeping it inside the strip.
func (s *mapScreen) selectTick(app *App, offset int) {
	ticks := TimelineTicks(app.State.Timeline, app.Width)
	s.selected = max(0, min(ticks-1, offset))
}

func (s *mapScreen) open(app *App) {
	if entries := EntriesAt(app.State.Timeline, s.selected); len(entries) > 0 {
		app.Push(&tickPopup{tick: app.State.Timeline.Now + int64(s.selected), entries: entries})
	}
}

// View implements [Screen].
func (s *mapScreen) View(app *App) string {
	width, height := app.Width, app.Height
	side := min(sidePanelWidth, width/3)
	mapWidth := max(1, width-side-1)
	bodyHeight := max(1, height-TimelineRows-2)

	strip := RenderTimeline(app.State.Timeline, width, s.selected)
	mapPane := RenderMap(app.State.Map, app.State.Party, mapWidth, bodyHeight)

	panel := []string{RenderClock(app.State.Clock, side), "", RenderParty(app.State.Party), ""}
	used := strings.Count(strings.Join(panel, "\n"), "\n") + 1
	panel = append(panel, RenderLog(app.State.Log, side, max(1, bodyHeight-used)))
	sidePane := lipgloss.NewStyle().Width(side).Height(bodyHeight).MaxHeight(bodyHeight).
		Render(strings.Join(panel, "\n"))

	body := lipgloss.JoinHorizontal(lipgloss.Top, mapPane, " ", sidePane)
	return strings.Join(
		[]string{strip, body, s.status(app), app.help.ShortHelpView(app.keys.HelpBindings(app.State))},
		"\n",
	)
}

// status is the status line: the reason a key did nothing, or the party's control mode.
func (s *mapScreen) status(app *App) string {
	if app.Status != "" {
		return app.Status
	}
	if mode := app.Mode(); mode != "" {
		return "Control: " + mode
	}
	return ""
}

// tickPopup lists everything on one timeline tick, by label.
type tickPopup struct {
	tick    int64
	entries []api.TimelineEntry
}

// Update implements [Screen]: esc or enter closes it.
func (p *tickPopup) Update(msg tea.Msg, _ *App) (Screen, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc", "enter":
			return nil, nil
		}
	}
	return p, nil
}

// View implements [Screen].
func (p *tickPopup) View(app *App) string {
	lines := []string{fmt.Sprintf("Tick %d", p.tick), ""}
	for _, e := range p.entries {
		lines = append(lines, e.Label)
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(app.Width, app.Height, lipgloss.Center, lipgloss.Center, box)
}
