// Package worldgen builds the generated deep: a [Source] runs a geography's named stages on each chunk, from
// the world seed, the export count and the chunk key alone, around the fixed places. Foundation layer: it
// reads content and draws from [rng] only, so any package may import it. What reacts to the game's state, such
// as reseeding on an export and keeping the party's pocket, belongs to the feature package procgen.
package worldgen
