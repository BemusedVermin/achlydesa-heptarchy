package hexmap

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Rotate turns h about the origin hex by whole sixths of a turn, so that what pointed [api.DirE] points
// to dir. The directions run anticlockwise, so each step past DirE is a sixth of a turn anticlockwise.
// An invalid dir leaves h as it is.
func Rotate(h api.Hex, dir api.Dir) api.Hex {
	if !dir.Valid() {
		return h
	}
	for range dir {
		h = api.Hex{Q: h.Q + h.R, R: -h.Q}
	}
	return h
}
