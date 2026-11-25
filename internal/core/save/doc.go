// Package save is the save file: one versioned JSON document holding the world, the random streams,
// the UID counter, the world's edits and the command history. It knows the file format and its migrations;
// game does the work of capturing and restoring a world. Top layer.
package save
