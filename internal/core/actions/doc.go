// Package actions runs the actions content defines: a step, a two-handed swing, a graft that takes
// both hands. An action names the channels it occupies, how long it winds up and recovers, what it
// targets and which effects apply when it starts and lands. This package supplies the machinery:
// validating an intent, starting it, landing it later, and listing every intent a body could start
// as the options the player chooses from. Combat rules stay in combat; they plug in through
// AddValidator, AddModifier and RegisterChoice.
package actions
