package tui

// MenuPane is the open action menu as the map screen draws it at the app's size, or empty when no
// menu is open.
func MenuPane(a *App) string {
	a.View()
	screen, ok := a.stack[len(a.stack)-1].(*mapScreen)
	if !ok || screen.menu == nil {
		return ""
	}
	return screen.menu.Render()
}
