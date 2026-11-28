package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

settlements: [string]: content.#SettlementDef

// The Reaper's crawlspace: shallow, with no water source inside. Its cistern holds what haulers fetch
// from the lake outside. Sizes and counts are first values, for tuning.
settlements: "settlement.reaper": {
	name: "text.settlement.reaper"
	at: {col: 100, row: 100, depth: 6}
	radius: 30
	street: "piece.street"
	pieces: {
		"piece.street": {weight: 8, count: 20}
		"piece.house": {weight: 4, count: 10}
		"piece.larder": {weight: 2, count: 2}
		"piece.cistern": {weight: 2, count: 2}
		"piece.shrine": {weight: 1, count: 1}
		"piece.graveyard": {weight: 1, count: 1}
		"piece.coal_pile": {weight: 1, count: 1}
	}
	exits: 2
	landmark_min: {
		"affordance.rest":     4
		"affordance.food":     1
		"affordance.water":    1
		"affordance.store":    2
		"affordance.worship":  1
		"affordance.the_dead": 1
	}
}

// The lake that waters the Reaper's crawlspace, a walk out in the deep.
settlements: "settlement.reaper_lake": {
	name: "text.settlement.reaper_lake"
	at: {col: 150, row: 100, depth: 6}
	radius: 10
	street: "piece.lake"
	pieces: "piece.lake": {weight: 1, count: 1}
	exits: 1
	landmark_min: "affordance.water": 2
}

// The salken village on the surface: fields and a well.
settlements: "settlement.salken_village": {
	name: "text.settlement.salken_village"
	at: {col: 300, row: 100, depth: 0}
	radius: 30
	street: "piece.street"
	pieces: {
		"piece.street": {weight: 8, count: 20}
		"piece.house": {weight: 4, count: 10}
		"piece.field": {weight: 3, count: 3}
		"piece.well": {weight: 2, count: 2}
		"piece.larder": {weight: 2, count: 2}
		"piece.shrine": {weight: 1, count: 1}
		"piece.graveyard": {weight: 1, count: 1}
	}
	exits: 2
	landmark_min: {
		"affordance.rest":     4
		"affordance.food":     3
		"affordance.water":    1
		"affordance.store":    1
		"affordance.worship":  1
		"affordance.the_dead": 1
	}
}

// The lake crawlspace in the middle depths, beside an underground lake.
settlements: "settlement.lake": {
	name: "text.settlement.lake"
	at: {col: 300, row: 300, depth: 18}
	radius: 30
	street: "piece.street"
	pieces: {
		"piece.street": {weight: 8, count: 20}
		"piece.house": {weight: 4, count: 10}
		"piece.lake": {weight: 2, count: 1}
		"piece.cistern": {weight: 2, count: 2}
		"piece.larder": {weight: 2, count: 2}
		"piece.shrine": {weight: 1, count: 1}
		"piece.graveyard": {weight: 1, count: 1}
	}
	exits: 2
	landmark_min: {
		"affordance.rest":     4
		"affordance.food":     1
		"affordance.water":    2
		"affordance.store":    1
		"affordance.worship":  1
		"affordance.the_dead": 1
	}
}

text: {
	"text.settlement.reaper":         "The Reaper's crawlspace"
	"text.settlement.reaper_lake":    "The Reaper's lake"
	"text.settlement.salken_village": "The salken village"
	"text.settlement.lake":           "The lake crawlspace"
}
