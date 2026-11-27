package explore

import (
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Pocket is the fixed pocket of an export as plain data, so a save can keep it.
type Pocket struct {
	Chunks   []PocketChunk
	Openings []api.Cell
}

// PocketChunk is one pocket chunk: its terrain in slot order as runs.
type PocketChunk struct {
	Key  hexmap.ChunkKey
	Runs []PocketRun
}

// PocketRun is Count consecutive slots of one terrain.
type PocketRun struct {
	Terrain api.ContentID
	Count   int
}

// slotCoords is the local coordinates of a slot, the inverse of the chunk's own indexing.
func slotCoords(slot int) (q, r, depth int) {
	return slot % hexmap.ChunkSize, slot / hexmap.ChunkSize % hexmap.ChunkSize, slot / (hexmap.ChunkSize * hexmap.ChunkSize)
}

// SavedPocket returns the world's pocket, zero before the first export. Chunks are in key order.
func SavedPocket(w *sim.World) Pocket {
	st := stateOf(w)
	if st.pocket == nil {
		return Pocket{}
	}
	keys := make([]hexmap.ChunkKey, 0, len(st.pocket.chunks))
	for key := range st.pocket.chunks {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, hexmap.CompareKeys)
	out := Pocket{Openings: slices.Clone(st.pocket.footprint.Openings)}
	for _, key := range keys {
		chunk := st.pocket.chunks[key]
		saved := PocketChunk{Key: key}
		for slot := range hexmap.ChunkCells {
			q, r, depth := slotCoords(slot)
			id := chunk.Get(q, r, depth).ID
			if last := len(saved.Runs) - 1; last >= 0 && saved.Runs[last].Terrain == id {
				saved.Runs[last].Count++
			} else {
				saved.Runs = append(saved.Runs, PocketRun{Terrain: id, Count: 1})
			}
		}
		out.Chunks = append(out.Chunks, saved)
	}
	return out
}

// fixed rebuilds the pocket's fixed layer over base, resolving terrain from content.
func (p Pocket) fixed(w *sim.World, base *WorldSource) (*pocketFixed, error) {
	fixed := &pocketFixed{
		Fixed:     base,
		chunks:    make(map[hexmap.ChunkKey]*hexmap.Chunk, len(p.Chunks)),
		footprint: hexmap.Footprint{Place: pocketPlace, Openings: slices.Clone(p.Openings)},
	}
	for _, saved := range p.Chunks {
		if len(saved.Runs) == 0 {
			return nil, fmt.Errorf("pocket chunk %v has no cells", saved.Key)
		}
		fill, err := Terrain(w.Content, saved.Runs[0].Terrain)
		if err != nil {
			return nil, fmt.Errorf("pocket chunk %v: %w", saved.Key, err)
		}
		chunk := hexmap.NewChunk(fill)
		slot := 0
		for _, run := range saved.Runs {
			t, err := Terrain(w.Content, run.Terrain)
			if err != nil {
				return nil, fmt.Errorf("pocket chunk %v: %w", saved.Key, err)
			}
			for range run.Count {
				if slot >= hexmap.ChunkCells {
					return nil, fmt.Errorf("pocket chunk %v has more than %d cells", saved.Key, hexmap.ChunkCells)
				}
				q, r, depth := slotCoords(slot)
				chunk.Set(q, r, depth, t)
				slot++
			}
		}
		if slot != hexmap.ChunkCells {
			return nil, fmt.Errorf("pocket chunk %v has %d cells, want %d", saved.Key, slot, hexmap.ChunkCells)
		}
		fixed.chunks[saved.Key] = chunk
	}
	return fixed, nil
}
