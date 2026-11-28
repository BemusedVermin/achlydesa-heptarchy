package data

recipes: {
	"recipe.hold_dead": {
		goal: "goal.hold_dead"
		when: "known[\"affordance.the_dead\"] && !near[\"affordance.the_dead\"]"
		steps: [{operator: "operator.go_to_dead"}, {recipe: "recipe.hold_dead_near"}]
	}
	"recipe.hold_dead_near": {
		goal: "goal.hold_dead"
		when: "near[\"affordance.the_dead\"]"
		steps: [{operator: "operator.consume_corpse"}]
	}
	"recipe.hold_dead_search": {
		goal: "goal.hold_dead"
		when: "!known[\"affordance.the_dead\"]"
		steps: [{operator: "operator.wander_dead"}]
	}
}
