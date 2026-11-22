package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

templates: [string]: content.#Template

templates: {
	"template.person.reaper": {
		appearance: {glyph: "@", color: "red"}
		actor: controller: "player"
		actions: ["action.step"]
	}
	"template.person.salken": {
		appearance: {glyph: "@", color: "cyan"}
		actor: controller: "player"
		actions: ["action.step"]
	}
	"template.party.main": party: members: ["template.person.reaper", "template.person.salken"]
}
