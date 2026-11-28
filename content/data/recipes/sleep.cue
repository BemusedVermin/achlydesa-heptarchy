package data

recipes: {
	"recipe.sleep": {
		goal: "goal.sleep"
		when: "known[\"affordance.rest\"] && !near[\"affordance.rest\"]"
		steps: [{operator: "operator.go_to_rest"}, {recipe: "recipe.sleep_at_rest"}]
	}
	"recipe.sleep_at_rest": {
		goal: "goal.sleep"
		when: "near[\"affordance.rest\"]"
		steps: [{operator: "operator.sleep"}]
	}
	"recipe.sleep_search": {
		goal: "goal.sleep"
		when: "!known[\"affordance.rest\"]"
		steps: [{operator: "operator.wander_rest"}]
	}
}
