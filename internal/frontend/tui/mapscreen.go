package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// sidePanelWidth is the right-hand column: the clock, the party, the acting body, and the log or the
// action menu.
const sidePanelWidth = 60

// mapScreen is the one screen of play: the timeline strip on top, the map on the left, the side
// panel on the right, then the status and help lines.
type mapScreen struct {
	// selected is the offset of the selected timeline tick.
	selected int
	// lastVertical is the step option the last w or s sent, or empty after any other key.
	lastVertical string
	// menu and look are nil while the action menu is closed and look mode is off.
	menu *menuState
	look *lookState
}

func newMapScreen() *mapScreen { return &mapScreen{} }

// openMenu opens the action menu when the view offers anything to put in it.
func (s *mapScreen) openMenu(app *App) bool {
	s.menu = newMenu(app.State, sidePanelWidth, 10)
	return s.menu != nil
}

func (s *mapScreen) startLook(app *App) {
	if app.State.Bodies == nil || app.State.Party == nil {
		return
	}
	s.look = &lookState{cursor: app.State.Party.Hex}
}

// menuKey handles a key while the action menu is open.
func (s *mapScreen) menuKey(msg tea.KeyPressMsg, app *App) tea.Cmd {
	id, closed := s.menu.update(msg)
	if closed {
		s.menu = nil
		return nil
	}
	if id == "" {
		return nil
	}
	option, _ := optionByID(app.State, id)
	cmd := app.Choose(id)
	if option.Enabled || cmd != nil {
		s.menu = nil
	}
	s.selectTick(app, s.selected)
	return cmd
}

// lookKey handles a key while look mode is on: the step keys move the cursor.
func (s *mapScreen) lookKey(msg tea.KeyPressMsg, app *App) tea.Cmd {
	up, down := key.Matches(msg, app.keys.up), key.Matches(msg, app.keys.down)
	switch {
	case key.Matches(msg, app.keys.quit) && msg.String() != "ctrl+c", key.Matches(msg, app.keys.look):
		s.look = nil
	case up || down:
		option := VerticalOptions(up, s.look.lastVertical)[0]
		s.look.lastVertical = option
		s.look.move(option)
	default:
		for _, b := range app.keys.moves {
			if key.Matches(msg, b.keys) {
				s.look.lastVertical = ""
				s.look.move(b.option)
			}
		}
	}
	return nil
}

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
	switch {
	case s.menu != nil:
		return s.menuKey(msg, app)
	case s.look != nil:
		return s.lookKey(msg, app)
	}
	up, down := key.Matches(msg, app.keys.up), key.Matches(msg, app.keys.down)
	if up || down {
		return s.vertical(app, up)
	}
	s.lastVertical = ""
	switch {
	case key.Matches(msg, app.keys.menu):
		if !s.openMenu(app) {
			s.open(app)
		}
	case key.Matches(msg, app.keys.look):
		s.startLook(app)
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
	option, _ := optionByID(app.State, id)
	cmd := app.Choose(id)
	if option.Enabled {
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
		app.Push(&tickPopup{entries: entries})
	}
}

// View implements [Screen].
func (s *mapScreen) View(app *App) string {
	width, height := app.Width, app.Height
	side := min(sidePanelWidth, width/3)
	mapWidth := max(1, width-side-1)
	bodyHeight := max(1, height-TimelineRows-2)

	strip := RenderTimelineMarked(app.State.Timeline, app.State.Bodies, width, s.selected)
	var target api.UID
	var cursor *api.Hex
	if s.menu != nil {
		target = s.menu.targetUID(app.State.Bodies)
	}
	if s.look != nil {
		cursor = &s.look.cursor
	}
	overlay := OverlayFor(app.State.Bodies, target, cursor)
	mapPane := RenderMapOverlay(app.State.Map, app.State.Party, overlay, mapWidth, bodyHeight)

	panel := []string{RenderClock(app.State.Clock, side), "", RenderParty(app.State.Party), ""}
	switch body, ok := currentBody(app.State.Bodies); {
	case s.look != nil:
		panel = append(panel, RenderLook(app.State, s.look.cursor, side), "")
	case ok:
		panel = append(panel, RenderBody(body, app.State.Bodies, side), "")
	}
	used := strings.Count(strings.Join(panel, "\n"), "\n") + 1
	if s.menu != nil {
		s.menu.list.SetSize(side, max(2, bodyHeight-used-1))
		panel = append(panel, s.menu.Render())
	} else {
		panel = append(panel, RenderLog(app.State.Log, side, max(1, bodyHeight-used)))
	}
	sidePane := lipgloss.NewStyle().Width(side).Height(bodyHeight).MaxHeight(bodyHeight).
		Render(strings.Join(panel, "\n"))

	body := lipgloss.JoinHorizontal(lipgloss.Top, mapPane, " ", sidePane)
	return strings.Join(
		[]string{strip, body, s.status(app), app.help.ShortHelpView(app.keys.HelpBindings(app.State))},
		"\n",
	)
}

// status is the status line: the reason a key did nothing, or the party's status from the view.
func (s *mapScreen) status(app *App) string {
	if app.Status != "" {
		return app.Status
	}
	if app.State.Party == nil {
		return ""
	}
	return strings.Join(app.State.Party.Status, " | ")
}

// tickPopup lists everything on one timeline tick, by label. It has at least one entry.
type tickPopup struct {
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
	lines := []string{p.entries[0].TickLabel, ""}
	for _, e := range p.entries {
		lines = append(lines, e.Label)
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(app.Width, app.Height, lipgloss.Center, lipgloss.Center, box)
}
