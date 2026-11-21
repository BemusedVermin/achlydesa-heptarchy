package sim

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
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
	// Now is the current tick; only clock writes it.
	Now     int64
	nextUID api.UID
}

// NewWorld returns an empty world whose random streams derive from seed.
func NewWorld(seed uint64) *World {
	return &World{
		ECS:      ecs.NewWorld(),
		RNG:      rng.New(seed),
		Bus:      NewBus(),
		Systems:  NewScheduler(),
		Commands: NewCommands(),
		Views:    NewViews(),
		Codec:    api.NewCodec(),
		nextUID:  1,
	}
}

// NewUID returns the next unused UID: 1, 2, 3, ... Never reused.
func (w *World) NewUID() api.UID {
	uid := w.nextUID
	w.nextUID++
	return uid
}

// NextUID returns the UID NewUID will hand out next, for saves.
func (w *World) NextUID() api.UID { return w.nextUID }

// SetNextUID sets the next UID, for loading saves.
func (w *World) SetNextUID(uid api.UID) { w.nextUID = uid }
