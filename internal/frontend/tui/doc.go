// Package tui is the terminal frontend: a Bubble Tea program that draws the core's view and sends
// only the commands the view offers.
//
// The root [App] owns a stack of [Screen]s. There is one map screen for exploring and fighting; the
// stack exists for overlays such as the timeline popup, and later the journal and dreams. Rendering
// lives in pure functions ([RenderMap], [RenderTimeline], ...) so golden tests can cover it.
//
// The package imports only the core's game, api and content packages.
package tui
