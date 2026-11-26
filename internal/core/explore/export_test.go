package explore_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

var exportRules = hexmap.StepRules{MaxStepUp: 1, MaxStepDown: 1, MaxDrop: 1}

type exportOutcome struct {
	terrain   []api.ContentID
	positions []api.Cell
}

// exportWorld is a small generated world with the party standing in the deep, away from every place.
type exportWorld struct {
	t        *testing.T
	w        *sim.World
	g        *game.Game
	cfg      game.Config
	src      *explore.WorldSource
	cells    []api.Cell
	home     hexmap.ChunkKey
	deep     hexmap.ChunkKey
	creature ecs.Entity
}

func newExportWorld(t *testing.T, seed uint64) *exportWorld {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.StepOnlyParty(&bundle)
	for id, def := range bundle.Places {
		def.Spawns = nil
		bundle.Places[id] = def
	}
	bundle.World.Geography = "geography.deep"
	bundle.World.Cols, bundle.World.Rows, bundle.World.MaxDepth = 100, 40, 7
	bundle.World.PocketExits = 4
	bundle.World.PocketChunks = 0
	reg := contenttest.Registry(t, bundle)
	g, w, cfg := gametest.Start(t, game.Config{Seed: seed, Content: reg})
	src, err := explore.NewWorldSource(reg)
	if err != nil {
		t.Fatal(err)
	}
	ew := &exportWorld{t: t, w: w, g: g, cfg: cfg, src: src}
	bounds := src.Bounds()
	for depth := bounds.MinDepth; depth <= bounds.MaxDepth; depth++ {
		for row := 0; row <= bounds.MaxRow; row++ {
			for col := 0; col <= bounds.MaxCol; col++ {
				ew.cells = append(ew.cells, api.Cell{Hex: hexmap.FromOffsetOddR(col, row), Depth: depth})
			}
		}
	}
	slices.SortFunc(ew.cells, hexmap.CompareCells)
	ew.placeMembersInDeep()
	return ew
}

func (ew *exportWorld) inPlaceChunk(key hexmap.ChunkKey) bool {
	return slices.Contains(ew.src.ChunkKeys(), key)
}

// standableIn lists the standable cells of chunks that hold no place, other than skip.
func (ew *exportWorld) standableIn(skip hexmap.ChunkKey) map[hexmap.ChunkKey][]api.Cell {
	out := map[hexmap.ChunkKey][]api.Cell{}
	for _, cell := range ew.cells {
		key := hexmap.KeyOf(cell)
		if key == skip || ew.inPlaceChunk(key) {
			continue
		}
		ok, err := ew.w.Map.Standable(cell)
		if err != nil {
			ew.t.Fatal(err)
		}
		if ok {
			out[key] = append(out[key], cell)
		}
	}
	return out
}

func (ew *exportWorld) placeMembersInDeep() {
	ew.t.Helper()
	candidates := ew.standableIn(hexmap.ChunkKey{Q: 1 << 20})
	keys := make([]hexmap.ChunkKey, 0, len(candidates))
	for key, cells := range candidates {
		if len(cells) >= 4 {
			keys = append(keys, key)
		}
	}
	slices.SortFunc(keys, hexmap.CompareKeys)
	if len(keys) < 2 {
		ew.t.Skipf("only %d deep chunks have room to stand", len(keys))
	}
	ew.home, ew.deep = keys[0], keys[1]
	positions := ecs.NewMap[comp.Pos](ew.w.ECS)
	query := ecs.NewFilter1[comp.MemberOf](ew.w.ECS).Query()
	var members []ecs.Entity
	for query.Next() {
		members = append(members, query.Entity())
	}
	slices.SortFunc(members, func(a, b ecs.Entity) int { return int(ew.w.UIDOf(a)) - int(ew.w.UIDOf(b)) })
	for i, member := range members {
		positions.Get(member).At = candidates[ew.home][i]
	}
	creature, err := ew.w.Spawn("template.creature.moving_clay", &comp.Pos{At: candidates[ew.deep][0]})
	if err != nil {
		ew.t.Fatal(err)
	}
	ew.creature = creature
	seen := ecs.NewMap[comp.Seen](ew.w.ECS)
	query2 := ecs.NewFilter1[comp.Seen](ew.w.ECS).Query()
	var parties []ecs.Entity
	for query2.Next() {
		parties = append(parties, query2.Entity())
	}
	for _, party := range parties {
		chunk := comp.SeenChunk{
			Q:     ew.deep.Q,
			R:     ew.deep.R,
			Depth: ew.deep.Depth,
			Bits:  make([]uint64, hexmap.ChunkCells/64),
		}
		chunk.Bits[0] = 1
		known := seen.Get(party)
		known.Chunks = append(known.Chunks, chunk)
		slices.SortFunc(known.Chunks, func(a, b comp.SeenChunk) int {
			return hexmap.CompareKeys(
				hexmap.ChunkKey{Q: a.Q, R: a.R, Depth: a.Depth},
				hexmap.ChunkKey{Q: b.Q, R: b.R, Depth: b.Depth},
			)
		})
	}
}

func (ew *exportWorld) terrain() []api.ContentID {
	ew.t.Helper()
	out := make([]api.ContentID, len(ew.cells))
	for i, cell := range ew.cells {
		terrain, err := ew.w.Map.At(cell)
		if err != nil {
			ew.t.Fatal(err)
		}
		out[i] = terrain.ID
	}
	return out
}

func (ew *exportWorld) bodies() []api.Cell {
	var out []api.Cell
	query := ecs.NewFilter1[comp.Pos](ew.w.ECS).Query()
	byUID := map[api.UID]api.Cell{}
	for query.Next() {
		byUID[ew.w.UIDOf(query.Entity())] = query.Get().At
	}
	uids := make([]api.UID, 0, len(byUID))
	for uid := range byUID {
		uids = append(uids, uid)
	}
	slices.Sort(uids)
	for _, uid := range uids {
		out = append(out, byUID[uid])
	}
	return out
}

func (ew *exportWorld) routeToPlace(from api.Cell) bool {
	seen := map[api.Cell]bool{from: true}
	frontier := []api.Cell{from}
	for len(frontier) > 0 {
		cell := frontier[0]
		frontier = frontier[1:]
		if ew.src.Covers(cell) {
			return true
		}
		for dir := range api.Dir(api.DirCount) {
			step, can, err := ew.w.Map.Step(cell, dir, exportRules)
			if err != nil {
				ew.t.Fatal(err)
			}
			if to := step.To.Cell; can && !seen[to] {
				seen[to] = true
				frontier = append(frontier, to)
			}
		}
	}
	return false
}

// runExport exports with the party below and checks what the export promises.
func runExport(t *testing.T, seed uint64) exportOutcome {
	t.Helper()
	ew := newExportWorld(t, seed)
	before := ew.terrain()
	ew.w.ExportCount = 1
	ew.w.Bus.Publish(api.ExportBegan{Day: 30, Count: 1})
	if _, err := ew.w.Bus.Drain(ew.w); err != nil {
		t.Fatal(err)
	}
	after := ew.terrain()

	changed := 0
	for i, cell := range ew.cells {
		key := hexmap.KeyOf(cell)
		switch {
		case key == ew.home:
			if before[i] != after[i] {
				t.Fatalf("pocket cell %v changed from %q to %q", cell, before[i], after[i])
			}
		case ew.src.Covers(cell):
			if before[i] != after[i] {
				t.Fatalf("place cell %v changed from %q to %q", cell, before[i], after[i])
			}
		case before[i] != after[i]:
			changed++
		}
	}
	if changed == 0 {
		t.Error("the deep around the pocket did not change")
	}

	query := ecs.NewFilter2[comp.MemberOf, comp.Pos](ew.w.ECS).Query()
	if !query.Next() {
		t.Fatal("no party member")
	}
	_, pos := query.Get()
	start := pos.At
	query.Close()
	if !ew.routeToPlace(start) {
		t.Errorf("no route from the party at %v to any place", start)
	}

	for _, at := range ew.bodies() {
		if ok, err := ew.w.Map.Standable(at); err != nil || !ok {
			t.Errorf("body at %v is not on open ground (err %v)", at, err)
		}
	}

	seenQuery := ecs.NewFilter1[comp.Seen](ew.w.ECS).Query()
	for seenQuery.Next() {
		for _, chunk := range seenQuery.Get().Chunks {
			if key := (hexmap.ChunkKey{Q: chunk.Q, R: chunk.R, Depth: chunk.Depth}); key == ew.deep {
				t.Errorf("chunk %v is still seen after regenerating", key)
			}
		}
	}
	return exportOutcome{terrain: after, positions: ew.bodies()}
}

func TestExportRegeneratesTheDeepAroundThePocket(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		first := runExport(t, seed)
		second := runExport(t, seed)
		if !slices.Equal(first.terrain, second.terrain) || !slices.Equal(first.positions, second.positions) {
			t.Errorf("seed %d: two runs of the same export differ", seed)
		}
	}
}

// TestExportedWorldSurvivesSaveAndLoad checks that a save after an export loads with the same ground and hash.
func TestExportedWorldSurvivesSaveAndLoad(t *testing.T) {
	ew := newExportWorld(t, 1)
	ew.w.ExportCount = 1
	ew.w.Bus.Publish(api.ExportBegan{Day: 30, Count: 1})
	if _, err := ew.w.Bus.Drain(ew.w); err != nil {
		t.Fatal(err)
	}
	wantHash, err := ew.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	want := ew.terrain()
	loaded, world, _ := gametest.Reload(t, ew.cfg, ew.g)
	reloaded := &exportWorld{t: t, w: world, cells: ew.cells}
	if got := reloaded.terrain(); !slices.Equal(got, want) {
		t.Error("the map differs after a load")
	}
	gotHash, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if gotHash != wantHash {
		t.Error("Game.Hash differs after a load")
	}
}
