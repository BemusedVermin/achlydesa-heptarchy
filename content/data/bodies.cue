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
		// Cut off, it crawls off and fights as this.
		crawls_as: "template.crawling.remnant_arm"
	}
	"part.remnant_core": {
		name:      "text.part.core"
		slot:      "torso"
		material:  "material.old_matter"
		integrity: 8
		vital:     true
		grants: []
	}
	"part.test.crawler_maw": {
		name:      "text.part.maw"
		slot:      "head"
		material:  "material.flesh"
		integrity: 5
		grants: [{kind: "targeting", value: 1}, {kind: "action", value: 1}]
	}
	"part.test.crawler_legs": {
		name:      "text.part.legs"
		slot:      "legs"
		material:  "material.flesh"
		integrity: 6
		grants: [{kind: "speed", value: 1}]
	}
	"part.test.crawler_core": {
		name:      "text.part.core"
		slot:      "torso"
		material:  "material.flesh"
		integrity: 8
		vital:     true
		grants: []
	}
	// The parts of the moving clay: clay through and through, tougher than the test crawler's,
	// with two maws to bite with. Placeholder numbers, tuned so a party that does not know
	// clay loses (M2-11).
	"part.moving_clay_maw": {
		name:      "text.part.maw"
		slot:      "head"
		material:  "material.clay"
		integrity: 10
		grants: [{kind: "targeting", value: 1}, {kind: "action", value: 1}]
	}
	"part.moving_clay_legs": {
		name:      "text.part.legs"
		slot:      "legs"
		material:  "material.clay"
		integrity: 10
		grants: [{kind: "speed", value: 1}]
	}
	"part.moving_clay_core": {
		name:      "text.part.core"
		slot:      "torso"
		material:  "material.clay"
		integrity: 12
		vital:     true
		grants: []
	}
	// A flesh creature of the shallows: a skull to bite with, an eye that sees for it, legs and a core.
	"part.beast_skull": {
		name:      "text.part.skull"
		slot:      "head"
		material:  "material.bone"
		integrity: 4
		grants: [{kind: "action", value: 1}]
	}
	"part.beast_eye": {
		name:      "text.part.eye"
		slot:      "eye"
		material:  "material.eye"
		integrity: 2
		grants: [{kind: "targeting", value: 1}]
	}
	"part.beast_legs": {
		name:      "text.part.legs"
		slot:      "legs"
		material:  "material.flesh"
		integrity: 5
		grants: [{kind: "speed", value: 2}]
	}
	"part.beast_torso": {
		name:      "text.part.torso"
		slot:      "torso"
		material:  "material.flesh"
		integrity: 7
		vital:     true
		grants: []
	}
}

bodies: {
	"body.moving_clay": parts: [
		{part: "part.moving_clay_maw"},
		{part: "part.moving_clay_maw"},
		{part: "part.moving_clay_legs"},
		{part: "part.moving_clay_core"},
	]
	"body.beast": parts: [
		{part: "part.beast_skull"},
		{part: "part.beast_eye"},
		{part: "part.beast_legs"},
		{part: "part.beast_torso"},
	]
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
	// A test fixture: a flesh body with a clay core, for tests of materials and severing.
	"body.test.crawler": parts: [
		{part: "part.test.crawler_maw"},
		{part: "part.test.crawler_legs"},
		{part: "part.test.crawler_core", material: "material.clay"},
	]
}
