package data

rules: [string]: string

// Named formulas any formula may use as rule.<name>. The compiler inlines them.
rules: {
	weak:   "chosen.type.id in target.material.lists.weak_to"
	immune: "chosen.type.id in target.material.lists.immune_to"
	// What a Strike takes off, provisional. Strike never severs; it hits harder: 3 on an ordinary
	// hit, 6 on a weakness. An immunity does nothing, and is not refused, so the menu never tells a
	// party what a part is made of before it has learned it.
	hit: "rule.immune ? 0 : (rule.weak ? 6 : 3)"
	// What a Sweep takes from each part it passes over.
	graze: "rule.immune ? 0 : (rule.weak ? 2 : 1)"
}
