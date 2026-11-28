package data

recipes: {
	"recipe.eat": {
		goal: "goal.eat"
		when: "has[\"edible\"] == 0 && known[\"affordance.food\"] && !near[\"affordance.food\"]"
		steps: [{operator: "operator.go_to_food"}, {recipe: "recipe.eat_at_food"}]
	}
	"recipe.eat_at_food": {
		goal: "goal.eat"
		when: "has[\"edible\"] == 0 && near[\"affordance.food\"]"
		steps: [{operator: "operator.take_food"}, {operator: "operator.eat"}]
	}
	"recipe.eat_carried": {
		goal: "goal.eat"
		when: "has[\"edible\"] > 0"
		steps: [{operator: "operator.eat"}]
	}
	"recipe.eat_search": {
		goal: "goal.eat"
		when: "has[\"edible\"] == 0 && !known[\"affordance.food\"]"
		steps: [{operator: "operator.wander_food"}]
	}
}
