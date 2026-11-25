package agents_test

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestExoticRecruitsAreHirableInTheSampleWorldAndNotGraftable(t *testing.T) {
	reg := contenttest.Sample(t)
	lake, ok := reg.Map(sampleLake)
	if !ok || lake.Region == nil || lake.Region.Town == nil {
		t.Fatal("the sample lake is not a town")
	}
	for _, id := range []api.ContentID{"template.generic.tamed_clay", "template.generic.leashed_feeder"} {
		template, ok := reg.Template(id)
		if !ok || template.Recruit == nil {
			t.Fatalf("%s is not hirable", id)
		}
		if slices.Contains(template.Tags, "graftable") {
			t.Errorf("%s can be grafted", id)
		}
		if !slices.ContainsFunc(lake.Region.Town.Generics, func(g content.GenericDef) bool { return g.Template == id }) {
			t.Errorf("the lake town does not keep %s", id)
		}
	}
}
