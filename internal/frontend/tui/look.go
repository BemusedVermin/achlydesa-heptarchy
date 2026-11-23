package tui

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// lookState is look mode: a cursor on the map that the step keys move.
type lookState struct {
	cursor api.Hex
	// lastVertical is the step option the last w or s moved by, so they zigzag as stepping does.
	lastVertical string
}

// hexDelta is how each step option moves a cursor. Drawing a cursor needs only this much of the
// map's geometry; whether anything can step there is the core's business and not asked here.
var hexDelta = map[string]api.Hex{
	api.StepOption("e"): {Q: 1}, api.StepOption("ne"): {Q: 1, R: -1}, api.StepOption("nw"): {R: -1},
	api.StepOption("w"): {Q: -1}, api.StepOption("sw"): {Q: -1, R: 1}, api.StepOption("se"): {R: 1},
}

func (l *lookState) move(option string) {
	d := hexDelta[option]
	l.cursor = api.Hex{Q: l.cursor.Q + d.Q, R: l.cursor.R + d.R}
}

// bodyAt is the body of the view standing at a hex, if one is in sight.
func bodyAt(v api.View, at api.Hex) (api.BodyView, bool) {
	if v.Map == nil || v.Bodies == nil {
		return api.BodyView{}, false
	}
	for _, e := range v.Map.Entities {
		if e.Hex != at {
			continue
		}
		for _, body := range v.Bodies.Bodies {
			if body.UID == e.UID {
				return body, true
			}
		}
	}
	return api.BodyView{}, false
}

// RenderLook is the look panel: the body under the cursor with its parts, materials as the party
// knows them, or a note that there is none.
func RenderLook(v api.View, at api.Hex, width int) string {
	body, ok := bodyAt(v, at)
	if !ok {
		return "Looking: nothing here"
	}
	return "Looking\n" + RenderBody(body, v.Bodies, v.Tick, width)
}
