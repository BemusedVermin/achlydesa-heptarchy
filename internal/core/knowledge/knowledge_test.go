package knowledge_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

func newKnower(w *sim.World) ecs.Entity {
	return ecs.NewMap1[comp.Identity](w.ECS).NewEntity(&comp.Identity{UID: w.NewUID()})
}

func TestLearnTwiceKeepsOneFactWithTheNewerValue(t *testing.T) {
	w := sim.NewWorld(1)
	knower := newKnower(w)
	fact := api.Fact{
		Subject:    7,
		Claim:      "claim.material",
		Origin:     "origin.sweep",
		Value:      "material.clay",
		Confidence: 0.5,
	}
	knowledge.Learn(w, knower, fact)
	w.Now = 40
	fact.Value = "material.flesh"
	knowledge.Learn(w, knower, fact)

	got := knowledge.About(w, knower, 7)
	if len(got) != 1 || got[0].Value != "material.flesh" || got[0].Tick != 40 {
		t.Errorf("About = %+v, want one fact: flesh at tick 40", got)
	}
}

func TestAboutIsSortedAndFiltersBySubject(t *testing.T) {
	w := sim.NewWorld(1)
	knower := newKnower(w)
	for _, f := range []api.Fact{
		{Subject: 9, Claim: "claim.name", Origin: "origin.rumor"},
		{Subject: 3, Claim: "claim.name", Origin: "origin.sweep"},
		{Subject: 3, Claim: "claim.material", Origin: "origin.sweep"},
		{Subject: 3, Claim: "claim.material", Origin: "origin.rumor"},
	} {
		knowledge.Learn(w, knower, f)
	}
	got := knowledge.About(w, knower, 3)
	var keys []string
	for _, f := range got {
		keys = append(keys, string(f.Claim)+"/"+string(f.Origin))
	}
	want := []string{"claim.material/origin.rumor", "claim.material/origin.sweep", "claim.name/origin.sweep"}
	if !slices.Equal(keys, want) {
		t.Errorf("About keys = %v, want %v", keys, want)
	}
	if got := knowledge.About(w, knower, 4); len(got) != 0 {
		t.Errorf("About(unknown subject) = %v, want none", got)
	}
}

func TestBestTieBreaks(t *testing.T) {
	tests := []struct {
		name  string
		facts []api.Fact // learned in order; the tick is the index times 10
		want  api.ContentID
	}{
		{
			"confidence first",
			[]api.Fact{
				{
					Origin:     "origin.a",
					Confidence: 0.9,
				},
				{Origin: "origin.b", Confidence: 0.4},
				{Origin: "origin.c", Confidence: 0.6},
			},
			"origin.a",
		},
		{"newer tick on equal confidence", []api.Fact{
			{Origin: "origin.a", Confidence: 0.5}, {Origin: "origin.b", Confidence: 0.5},
		}, "origin.b"},
		{"source ID on equal confidence and tick", []api.Fact{
			{Origin: "origin.b", Confidence: 0.5}, {Origin: "origin.a", Confidence: 0.5},
		}, "origin.a"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := sim.NewWorld(1)
			knower := newKnower(w)
			for i, f := range test.facts {
				f.Subject, f.Claim = 1, "claim.material"
				w.Now = int64(i * 10)
				if test.name == "source ID on equal confidence and tick" {
					w.Now = 5
				}
				knowledge.Learn(w, knower, f)
			}
			best, ok := knowledge.Best(w, knower, 1, "claim.material")
			if !ok || best.Origin != test.want {
				t.Errorf("Best = %+v, %v; want source %s", best, ok, test.want)
			}
		})
	}
	w := sim.NewWorld(1)
	if _, ok := knowledge.Best(w, newKnower(w), 1, "claim.material"); ok {
		t.Error("Best with no facts reported one")
	}
}

func TestLearnAddsKnowledgeAndPublishesEvents(t *testing.T) {
	w := sim.NewWorld(1)
	var learned []api.Event
	w.Bus.SubscribeAll("test", func(_ *sim.World, e api.Event) { learned = append(learned, e) })
	member := newKnower(w)
	stranger := newKnower(w)
	party := ecs.NewMap1[comp.Controlled](w.ECS).NewEntity(&comp.Controlled{})
	ecs.NewMap[comp.MemberOf](w.ECS).Add(member, &comp.MemberOf{}, party)

	fact := api.Fact{Subject: 2, Claim: "claim.material", Value: "material.clay", Origin: "origin.sweep"}
	knowledge.Learn(w, member, fact)
	knowledge.Learn(w, stranger, fact)
	if _, err := w.Bus.Drain(w); err != nil {
		t.Fatal(err)
	}
	var noticed, learnedCount int
	for _, e := range learned {
		switch e.(type) {
		case api.FactNoticed:
			noticed++
		case api.FactLearned:
			learnedCount++
		}
	}
	if learnedCount != 2 || noticed != 1 {
		t.Errorf("events = %v, want 2 FactLearned and 1 FactNoticed (the party member's only)", learned)
	}
}

func TestNoticedFactLogsTheValueAsText(t *testing.T) {
	reg := contenttest.Sample(t)
	w := sim.NewWorld(1)
	w.Content = reg
	var noticed []api.FactNoticed
	w.Bus.SubscribeAll("test", func(_ *sim.World, e api.Event) {
		if n, ok := e.(api.FactNoticed); ok {
			noticed = append(noticed, n)
		}
	})
	knower := newKnower(w)
	ecs.NewMap[comp.Controlled](w.ECS).Add(knower, &comp.Controlled{})
	clay := api.Fact{Subject: 2, Claim: "claim.material", Value: "material.clay", Origin: "origin.sweep"}
	bram := api.Fact{Subject: 2, Claim: "claim.name", Value: "Bram", Origin: "origin.sweep"}
	knowledge.Learn(w, knower, clay)
	knowledge.Learn(w, knower, bram)
	if _, err := w.Bus.Drain(w); err != nil {
		t.Fatal(err)
	}
	if len(noticed) != 2 {
		t.Fatalf("noticed = %v, want 2", noticed)
	}
	for i, want := range []string{"You learn what it is made of: clay.", "You learn its name: Bram."} {
		if got := reg.Render(noticed[i].LogText()); got != want {
			t.Errorf("log line %d = %q, want %q", i, got, want)
		}
	}
}
