package prototype

import hex "github.com/gravitton/hexagon"

type Terrain uint8

const (
	Floor Terrain = iota
	Wall
)

// Command is something the player asks the core to do.
type Command interface{ isCommand() }

// Move steps the party one hex.
type Move struct{ Dir hex.Direction }

func (Move) isCommand() {}

// Event is something that happened because of a command.
type Event interface{ isEvent() }

type Moved struct{ From, To hex.Hex }
type Bumped struct{ At hex.Hex }

func (Moved) isEvent()  {}
func (Bumped) isEvent() {}

// Snapshot is what a frontend draws. Treat it as read-only.
type Snapshot struct {
	Tiles   map[hex.Hex]Terrain
	Visible map[hex.Hex]bool
	Party   hex.Hex
}

type Game struct {
	tiles map[hex.Hex]Terrain
	party hex.Hex
}

const (
	mapRadius   = 10
	sightRadius = 6
)

// New builds a hand-made test map: a hex-shaped room with a few pillars.
func New() *Game {
	g := &Game{tiles: map[hex.Hex]Terrain{}}
	center := hex.Pt(0, 0)
	for _, h := range center.Range(mapRadius) {
		g.tiles[h] = Floor
	}
	for _, h := range center.Ring(mapRadius) {
		g.tiles[h] = Wall
	}
	for _, h := range []hex.Hex{hex.Pt(3, -1), hex.Pt(-2, 4), hex.Pt(1, 2)} {
		g.tiles[h] = Wall
	}
	return g
}

func (g *Game) blocked(h hex.Hex) bool {
	t, ok := g.tiles[h]
	return !ok || t == Wall
}

// Apply runs one command and returns what happened.
func (g *Game) Apply(cmd Command) []Event {
	switch c := cmd.(type) {
	case Move:
		to := g.party.Neighbor(c.Dir)
		if g.blocked(to) {
			return []Event{Bumped{At: to}}
		}
		from := g.party
		g.party = to
		return []Event{Moved{From: from, To: to}}
	}
	return nil
}

// View returns the map and what the party can see.
func (g *Game) View() Snapshot {
	visible := map[hex.Hex]bool{}
	for _, h := range g.party.FieldOfView(sightRadius, g.blocked) {
		visible[h] = true
	}
	return Snapshot{Tiles: g.tiles, Visible: visible, Party: g.party}
}
