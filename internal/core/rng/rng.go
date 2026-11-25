package rng

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
)

// Generator draws numbers from a sequence that [Streams.Derived] fixed by its inputs. It is not a stream: it is
// not registered or saved, and its methods are the only draws the core may make from it.
type Generator struct {
	rand *rand.Rand
}

// Uint64 returns the next number in the sequence.
func (g *Generator) Uint64() uint64 { return g.rand.Uint64() }

// IntN returns a number in [0, n); n must be positive.
func (g *Generator) IntN(n int) int { return g.rand.IntN(n) }

// Float64 returns a number in [0, 1).
func (g *Generator) Float64() float64 { return g.rand.Float64() }

// stream pairs a generator with the PCG source behind it, so the source's
// state can be saved and restored.
type stream struct {
	source *rand.PCG
	rand   *rand.Rand
}

// Streams hands out one PCG stream per name, all derived from the world seed.
// It is not safe for concurrent use: only the goroutine that changes the world
// may draw from it.
type Streams struct {
	seed    uint64
	streams map[string]*stream
}

// New returns the streams for a world seed.
func New(worldSeed uint64) *Streams {
	return &Streams{seed: worldSeed, streams: make(map[string]*stream)}
}

// Seed returns the world seed the streams derive from.
func (s *Streams) Seed() uint64 { return s.seed }

// Get returns the named stream, creating it on first use. The sequence depends
// only on the world seed and the name, never on the order streams were created in.
func (s *Streams) Get(name string) *rand.Rand {
	if existing, ok := s.streams[name]; ok {
		return existing.rand
	}
	return s.add(name, rand.NewPCG(s.seed, fnv64a(name))).rand
}

// Derived returns a fresh generator for (name, salt) that is not registered or
// saved: callers recreate it from the same inputs. Procgen uses ("procgen", exportCount).
func (s *Streams) Derived(name string, salt uint64) *Generator {
	return &Generator{rand: rand.New(rand.NewPCG(splitmix64(s.seed^salt), fnv64a(name)))}
}

// State returns the serialized state of every stream created with Get, keyed by name.
func (s *Streams) State() (map[string][]byte, error) {
	state := make(map[string][]byte, len(s.streams))
	for name, st := range s.streams {
		data, err := st.source.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("rng: marshal stream %q: %w", name, err)
		}
		state[name] = data
	}
	return state, nil
}

// Restore replaces the streams with those in state, as returned by State.
// Streams not in state are dropped and will restart from their seed on next use.
func (s *Streams) Restore(state map[string][]byte) error {
	restored := make(map[string]*stream, len(state))
	for name, data := range state {
		source := rand.NewPCG(0, 0)
		if err := source.UnmarshalBinary(data); err != nil {
			return fmt.Errorf("rng: restore stream %q: %w", name, err)
		}
		restored[name] = &stream{source: source, rand: rand.New(source)}
	}
	s.streams = restored
	return nil
}

func (s *Streams) add(name string, source *rand.PCG) *stream {
	st := &stream{source: source, rand: rand.New(source)}
	s.streams[name] = st
	return st
}

// fnv64a hashes a stream name with FNV-1a 64. It is stable across processes,
// unlike hash/maphash.
func fnv64a(name string) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(name))
	return hasher.Sum64()
}

// splitmix64 is the standard SplitMix64 finalizer.
func splitmix64(x uint64) uint64 {
	x += 0x9E3779B97F4A7C15
	x = (x ^ (x >> 30)) * 0xBF58476D1CE4E5B9
	x = (x ^ (x >> 27)) * 0x94D049BB133111EB
	return x ^ (x >> 31)
}
