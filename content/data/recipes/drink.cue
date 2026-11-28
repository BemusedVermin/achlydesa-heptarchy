package data

recipes: {
	// Not carrying water: walk to the nearest believed water, take some and drink it.
	"recipe.drink": {
		goal: "goal.drink"
		when: "has[\"drinkable\"] == 0 && known[\"affordance.water\"] && !near[\"affordance.water\"]"
		steps: [{operator: "operator.go_to"}, {recipe: "recipe.drink_at_water"}]
	}
	"recipe.drink_at_water": {
		goal: "goal.drink"
		when: "has[\"drinkable\"] == 0 && near[\"affordance.water\"]"
		steps: [{operator: "operator.take"}, {operator: "operator.consume"}]
	}
	"recipe.drink_carried": {
		goal: "goal.drink"
		when: "has[\"drinkable\"] > 0"
		steps: [{operator: "operator.consume"}]
	}
	// Knowing no water: look for some.
	"recipe.drink_search": {
		goal: "goal.drink"
		when: "has[\"drinkable\"] == 0 && !known[\"affordance.water\"]"
		steps: [{operator: "operator.wander"}]
	}
}
