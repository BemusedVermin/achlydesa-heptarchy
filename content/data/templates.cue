package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

templates: [string]: content.#Template

templates: {
	"template.person.reaper": {
		name: "text.template.person.reaper"
		appearance: {glyph: "@", color: "red"}
		actor: controller: "player"
		actions:   ["action.step"]
		body:      "body.human"
		protected: true
	}
	"template.person.salken": {
		name: "text.template.person.salken"
		appearance: {glyph: "@", color: "cyan"}
		actor: controller: "follow"
		actions:   ["action.step"]
		body:      "body.human"
		protected: true
	}
	"template.party.main": party: members: ["template.person.reaper", "template.person.salken"]
}
