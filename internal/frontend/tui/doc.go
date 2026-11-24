// Package tui is the terminal frontend: a Bubble Tea program that draws the core's view and sends
// only the commands the view offers.
//
// The root [App] owns a stack of [Screen]s. There is one map screen for exploring and fighting; the
// stack exists for overlays such as the timeline popup, and later the journal and dreams. Rendering
// lives in pure functions ([RenderMap], [RenderTimeline], ...) so golden tests can cover it.
//
// Every word of game text comes rendered in the view; only key help is written here. From the core,
// the package imports only game and api: hot reload reaches it as an opaque function over the game.
package tui
