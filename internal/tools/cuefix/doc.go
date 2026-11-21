// Command cuefix rewrites the Go module path in generated CUE files to lower case.
//
// CUE module paths must be lower case, but this module's Go path is not, so the
// imports that "cue get go" writes would not resolve. Run by "task schema".
package main
