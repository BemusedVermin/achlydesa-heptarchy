package data

rules: [string]: string

// Named formulas any formula may use as rule.<name>. The compiler inlines them.
rules: {
	weak:   "chosen.type.id in target.material.lists.weak_to"
	immune: "chosen.type.id in target.material.lists.immune_to"
}
