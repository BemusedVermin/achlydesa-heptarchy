package combat_test

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// pinNumbers overwrites the tuned numbers of the sample content with the ones these tests are
// written against: part integrity, the forms' timing and damage, the types' timing and landing
// effects and which materials are weak or immune to what. The sample content's numbers are
// placeholders for M7's tuning and may change freely; what the tests check is how the pieces fit, so
// they carry their own numbers instead of reading the tuned ones.
func pinNumbers(b *content.Bundle) {
	for id, integrity := range map[string]int{
		"part.human_head": 3, "part.human_arm": 3, "part.human_legs": 3, "part.human_torso": 5,
		"part.crawler_maw": 5, "part.crawler_legs": 6, "part.crawler_core": 8,
	} {
		part := b.Parts[id]
		part.Integrity = integrity
		b.Parts[id] = part
	}
	cut, press, relic, name, soak := api.ContentID("type.cut"), api.ContentID("type.press"),
		api.ContentID("type.relic"), api.ContentID("type.name"), api.ContentID("type.soak")
	for id, weakTo := range map[string][]api.ContentID{
		"material.flesh": {cut, relic}, "material.clay": {soak, relic}, "material.old_matter": nil,
	} {
		material := b.Materials[id]
		material.WeakTo = weakTo
		material.ImmuneTo = nil
		if id == "material.old_matter" {
			material.ImmuneTo = []api.ContentID{relic}
		}
		b.Materials[id] = material
	}
	pinType(b, string(soak), "", "2", nil)
	pinType(b, string(press), "", "", []content.EffectRef{
		{Do: "interrupt", Params: map[string]string{"bodies": "target", "when": "winding_up"}},
	})
	for _, id := range []api.ContentID{cut, relic} {
		pinType(b, string(id), "", "", nil)
	}
	nameItem := api.ContentID("item.name")
	nameType := b.Types[string(name)]
	nameType.Consumes = &nameItem
	nameType.Condition = "target_uid == item_bound_to"
	b.Types[string(name)] = nameType

	pinForm(b, "action.strike", "3", "4", map[int]map[string]string{
		0: {"parts": "target_part", "amount": "weak ? 99 : 3"},
	})
	pinForm(b, "action.sweep", "4", "5", map[int]map[string]string{
		0: {"parts": "target:all", "amount": "weak ? 2 : 1"},
	})
	pinForm(b, "action.sever", "2", "5", map[int]map[string]string{
		0: {"parts": "target_part", "amount": "1"}, 1: {"parts": "target_part", "when": "weak"},
	})
	pinForm(b, "action.pin", "1", "3", map[int]map[string]string{0: {"channels": "target_part", "ticks": "8"}})
	pinForm(b, "action.stand", "0", "1", map[int]map[string]string{
		0: {"bodies": "self", "status": "status.standing", "ticks": "10"},
		1: {"channels": "crawling:all", "ticks": "8"},
		2: {"channels": "self:all", "count": "99"},
	})
	pinForm(b, "action.give", "0", "6", nil)
	pinForm(b, "action.graft", "6", "6", nil)
}

func pinType(b *content.Bundle, id, windup, recovery string, onLand []content.EffectRef) {
	def := b.Types[id]
	def.Windup, def.Recovery, def.OnLand = windup, recovery, onLand
	b.Types[id] = def
}

// pinForm sets a form's wind-up and recovery and the parameters of its landing effects, by index.
func pinForm(b *content.Bundle, id, windup, recovery string, params map[int]map[string]string) {
	def := b.Actions[id]
	def.Windup, def.Recovery = windup, recovery
	def.OnLand = append([]content.EffectRef(nil), def.OnLand...)
	for at, p := range params {
		def.OnLand[at] = content.EffectRef{Do: def.OnLand[at].Do, Params: p}
	}
	b.Actions[id] = def
}
