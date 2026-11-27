package hexmap_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

func TestRotateTurnsEastToEachDirectionAndBack(t *testing.T) {
	origin := api.Cell{}
	east := hexmap.Prisms{}.Neighbor(origin, api.DirE).Cell.Hex
	far := api.Hex{Q: 3, R: -1}
	for dir := range api.Dir(api.DirCount) {
		want := hexmap.Prisms{}.Neighbor(origin, dir).Cell.Hex
		if got := hexmap.Rotate(east, dir); got != want {
			t.Errorf("Rotate(east, %d) = %v, want %v", dir, got, want)
		}
		turned := hexmap.Rotate(far, dir)
		if hexmap.Distance(api.Hex{}, turned) != hexmap.Distance(api.Hex{}, far) {
			t.Errorf("Rotate(%v, %d) = %v, which is not as far from the origin", far, dir, turned)
		}
		if back := hexmap.Rotate(turned, (api.DirCount-dir)%api.DirCount); back != far {
			t.Errorf("turning %v by %d and back gives %v", far, dir, back)
		}
	}
}
