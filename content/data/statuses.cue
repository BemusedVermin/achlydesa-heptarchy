package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// Statuses a body can be under for a while. The forms that grant them are in forms.cue.
statuses: [string]: content.#Status

statuses: {
	// Granted by Stand. It carries no rules of its own yet.
	"status.standing": {name: "text.status.standing"}
	// Put on by hunger and thirst. rule.slowed (rules.cue) makes steps and wind-ups longer.
	"status.slowed": {name: "text.status.slowed"}
}

text: {
	"text.status.standing": "standing"
	"text.status.slowed":   "slowed"
}
