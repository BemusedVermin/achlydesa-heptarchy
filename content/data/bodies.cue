package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

parts: [string]:  content.#PartTemplate
bodies: [string]: content.#BodyPlan

// Placeholder numbers, for tuning. Lethal from the start: an ordinary Strike (3) takes off an arm or a
// head in one hit and a torso (5) in two.
parts: {
	"part.human_head": {
		name:      "text.part.head"
		slot:      "head"
		material:  "material.flesh"
		integrity: 3
		grants: [{kind: "targeting", value: 1}]
	}
	"part.human_arm": {
		name:      "text.part.arm"
		slot:      "arm"
		material:  "material.flesh"
		integrity: 3
		grants: [{kind: "action", value: 1}]
	}
	"part.human_legs": {
		name:      "text.part.legs"
		slot:      "legs"
		material:  "material.flesh"
		integrity: 3
		grants: [{kind: "speed", value: 2}]
	}
	"part.human_torso": {
		name:      "text.part.torso"
		slot:      "torso"
		material:  "material.flesh"
		integrity: 5
		vital:     true
		grants: []
	}
	// The arm of a piece of Serilmen, and the limbs of a pre-entropy remnant.
	"part.serilmen_arm": {
		name:      "text.part.serilmen_arm"
		slot:      "arm"
		material:  "material.flesh"
		integrity: 4
		grants: [{kind: "action", value: 1}]
	}
	"part.remnant_arm": {
		name:      "text.part.remnant_arm"
		slot:      "arm"
		material:  "material.old_matter"
		integrity: 4
		grants: [{kind: "action", value: 1}]
	}
	"part.remnant_core": {
		name:      "text.part.core"
		slot:      "torso"
		material:  "material.old_matter"
		integrity: 8
		vital:     true
		grants: []
	}
	"part.crawler_maw": {
		name:      "text.part.maw"
		slot:      "head"
		material:  "material.flesh"
		integrity: 5
		grants: [{kind: "targeting", value: 1}, {kind: "action", value: 1}]
	}
	"part.crawler_legs": {
		name:      "text.part.legs"
		slot:      "legs"
		material:  "material.flesh"
		integrity: 6
		grants: [{kind: "speed", value: 1}]
	}
	"part.crawler_core": {
		name:      "text.part.core"
		slot:      "torso"
		material:  "material.flesh"
		integrity: 8
		vital:     true
		grants: []
	}
}

bodies: {
	"body.human": parts: [
		{part: "part.human_head"},
		{part: "part.human_arm"},
		{part: "part.human_arm"},
		{part: "part.human_legs"},
		{part: "part.human_torso"},
	]
	"body.remnant": parts: [
		{part: "part.remnant_arm"},
		{part: "part.remnant_arm"},
		{part: "part.remnant_core"},
	]
	// A clay-bodied creature of the soop: a maw to bite with, slow legs and a core.
	"body.crawler": parts: [
		{part: "part.crawler_maw"},
		{part: "part.crawler_legs"},
		{part: "part.crawler_core", material: "material.clay"},
	]
}
