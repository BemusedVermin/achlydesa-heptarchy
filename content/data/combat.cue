package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

combat: content.#CombatConfig

// Placeholder numbers, for tuning.
combat: {
	// An alert ends once no hostile has been in sight for this many ticks.
	calm_ticks: 30
	// A creature with no mind of its own keeps still for this long between looks.
	still_ticks:  50
	sight_radius: 6
	sight_depth:  6
	// A dead body stays where it fell, drawn like this.
	corpse_glyph: "%"
	corpse_color: "gray"
	// A hunting creature attacks with an action carrying this tag, and closes in with this one.
	attack_tag:  "attack"
	step_action: "action.step"
	// An action carrying this tag may be aimed at a dead body.
	corpse_tag: "corpse"
	// The claim that says what a part is made of.
	material_claim: "claim.material"
	// A body that cannot aim (no part grants targeting) swings at random.
	aims: "grant.targeting > 0"
	// How a hunting creature rates what it could do: the higher the better. What a swing takes off
	// (and so a weakness or an immunity the creature knows of), a kill about to happen, a party
	// member over anyone else, a quick wind-up, and the high ground. Placeholder numbers.
	ai_score: "(material_known && rule.immune ? -100 : rule.hit) + (target.integrity > 0 && rule.hit >= target.integrity ? 4 : 0) + (target.is_party ? 1 : 0) - windup * 0.1 + height * 0.5"
}
