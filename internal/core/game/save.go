package game

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	arkserde "github.com/mlange-42/ark-serde"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/save"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Save writes the whole game as one versioned document. Equal games write equal bytes.
func (g *Game) Save(w io.Writer) error {
	if g.broken != nil {
		return g.broken
	}
	snapshot, err := g.snapshot()
	if err != nil {
		return err
	}
	return save.Write(w, snapshot)
}

func (g *Game) snapshot() (save.Snapshot, error) {
	// Resources hold hooks and caches that content and Install rebuild, so only entities are saved.
	// The timeline, the one resource with state of its own, is exported by timeline.
	world, err := arkserde.Serialize(g.world.ECS, arkserde.Opts.SkipAllResources())
	if err != nil {
		return save.Snapshot{}, fmt.Errorf("game: save: world: %w", err)
	}
	if world, err = canonical(world); err != nil {
		return save.Snapshot{}, fmt.Errorf("game: save: world: %w", err)
	}
	streams, err := g.world.RNG.State()
	if err != nil {
		return save.Snapshot{}, fmt.Errorf("game: save: %w", err)
	}
	history := make([]json.RawMessage, 0, len(g.history))
	for i, cmd := range g.history {
		data, err := g.world.Codec.EncodeCommand(cmd)
		if err != nil {
			return save.Snapshot{}, fmt.Errorf("game: save: history command %d: %w", i, err)
		}
		history = append(history, data)
	}
	edits := map[api.MapID][]save.MapEdit{}
	for _, id := range g.world.Maps.IDs() {
		for _, edit := range g.world.Maps.Get(id).Edits() {
			edits[id] = append(edits[id], save.MapEdit{At: edit.Cell, Terrain: edit.Terrain.ID})
		}
	}
	return save.Snapshot{
		Header: save.Header{
			Format: save.Format, Version: save.CurrentVersion, ContentHash: g.world.Content.Hash(),
			Seed: g.world.RNG.Seed(), Tick: g.world.Now,
		},
		World:    world,
		RNG:      streams,
		NextUID:  g.world.NextUID(),
		Timeline: timeline.Export(g.world),
		MapEdits: edits,
		History:  history,
		Log:      append([]api.LogLine{}, g.log...),
	}, nil
}

// Load reads a save into a game that carries on as if it had never stopped. cfg.Seed is ignored: the
// save's seed is used. It builds the game the way New does but skips the start scenario, so
// cfg.Spawn is not run; cfg.Install is. A save made with different content
// still loads; the warnings say so.
func Load(cfg Config, r io.Reader) (*Game, []string, error) {
	snapshot, err := save.Read(r)
	if err != nil {
		return nil, nil, err
	}
	cfg.Seed = snapshot.Header.Seed
	g, err := build(cfg, explore.LoadMaps, nil)
	if err != nil {
		return nil, nil, err
	}
	if err := g.restore(snapshot); err != nil {
		return nil, nil, err
	}
	return g, snapshot.Header.CheckContent(cfg.Content.Hash()), nil
}

func (g *Game) restore(snapshot save.Snapshot) error {
	w := g.world
	if err := arkserde.Deserialize(snapshot.World, w.ECS, arkserde.Opts.SkipAllResources()); err != nil {
		return fmt.Errorf("game: load: world (does Config.Install spawn entities?): %w", err)
	}
	if err := w.ReindexUIDs(); err != nil {
		return fmt.Errorf("game: load: %w", err)
	}
	w.SetNextUID(snapshot.NextUID)
	w.Now = snapshot.Header.Tick
	if err := w.RNG.Restore(snapshot.RNG); err != nil {
		return fmt.Errorf("game: load: %w", err)
	}
	if err := applyEdits(w, snapshot.MapEdits); err != nil {
		return fmt.Errorf("game: load: %w", err)
	}
	if err := timeline.Import(w, snapshot.Timeline); err != nil {
		return fmt.Errorf("game: load: %w", err)
	}
	g.history = make([]api.Command, 0, len(snapshot.History))
	for i, data := range snapshot.History {
		cmd, err := w.Codec.DecodeCommand(data)
		if err != nil {
			return fmt.Errorf("game: load: history command %d: %w", i, err)
		}
		g.history = append(g.history, cmd)
	}
	g.log = snapshot.Log
	return nil
}

// applyEdits reapplies saved map edits, resolving each terrain from content.
func applyEdits(w *sim.World, edits map[api.MapID][]save.MapEdit) error {
	for id, list := range edits {
		m := w.Maps.Get(id)
		if m == nil {
			return fmt.Errorf("saved edits for unknown map %d", id)
		}
		for _, edit := range list {
			terrain, err := explore.Terrain(w.Content, edit.Terrain)
			if err != nil {
				return fmt.Errorf("saved edit: %w", err)
			}
			if err := m.Set(edit.At, terrain); err != nil {
				return err
			}
		}
	}
	return nil
}

// canonical rewrites ark-serde's JSON so that equal worlds give equal bytes. ark-serde writes the
// component types and each entity's components in Go map order, and lists the entities in the order
// of the ECS's archetypes, which differs between a played world and a loaded one. So this sorts the
// type names, sorts the entities by ECS ID (keeping each one's components with it), and lets
// encoding/json sort every object's keys. Numbers pass through untouched.
func canonical(data []byte) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var dump struct {
		World      json.RawMessage
		Types      []string
		Components []json.RawMessage
		Resources  json.RawMessage
	}
	if err := decoder.Decode(&dump); err != nil {
		return nil, err
	}
	var pool struct {
		Alive []uint32
	}
	if err := json.Unmarshal(dump.World, &pool); err != nil {
		return nil, err
	}
	if len(pool.Alive) != len(dump.Components) {
		return nil, fmt.Errorf("%d entities but %d component sets", len(pool.Alive), len(dump.Components))
	}
	order := make([]int, len(pool.Alive))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int { return cmp.Compare(pool.Alive[a], pool.Alive[b]) })
	alive := make([]uint32, len(order))
	components := make([]json.RawMessage, len(order))
	for to, from := range order {
		alive[to], components[to] = pool.Alive[from], dump.Components[from]
	}
	slices.Sort(dump.Types)
	var world map[string]any
	worldDecoder := json.NewDecoder(bytes.NewReader(dump.World))
	worldDecoder.UseNumber()
	if err := worldDecoder.Decode(&world); err != nil {
		return nil, err
	}
	world["Alive"] = alive
	return normalize(map[string]any{
		"World": world, "Types": dump.Types, "Components": components, "Resources": dump.Resources,
	})
}

// normalize encodes v with every object's keys sorted: it decodes the encoding into generic maps and
// encodes that again.
func normalize(v any) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var generic any
	if err := decoder.Decode(&generic); err != nil {
		return nil, err
	}
	return json.Marshal(generic)
}
