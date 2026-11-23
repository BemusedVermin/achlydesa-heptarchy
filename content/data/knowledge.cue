package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

claims: [string]:  content.#Claim
origins: [string]: content.#Origin

claims: {
	"claim.material": {name: "text.claim.material", value: "material"}
	"claim.name": name:     "text.claim.name"
}

origins: {
	"origin.sweep": name: "text.origin.sweep"
	"origin.rumor": name: "text.origin.rumor"
}

text: {
	"text.claim.material":    "material"
	"text.claim.name":        "name"
	"text.origin.sweep":      "a sweep"
	"text.origin.rumor":      "a rumor"
	"text.knowledge.material": "You learn what it is made of: {{.value}}."
	"text.knowledge.name":     "You learn its name: {{.value}}."
}
