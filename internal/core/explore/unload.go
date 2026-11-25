package explore

import (
	"cmp"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// unloadFar drops loaded chunks past the world's chunk budget, farthest from every party member first,
// ties by key. A chunk within the near radius of a member is kept even over budget.
func unloadFar(w *sim.World) {
	world := w.Content.World()
	if w.Map == nil || world.ChunkBudget <= 0 {
		return
	}
	loaded := w.Map.Loaded()
	if len(loaded) <= world.ChunkBudget {
		return
	}
	members := memberChunks(w)
	type far struct {
		key      hexmap.ChunkKey
		distance int
	}
	candidates := make([]far, 0, len(loaded))
	for _, key := range loaded {
		distance := -1
		for _, member := range members {
			if d := hexmap.ChunkDistance(key, member); distance < 0 || d < distance {
				distance = d
			}
		}
		if distance < 0 || distance > world.NearChunks {
			candidates = append(candidates, far{key: key, distance: distance})
		}
	}
	slices.SortFunc(candidates, func(a, b far) int {
		return cmp.Or(cmp.Compare(b.distance, a.distance), hexmap.CompareKeys(a.key, b.key))
	})
	excess := min(len(loaded)-world.ChunkBudget, len(candidates))
	drop := make(map[hexmap.ChunkKey]bool, excess)
	for _, c := range candidates[:excess] {
		drop[c.key] = true
	}
	w.Map.Unload(func(key hexmap.ChunkKey) bool { return !drop[key] })
}

// memberChunks lists the chunk of every party member that stands somewhere.
func memberChunks(w *sim.World) []hexmap.ChunkKey {
	var out []hexmap.ChunkKey
	query := ecs.NewFilter2[comp.MemberOf, comp.Pos](w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		out = append(out, hexmap.KeyOf(pos.At))
	}
	return out
}
