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
	// A hunting creature attacks with an action carrying this tag, and closes in with this one.
	attack_tag:  "attack"
	step_action: "action.step"
	// The claim that says what a part is made of.
	material_claim: "claim.material"
	// A body that cannot aim (no part grants targeting) swings at random.
	aims: "grant.targeting > 0"
	// How a hunting creature rates what it could do: the higher the better. What a swing takes off
	// (and so a weakness or an immunity the creature knows of), a kill about to happen, a party
	// member over anyone else, a quick wind-up, and the high ground. Placeholder numbers.
	ai_score: "(immune_known && rule.immune ? -100 : rule.hit) + (target_integrity > 0 && rule.hit >= target_integrity ? 4 : 0) + (target_is_party ? 1 : 0) - windup * 0.1 + height * 0.5"
}

text: {
	"text.combat.alert_started": "{{.name}} is in sight."
	"text.combat.alert_ended":   "The danger has passed."
	"text.combat.body_died":     "{{.name}} dies."
	"text.combat.body_downed":   "{{.name}} is out of the fight."
	"text.combat.actor_gone":    "It was out of the fight before it could act."
	"text.combat.party_wiped":   "The party is wiped out."
	"text.game.lost":            "The party is lost. Nothing more can be done."
}
