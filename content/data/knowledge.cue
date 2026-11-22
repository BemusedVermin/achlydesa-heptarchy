package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

claims: [string]:  content.#Claim
sources: [string]: content.#Source

claims: {
	"claim.material": {name: "text.claim.material", value: "material"}
	"claim.name": name:     "text.claim.name"
}

sources: {
	"source.sweep": name: "text.source.sweep"
	"source.rumor": name: "text.source.rumor"
}

text: {
	"text.claim.material":    "material"
	"text.claim.name":        "name"
	"text.source.sweep":      "a sweep"
	"text.source.rumor":      "a rumor"
	"text.knowledge.material": "You learn what it is made of: {{.value}}."
	"text.knowledge.name":     "You learn its name: {{.value}}."
}
