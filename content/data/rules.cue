package data

rules: [string]: string

// Named formulas any formula may use as rule.<name>. The compiler inlines them.
rules: {
	slowed: "\"status.slowed\" in source.statuses"
	// The ticks a slowed body adds to a wind-up, provisional.
	lag:  "rule.slowed ? 2 : 0"
	weak: "chosen.type.id in target.material.lists.weak_to"
	// A name or a keepsake works only on the body it is bound to. Not held for this body, it is not
	// refused: the swing lands and does nothing, as an immunity does.
	unheld: "(chosen.type.id == \"type.name\" && inventory[\"item.name\"] == 0) || (chosen.type.id == \"type.keepsake\" && inventory[\"item.keepsake\"] == 0)"
	immune: "(chosen.type.id in target.material.lists.immune_to) || rule.unheld"
	// What a Strike takes off, provisional. Strike never severs; it hits harder: 3 on an ordinary
	// hit, 6 on a weakness. An immunity does nothing, and is not refused, so the menu never tells a
	// party what a part is made of before it has learned it.
	hit: "rule.immune ? 0 : (rule.weak ? 6 : 3)"
	// What a Sweep takes from each part it passes over.
	// How an attack is rated: what a swing takes off (and so a weakness or an immunity the creature
	// knows of), a kill about to happen, a party member over anyone else, a quick wind-up, and the
	// high ground. Placeholder numbers.
	attack_score: "(material_known && rule.immune ? -100 : rule.hit) + (target.integrity > 0 && rule.hit >= target.integrity ? 4 : 0) + (target.is_party ? 1 : 0) - windup * 0.1 + height * 0.5"
	// A creature with needs rates the same, and prefers the slowly dying. Placeholder numbers.
	agents_score: "rule.attack_score + (target.dying ? 3 : 0)"
	graze:        "rule.immune ? 0 : (rule.weak ? 2 : 1)"
}
