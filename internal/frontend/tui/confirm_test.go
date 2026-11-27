package tui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/frontend/tui"
)

func askingApp(t *testing.T) (*tui.App, string) {
	t.Helper()
	app := startApp(t)
	for i, option := range app.State.Options {
		if option.Enabled && strings.HasPrefix(option.ID, "step/") {
			app.State.Options[i].Confirm = "Jump down 3 cells?"
			return app, option.ID
		}
	}
	t.Fatal("no enabled step option")
	return nil, ""
}

func TestAnOptionWithAQuestionSendsOnYesAndNothingOnNo(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
		sent bool
	}{
		{"y", tea.KeyPressMsg{Code: 'y', Text: "y"}, true},
		{"enter", tea.KeyPressMsg{Code: tea.KeyEnter}, true},
		{"n", tea.KeyPressMsg{Code: 'n', Text: "n"}, false},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEscape}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, id := askingApp(t)
			before := len(app.Game.History())
			app.Choose(id)
			if got := len(app.Game.History()); got != before {
				t.Fatalf("history grew to %d before the answer", got)
			}
			if !strings.Contains(app.View().Content, "Jump down 3 cells?") {
				t.Fatalf("no question on screen:\n%s", app.View().Content)
			}
			app.Update(tt.key)
			if sent := len(app.Game.History()) > before; sent != tt.sent {
				t.Errorf("sent = %v, want %v", sent, tt.sent)
			}
			if strings.Contains(app.View().Content, "Jump down 3 cells?") {
				t.Error("question still on screen after the answer")
			}
		})
	}
}
