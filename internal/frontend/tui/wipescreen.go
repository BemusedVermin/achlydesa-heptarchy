package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// wipeScreen is the full-screen stop after a party wipe. It closes when the player chooses the
// wake option, and a lost party has only quitting left.
type wipeScreen struct{}

func newWipeScreen() *wipeScreen { return &wipeScreen{} }

// Update implements [Screen].
func (s *wipeScreen) Update(msg tea.Msg, app *App) (Screen, tea.Cmd) {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch {
	case key.Matches(press, app.keys.quit):
		return s, tea.Quit
	case key.Matches(press, app.keys.menu) && hasOption(app.State, api.OptionWake):
		cmd := app.Choose(api.OptionWake)
		if app.State.Wipe == nil {
			return nil, cmd
		}
	}
	return s, nil
}

// View implements [Screen].
func (s *wipeScreen) View(app *App) string {
	wipe := app.State.Wipe
	if wipe == nil {
		return ""
	}
	width := max(1, app.Width-2)
	rows := []string{lipgloss.NewStyle().Bold(true).Render(wipe.Title), ""}
	for _, line := range wipe.Lines {
		rows = append(rows, lipgloss.NewStyle().Width(width).Render(line))
	}
	rows = append(rows, "")
	for _, option := range app.State.Options {
		rows = append(rows, "> "+option.Label)
	}
	choose := key.NewBinding(key.WithHelp("enter", "choose"))
	rows = append(rows, "", app.help.ShortHelpView([]key.Binding{choose, app.keys.quit}))
	return strings.Join(rows, "\n")
}
