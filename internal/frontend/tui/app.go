package tui

import (
	"io"
	"slices"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// Screen is one full-screen mode. The root model owns a stack of them.
type Screen interface {
	// Update returns the screen to keep in its place, or nil to pop it.
	Update(msg tea.Msg, app *App) (Screen, tea.Cmd)
	View(app *App) string
}

// App is the root model: the game, the latest view, the window size and the screen stack.
type App struct {
	Game *game.Game
	// State is the latest view, re-read after every command.
	State  api.View
	Width  int
	Height int
	// Status is the one-line message under the map: a disabled option's reason, or empty.
	Status string
	// Record, when set, runs after every command the game accepts, so a crash loses nothing.
	Record func(*game.Game) error

	keys  KeyMap
	help  help.Model
	stack []Screen
	err   error
}

// NewApp starts the app on the map screen of g.
func NewApp(g *game.Game) (*App, error) {
	app := &App{Game: g, Width: 100, Height: 30, keys: DefaultKeyMap(), help: help.New()}
	if err := app.refresh(); err != nil {
		return nil, err
	}
	app.stack = []Screen{newMapScreen()}
	app.openWipe()
	return app, nil
}

// Err is the error that ended the game, if one did. A command that fails is a bug, not player
// input, so the app shows it and quits.
func (a *App) Err() error { return a.err }

// Push opens an overlay over the current screen.
func (a *App) Push(s Screen) { a.stack = append(a.stack, s) }

func (a *App) refresh() error {
	view, err := a.Game.View()
	if err != nil {
		return err
	}
	a.State = view
	return nil
}

func (a *App) fail(err error) tea.Cmd {
	a.err = err
	return tea.Quit
}

// Choose sends the command of the option with this ID, then re-reads the view. An option that is
// absent does nothing; one that is disabled shows its reason and sends nothing. An option with a
// question asks it first.
func (a *App) Choose(id string) tea.Cmd {
	option, ok := optionByID(a.State, id)
	if ok && option.Enabled && option.Confirm != "" {
		a.Status = ""
		a.Push(&confirmPopup{id: id, question: option.Confirm})
		return nil
	}
	return a.choose(id)
}

func (a *App) choose(id string) tea.Cmd {
	option, ok := optionByID(a.State, id)
	switch {
	case !ok:
		return nil
	case !option.Enabled:
		a.Status = option.Reason
		return nil
	}
	a.Status = ""
	before := lastSeq(a.State)
	if _, err := a.Game.Apply(option.Command); err != nil {
		return a.fail(err)
	}
	if a.Record != nil {
		if err := a.Record(a.Game); err != nil {
			return a.fail(err)
		}
	}
	if err := a.refresh(); err != nil {
		return a.fail(err)
	}
	a.openWipe()
	if lines := newLines(a.State, before); hasPause(lines) {
		a.Push(newPlayByPlay(lines))
	}
	return nil
}

// openWipe puts the wipe screen on the stack while the view has one, beneath any play-by-play
// pushed after it.
func (a *App) openWipe() {
	if a.State.Wipe == nil {
		return
	}
	if slices.ContainsFunc(a.stack, func(s Screen) bool { _, ok := s.(*wipeScreen); return ok }) {
		return
	}
	a.Push(newWipeScreen())
}

func lastSeq(v api.View) int64 {
	if len(v.Log) == 0 {
		return 0
	}
	return v.Log[len(v.Log)-1].Seq
}

// ContentMsg carries a content reload in a dev build. The watcher sends it to the program, so Reload
// runs on the Bubble Tea goroutine, between commands. An error Reload returns goes to the message
// log, and the game plays on. Reload is required.
type ContentMsg struct {
	Reload func(*game.Game) error
}

func (a *App) reloadContent(msg ContentMsg) tea.Cmd {
	if err := msg.Reload(a.Game); err != nil {
		a.Game.ReloadFailed(err)
	}
	if err := a.refresh(); err != nil {
		return a.fail(err)
	}
	return nil
}

// Init implements [tea.Model].
func (a *App) Init() tea.Cmd { return nil }

// Update implements [tea.Model]: it handles the window size and ctrl+c, and gives everything else
// to the top screen.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width, a.Height = msg.Width, msg.Height
		return a, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case ContentMsg:
		return a, a.reloadContent(msg)
	}
	top := len(a.stack) - 1
	if top < 0 {
		return a, tea.Quit
	}
	next, cmd := a.stack[top].Update(msg, a)
	switch {
	case a.err != nil:
		return a, tea.Quit
	case next == nil:
		a.stack = slices.Delete(a.stack, top, top+1)
		if len(a.stack) == 0 {
			return a, tea.Quit
		}
	default:
		// An overlay pushed by Update sits above top; keep the stack and replace only top.
		a.stack[top] = next
	}
	return a, cmd
}

// View implements [tea.Model].
func (a *App) View() tea.View {
	content := ""
	if len(a.stack) > 0 {
		content = a.stack[len(a.stack)-1].View(a)
	}
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// WriteReplay writes the game's command history as a replay log.
func WriteReplay(g *game.Game, w io.Writer) error {
	return g.Codec().WriteLog(w, g.History())
}
