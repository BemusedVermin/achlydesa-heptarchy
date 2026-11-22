package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

templates: [string]: content.#Template

templates: {
	"template.person.reaper": {
		appearance: {glyph: "@", color: "red"}
		turn: controller: "player"
	}
	"template.person.salken": {
		appearance: {glyph: "@", color: "cyan"}
		turn: controller: "player"
	}
	"template.party.main": party: members: ["template.person.reaper", "template.person.salken"]
}
