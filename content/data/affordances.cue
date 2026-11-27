package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

affordances: [string]: content.#Affordance

// What a landmark gives: a template's affordances, a place's marked cells, or an afford effect name
// them, and agents remember what they see giving each.
affordances: {
	"affordance.water": name:    "text.affordance.water"
	"affordance.food": name:     "text.affordance.food"
	"affordance.store": name:    "text.affordance.store"
	"affordance.rest": name:     "text.affordance.rest"
	"affordance.worship": name:  "text.affordance.worship"
	"affordance.the_dead": name: "text.affordance.the_dead"
}

text: {
	"text.affordance.water":    "water"
	"text.affordance.food":     "food"
	"text.affordance.store":    "storage"
	"text.affordance.rest":     "rest"
	"text.affordance.worship":  "worship"
	"text.affordance.the_dead": "the dead"
}
