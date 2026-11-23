package content

import (
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

func TestBodiesAreLoaded(t *testing.T) {
	registry, err := Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := registry.Body("body.human")
	if !ok || len(plan.Parts) != 5 {
		t.Fatalf("body.human = %+v, %v; want five parts", plan, ok)
	}
	part, ok := registry.Part("part.human_torso")
	if !ok || !part.Vital {
		t.Errorf("part.human_torso = %+v, %v; want a vital part", part, ok)
	}
	reaper, _ := registry.Template("template.person.reaper")
	if reaper.Body == nil || *reaper.Body != "body.human" || !reaper.Protected {
		t.Errorf("reaper = %+v, want a protected human", reaper)
	}
}

func TestBodiesRejectBadReferences(t *testing.T) {
	missing := api.ContentID("body.missing")
	tests := []struct {
		name   string
		mutate func(*Bundle)
		want   string
	}{
		{"unknown grant kind", func(b *Bundle) {
			part := b.Parts["part.human_arm"]
			part.Grants = []comp.Grant{{Kind: "fly"}}
			b.Parts["part.human_arm"] = part
		}, `grants[0].kind: unknown kind "fly"`},
		{"unknown part material", func(b *Bundle) {
			part := b.Parts["part.human_arm"]
			part.Material = "material.nope"
			b.Parts["part.human_arm"] = part
		}, `unknown material "material.nope"`},
		{"unknown part in plan", func(b *Bundle) {
			b.Bodies["body.human"] = BodyPlan{Parts: []PlanPart{{Part: "part.nope"}}}
		}, `unknown part "part.nope"`},
		{"unknown override material", func(b *Bundle) {
			b.Bodies["body.human"] = BodyPlan{Parts: []PlanPart{{Part: "part.human_arm", Material: "material.nope"}}}
		}, `unknown material "material.nope"`},
		{"unknown body plan", func(b *Bundle) {
			template := b.Templates["template.person.reaper"]
			template.Body = &missing
			b.Templates["template.person.reaper"] = template
		}, `unknown body plan "body.missing"`},
		{"zero integrity", func(b *Bundle) {
			part := b.Parts["part.human_arm"]
			part.Integrity = 0
			b.Parts["part.human_arm"] = part
		}, "integrity: must be at least 1"},
		{"missing part name text", func(b *Bundle) {
			part := b.Parts["part.human_arm"]
			part.Name = "text.part.nope"
			b.Parts["part.human_arm"] = part
		}, `text reference "text.part.nope" does not exist`},
		{"empty plan", func(b *Bundle) {
			b.Bodies["body.human"] = BodyPlan{}
		}, "a body needs at least one part"},
		{"crawls as an unknown template", func(b *Bundle) {
			part := b.Parts["part.human_arm"]
			part.CrawlsAs = "template.nope"
			b.Parts["part.human_arm"] = part
		}, `parts["part.human_arm"].crawls_as: unknown template "template.nope"`},
		{"crawls as a template with a body", func(b *Bundle) {
			part := b.Parts["part.human_arm"]
			part.CrawlsAs = "template.test.dummy"
			b.Parts["part.human_arm"] = part
		}, `crawls_as: template "template.test.dummy" sets body, which a crawling part ignores`},
		{"crawls as a template with a faction", func(b *Bundle) {
			part := b.Parts["part.human_arm"]
			part.CrawlsAs = "template.crawling.remnant_arm"
			b.Parts["part.human_arm"] = part
			template := b.Templates["template.crawling.remnant_arm"]
			faction := api.ContentID("faction.deep")
			template.Faction = &faction
			b.Templates["template.crawling.remnant_arm"] = template
		}, `sets faction, which a crawling part ignores`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := LoadBundle(sampleDir)
			if err != nil {
				t.Fatal(err)
			}
			tt.mutate(&bundle)
			if _, err := New(bundle); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}
