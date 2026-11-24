// Command cuefix fixes the CUE files "cue get go" generates, so they resolve and match the JSON.
//
//	cuefix DIR...
//
// It lower-cases this module's Go path in their imports, because CUE module paths must be lower
// case, and turns the fields written for embedded Go structs into CUE embeddings, so the embedded
// fields sit at the level of the embedding struct, as encoding/json has them. It defines each named
// type that has constants as the disjunction of them ("#Dir: #enumDir"), where cue get go only names
// the set in a comment, so a value outside the set fails the schema. Run by "task schema".
package main
