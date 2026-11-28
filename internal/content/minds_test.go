package content

import "testing"

func TestOperatorsGoalsAndRecipesAreChecked(t *testing.T) {
	data := validData + `
vocabulary: tags: items: ["edible"]
drives: "drive.water": {name: "text.drive.water", kind: "need", rise: "1", urge: "1"}
items: "item.water": name: "text.item.water"
text: {"text.drive.water": "water", "text.item.water": "water"}
operators: "operator.drink": {
	pre: [{fact: "has", tag: "drinkable", n: 1}, {fact: "near", tag: "affordance.well"}]
	eff: [{fact: "has", tag: "edible", n: -1}, {fact: "drive_met", tag: "drive.water"}, {fact: "wet", tag: "item.water"}]
	cost: 0
	does: {action: "action.drink", target: "self"}
	kind: "work"
}
goals: "goal.drink": {drive: "drive.thirst", want: [{fact: "drive_met", tag: "drive.water"}]}
recipes: "recipe.drink": {goal: "goal.drink", steps: [{operator: "operator.sip"}, {}]}
recipes: "recipe.a": {goal: "goal.drink", when: "1 +", steps: [{recipe: "recipe.b"}]}
recipes: "recipe.b": {goal: "goal.drink", steps: [{recipe: "recipe.c"}]}
recipes: "recipe.c": {goal: "goal.drink", steps: [{recipe: "recipe.a"}]}
`
	_, err := Load(writeData(t, data))
	requireErrContains(
		t,
		err,
		`operators["operator.drink"].pre[0].tag: "drinkable" is not an item tag in vocabulary.tags.items (want one of [edible])`,
		`operators["operator.drink"].pre[1].tag: unknown affordance "affordance.well"`,
		`operators["operator.drink"].eff[2].fact: unknown fact "wet"`,
		`operators["operator.drink"].cost: must be above 0`,
		`operators["operator.drink"].does.action: unknown action "action.drink"`,
		`goals["goal.drink"].drive: unknown drive "drive.thirst"`,
		`recipes["recipe.drink"].steps[0].operator: unknown operator "operator.sip"`,
		`recipes["recipe.drink"].steps[1]: name exactly one of operator and recipe`,
		`recipes["recipe.a"].when:`,
		`recipes["recipe.a"]: cycles through its sub-recipes: recipe.a -> recipe.b -> recipe.c -> recipe.a`,
		`agents.met: required when an operator or goal names drive_met`,
	)
}
