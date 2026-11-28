package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

goals: [string]: content.#Goal

// An agent weighs its goals every think_ticks ticks. A goal's weight is the drive's urge times the
// individual's weight on it; wander is the weakest and always applies. First values, for tuning.
agents: {
	think_ticks: 4
	keep_bonus:  0.1
	// An agent far from the party decides this often instead: about a tenth of a phase.
	far_ticks: 6
}

goals: {
	"goal.drink": {
		drive: "drive.water"
		want: [{fact: "drive_met", tag: "drive.water"}]
		when:   "!met[\"drive.water\"]"
		weight: "urge * weight"
	}
	"goal.eat": {
		drive: "drive.food"
		want: [{fact: "drive_met", tag: "drive.food"}]
		when:   "!met[\"drive.food\"]"
		weight: "urge * weight"
	}
	"goal.sleep": {
		drive: "drive.rest"
		want: [{fact: "drive_met", tag: "drive.rest"}]
		when:   "!met[\"drive.rest\"]"
		weight: "urge * weight"
	}
	"goal.hold_dead": {
		drive: "drive.the_dead"
		want: [{fact: "drive_met", tag: "drive.the_dead"}]
		when:   "!met[\"drive.the_dead\"]"
		weight: "urge * weight"
	}
	"goal.old_ground": {
		drive: "drive.old_ground"
		want: [{fact: "drive_met", tag: "drive.old_ground"}]
		when:   "!met[\"drive.old_ground\"]"
		weight: "urge * weight"
	}
	"goal.wander": {
		weight: "0.05"
	}
	// Tell a near agent what it does not know. Only applies when one is in reach with something to
	// hear; weaker than any drive, stronger than wandering.
	"goal.chat": {
		weight: "0.08"
	}
	// Fight a hostile believed near. Weighed against the drives like any goal.
	"goal.fight": {
		when:   "hostile"
		weight: "0.5"
	}
	// Run from it when safety, which is terror, is low.
	"goal.flee": {
		drive: "drive.safety"
		want: [{fact: "drive_met", tag: "drive.safety"}]
		when:   "hostile && !met[\"drive.safety\"]"
		weight: "urge * weight"
	}
}
