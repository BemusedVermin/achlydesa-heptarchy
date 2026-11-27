package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

drives: [string]: content.#Drive

// A level counts roughly the phases a drive has gone unmet; a day is four phases. First values, for
// tuning. A need's thresholds harm the vital torso every phase they hold, so going without kills,
// slowly.
drives: {
	"drive.water": {
		name: "text.drive.water"
		kind: "need"
		rise: "1"
		urge: "min(level / 12, 1)"
		thresholds: [{at: 12, do: [{do: "damage", params: {parts: "self:torso", amount: "1"}}]}]
	}
	"drive.food": {
		name: "text.drive.food"
		kind: "need"
		rise: "1"
		urge: "min(level / 40, 1)"
		thresholds: [{at: 40, do: [{do: "damage", params: {parts: "self:torso", amount: "1"}}]}]
	}
	"drive.rest": {
		name: "text.drive.rest"
		kind: "need"
		rise: "1"
		urge: "min(level / 8, 1)"
		thresholds: [{at: 24, do: [{do: "damage", params: {parts: "self:torso", amount: "1"}}]}]
	}
	// TODO(M6-03.F1): decide what raises safety; time alone does not, so it rises by nothing yet.
	"drive.safety": {
		name: "text.drive.safety"
		kind: "need"
		rise: "0"
		urge: "min(level / 4, 1)"
	}
	// Raised by going unpaid, once something pays and hires.
	"drive.pay": {
		name: "text.drive.pay"
		kind: "desire"
		urge: "min(level / 10, 1)"
	}
}

// An individual's weight lies between half and one and a half times its template's.
agents: weight: "typical * (0.5 + roll)"

text: {
	"text.drive.water":  "water"
	"text.drive.food":   "food"
	"text.drive.rest":   "rest"
	"text.drive.safety": "safety"
	"text.drive.pay":    "pay"
}
