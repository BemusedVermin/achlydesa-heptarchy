package sim

import (
	"sync/atomic"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
)

// World is the whole simulation: the Ark ECS world plus the registries every
// package plugs into.
type World struct {
	ECS      *ecs.World
	RNG      *rng.Streams
	Bus      *Bus
	Systems  *Scheduler
	Commands *Commands
	Views    *Views
	Codec    *api.Codec
	// Map is the world's cells: nil until explore builds it from content.
	Map *hexmap.Map
	// Content is the loaded content; game.New sets it.
	Content *content.Registry
	// Now is the current tick; only clock writes it.
	Now     int64
	nextUID atomic.Uint64

	// Spawning state belongs to the world's one writer, like the ECS.
	maps     *mappers
	spawners []Spawner
	byUID    map[api.UID]ecs.Entity
}

// NewWorld returns an empty world whose random streams derive from seed.
func NewWorld(seed uint64) *World {
	w := &World{
		ECS:      ecs.NewWorld(),
		RNG:      rng.New(seed),
		Bus:      NewBus(),
		Systems:  NewScheduler(),
		Commands: NewCommands(),
		Views:    NewViews(),
		Codec:    api.NewCodec(),
	}
	w.nextUID.Store(1)
	comp.Register(w.ECS)
	w.maps = newMappers(w.ECS)
	w.byUID = map[api.UID]ecs.Entity{}
	w.registerCoreSpawners()
	return w
}

// NewUID returns the next unused UID: 1, 2, 3, ... Never reused. Safe for concurrent use.
func (w *World) NewUID() api.UID {
	return api.UID(w.nextUID.Add(1) - 1)
}

// NextUID returns the UID NewUID will hand out next, for saves.
func (w *World) NextUID() api.UID { return api.UID(w.nextUID.Load()) }

// SetNextUID sets the next UID, for loading saves.
func (w *World) SetNextUID(uid api.UID) { w.nextUID.Store(uint64(uid)) }
