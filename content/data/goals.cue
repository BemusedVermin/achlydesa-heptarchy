package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

goals: [string]: content.#Goal

// An agent weighs its goals every think_ticks ticks. A goal's weight is the drive's urge times the
// individual's weight on it; wander is the weakest and always applies. First values, for tuning.
agents: {
	think_ticks: 4
	keep_bonus:  0.1
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
	"goal.wander": {
		weight: "0.05"
	}
}
