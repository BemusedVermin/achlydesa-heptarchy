package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

recipes: [string]: content.#Recipe

// HTN tries the recipes for a goal in ID order and takes the first whose `when` holds and whose
// operators can run one after another from what the agent believes. Make `when` exclusive, so the
// order doesn't matter. Needs live in recipes/, one file each.
recipes: {
	"recipe.wander": {
		goal: "goal.wander"
		steps: [{operator: "operator.roam"}]
	}
}
