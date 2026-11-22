package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// Statuses a body can be under for a while. The forms that grant them are in forms.cue.
statuses: [string]: content.#Status

statuses: {}

text: {
	"text.combat.severed":         "The part was cut away."
	"text.combat.pack_short":      "The pack has no {{.item}}."
	"text.combat.no_severed_part": "There is no severed part to fit."
}
