package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	hex "github.com/gravitton/hexagon"

	prototype "github.com/BemusedVermin/achlydesa-heptarchy/internal/prototype"
)

var moveKeys = map[string]hex.Direction{
	"d": hex.PointyTopEast,
	"a": hex.PointyTopWest,
	"e": hex.PointyTopNorthEast,
	"q": hex.PointyTopNorthWest,
	"c": hex.PointyTopSouthEast,
	"z": hex.PointyTopSouthWest,
}

type model struct {
	g   *prototype.Game
	msg string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	}
	if dir, ok := moveKeys[key.String()]; ok {
		m.msg = ""
		for _, ev := range m.g.Apply(prototype.Move{Dir: dir}) {
			if _, bumped := ev.(prototype.Bumped); bumped {
				m.msg = "Packed clay. No way through."
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	snap := m.g.View()

	// Bounds of the map in double-width coordinates.
	minX, minY, maxX, maxY := 0, 0, 0, 0
	for h := range snap.Tiles {
		p := hex.DoubleWidth.To(h)
		minX, maxX = min(minX, p.X), max(maxX, p.X)
		minY, maxY = min(minY, p.Y), max(maxY, p.Y)
	}
	rows := make([][]rune, maxY-minY+1)
	for y := range rows {
		rows[y] = []rune(strings.Repeat(" ", maxX-minX+1))
	}
	for h, t := range snap.Tiles {
		if !snap.Visible[h] {
			continue
		}
		p := hex.DoubleWidth.To(h)
		glyph := '.'
		if t == prototype.Wall {
			glyph = '#'
		}
		rows[p.Y-minY][p.X-minX] = glyph
	}
	pp := hex.DoubleWidth.To(snap.Party)
	rows[pp.Y-minY][pp.X-minX] = '@'

	var b strings.Builder
	for _, row := range rows {
		b.WriteString(string(row))
		b.WriteByte('\n')
	}
	b.WriteString("\nq e / a d / z c to move, esc to quit\n")
	b.WriteString(m.msg)

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func main() {
	if os.Getenv("DEBUG") != "" {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
	}

	if _, err := tea.NewProgram(model{g: prototype.New()}).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
