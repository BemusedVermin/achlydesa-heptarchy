package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

templates: [string]: content.#Template

templates: {
	"template.person.reaper": appearance: {glyph: "@", color: "red"}
	"template.person.salken": appearance: {glyph: "@", color: "cyan"}
	"template.party.main": party: members: ["template.person.reaper", "template.person.salken"]
}
