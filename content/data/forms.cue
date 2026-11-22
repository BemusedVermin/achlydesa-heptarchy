package data

// The seven forms. Numbers are placeholders for M7's tuning, but they are lethal from the start: a
// human part has 3 integrity (5 for the torso), so an ordinary Strike takes off an arm in one
// hit. Every attack is sight: true; connects is the default (still in range and sight).
// "weak" is true when the target part's material is weak to the chosen type.
actions: {
	"action.strike": {
		name: "text.action.strike"
		key:  "strike"
		uses: [{tag: "grant:action", count: 1}]
		target:   "part"
		range:    1
		sight:    true
		choices: ["type"]
		windup:   "3"
		recovery: "4"
		on_land: [{do: "damage", params: {parts: "target_part", amount: "weak ? 99 : 3"}}]
		tags: ["attack"]
	}
	"action.sweep": {
		name: "text.action.sweep"
		key:  "sweep"
		uses: [{tag: "grant:action", count: 2}]
		target:   "body"
		range:    1
		sight:    true
		choices: ["type"]
		windup:   "4"
		recovery: "5"
		on_land: [
			{do: "damage", params: {parts: "target:all", amount: "weak ? 2 : 1"}},
			{do: "reveal", params: {claim: "claim.material"}},
		]
		tags: ["attack"]
	}
	"action.sever": {
		name: "text.action.sever"
		key:  "sever"
		uses: [{tag: "grant:action", count: 1}]
		target:   "part"
		range:    1
		sight:    true
		choices: ["type"]
		windup:   "2"
		recovery: "5"
		on_land: [
			{do: "damage", params: {parts: "target_part", amount: "1"}},
			{do: "sever", params: {parts: "target_part", when: "weak"}},
		]
		tags: ["attack"]
	}
	"action.pin": {
		name: "text.action.pin"
		key:  "pin"
		uses: [{tag: "grant:action", count: 1}]
		target:   "part"
		range:    1
		sight:    true
		choices: ["type"]
		windup:   "1"
		recovery: "3"
		on_land: [{do: "lock", params: {channels: "target_part", ticks: "8"}}]
		tags: ["attack"]
	}
	// Stand braces the body, freezes every crawling body in sight and readies the body at once. A body can't stand
	// again while still braced, or it could freeze the world for ever. Recovery is 1, not 0: an action
	// that takes no time is refused (actions.too_quick), and the ready effect frees the channel anyway.
	// The canon: "anyone who stands freezes everyone who crawls". Bodies that crawl carry the crawls
	// marker (a template's combatant.crawls, or a part that got up).
	"action.stand": {
		name: "text.action.stand"
		key:  "stand"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		requires: "\"status.standing\" not in statuses"
		windup:   "0"
		recovery: "1"
		on_land: [
			{do: "add_status", params: {bodies: "self", status: "status.standing", ticks: "10"}},
			{do: "delay", params: {channels: "crawling:all", ticks: "8"}},
			{do: "ready", params: {channels: "self:all", count: "99"}},
		]
		tags: ["stance"]
	}
	"action.give": {
		name: "text.action.give"
		key:  "give"
		uses: [{tag: "grant:action", count: 1}]
		target:   "ally"
		range:    1
		sight:    true
		windup:   "0"
		recovery: "6"
		on_land: [{do: "ready", params: {channels: "target:all", count: "1"}}]
		tags: ["support"]
	}
	// Graft fits a piece from the pack onto a graftable ally: one arm holds it, the other fits it.
	"action.graft": {
		name: "text.action.graft"
		key:  "graft"
		uses: [{tag: "grant:action", count: 2}]
		target:   "ally"
		range:    1
		sight:    true
		choices: ["piece"]
		windup:   "6"
		recovery: "6"
		on_land: [{do: "attach_part", params: {from: "pack", item: "choice.piece"}}]
		tags: ["support", "graft"]
	}
}

statuses: "status.standing": name: "text.status.standing"

text: {
	"text.action.strike":         "Strike"
	"text.action.sweep":          "Sweep"
	"text.action.sever":          "Sever"
	"text.action.pin":            "Pin"
	"text.action.stand":          "Stand"
	"text.action.give":           "Give"
	"text.action.graft":          "Graft"
	"text.status.standing":       "standing"
	"text.combat.immune":         "That part is immune to {{.type}}."
	"text.combat.type_barred":    "{{.type}} cannot be used on that."
	"text.combat.not_graftable":  "Nothing can be grafted onto that."
}
