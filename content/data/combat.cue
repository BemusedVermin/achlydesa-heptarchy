package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

combat: content.#CombatConfig

// Placeholder numbers, for tuning.
combat: {
	// An alert ends once no hostile has been in sight for this many ticks.
	calm_ticks:   30
	// A creature with no mind of its own keeps still for this long between looks.
	rest_ticks:   50
	sight_radius: 6
	sight_depth:  6
	// A dead body stays where it fell, drawn like this.
	corpse_glyph: "%"
	corpse_color: "gray"
}

text: {
	"text.combat.alert_started": "{{.name}} is in sight."
	"text.combat.alert_ended":   "The danger has passed."
	"text.combat.body_died":     "{{.name}} dies."
	"text.combat.body_downed":   "{{.name}} falls."
	"text.combat.actor_gone":    "It fell before it could act."
	"text.combat.party_lost":   "The party is lost."
	"text.game.lost":            "The party is lost. Nothing more can be done."
}
