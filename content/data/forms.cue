package data

// The seven forms. Numbers are placeholders for M7's tuning, but they are lethal from the start: a
// human part has 3 integrity (5 for the torso), so an ordinary Strike takes off an arm in one hit.
// Every attack has sight: true, and connects is the default (still in range and in sight).
// An attack form is loaded with a type (choices.cue); what the type does to the form is in types.cue
// and the rules that read it (weak, immune, hit, graze) are in rules.cue.
actions: {
	"action.strike": {
		name: "text.action.strike"
		key:  "strike"
		uses: [{tag: "grant:action", count: 1}]
		target: "part"
		range:  1
		sight:  true
		choices: ["type"]
		windup:   "3"
		recovery: "4"
		on_land: [{do: "damage", params: {parts: "target_part", amount: "rule.hit"}}]
		tags: ["attack"]
	}
	// Sweep hits every part of the body, each by its own material, and teaches what each is made of.
	"action.sweep": {
		name: "text.action.sweep"
		key:  "sweep"
		uses: [{tag: "grant:action", count: 2}]
		target: "body"
		range:  1
		sight:  true
		choices: ["type"]
		windup:   "4"
		recovery: "5"
		on_land: [
			{do: "damage", params: {parts: "target:all", amount: "rule.graze"}},
			{do: "reveal", params: {
				parts: "target:all", claim: "claim.material", value: "= target.material.id", source: "source.sweep"
			}},
		]
		tags: ["attack"]
	}
	// Sever wears a part by one and cuts it off outright where it is weak.
	"action.sever": {
		name: "text.action.sever"
		key:  "sever"
		uses: [{tag: "grant:action", count: 1}]
		target: "part"
		range:  1
		sight:  true
		choices: ["type"]
		windup:   "2"
		recovery: "5"
		on_land: [
			{do: "damage", params: {parts: "target_part", amount: "rule.immune ? 0 : 1"}},
			{do: "sever", params: {parts: "target_part", when: "rule.weak"}},
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
		windup:   "1"
		recovery: "3"
		on_land: [{do: "lock", params: {channels: "target_part", ticks: "8"}}]
		tags: ["attack"]
	}
	// Stand braces the body and freezes the others, then frees the stander's own channels at once.
	// Recovery is 1, not 0, because an action that takes no time is refused; the ready effect frees
	// the channels anyway. The canon: "In the deep, anyone who stands freezes everyone who crawls."
	// A body crawls when its template carries the crawls tag.
	"action.stand": {
		name: "text.action.stand"
		key:  "stand"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		windup:   "0"
		recovery: "1"
		on_land: [
			{do: "add_status", params: {bodies: "self", status: "status.standing", ticks: "10"}},
			{do: "delay", params: {channels: "others[crawls]:all", ticks: "8"}},
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
		target: "ally"
		range:  1
		sight:  true
		choices: ["piece"]
		requires: "\"graftable\" in target.template.tags"
		windup:   "6"
		recovery: "6"
		on_land: [{do: "attach_part", params: {from: "pack", item: "= chosen.piece.id"}}]
		tags: ["support", "graft"]
	}
}

text: {
	"text.action.strike": "Strike"
	"text.action.sweep":  "Sweep"
	"text.action.sever":  "Sever"
	"text.action.pin":    "Pin"
	"text.action.stand":  "Stand"
	"text.action.give":   "Give"
	"text.action.graft":  "Graft"
}
