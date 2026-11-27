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
		facts []api.Fact
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

func cappedWorld(tb testing.TB, capacity int) *sim.World {
	tb.Helper()
	bundle := contenttest.SampleBundle(tb)
	bundle.Knowledge.Cap = capacity
	w := sim.NewWorld(1)
	w.Content = contenttest.Registry(tb, bundle)
	return w
}

func TestLearnPastTheCapForgetsTheLeastSureThenTheOldest(t *testing.T) {
	w := cappedWorld(t, 3)
	knower := newKnower(w)
	var learned []api.UID
	w.Bus.SubscribeAll("test", func(_ *sim.World, e api.Event) {
		if l, ok := e.(api.FactLearned); ok {
			learned = append(learned, l.Fact.Subject)
		}
	})
	learn := func(subject api.UID, confidence float64) {
		knowledge.Learn(
			w,
			knower,
			api.Fact{Subject: subject, Claim: "claim.at", Origin: "origin.seen", Confidence: confidence},
		)
		w.Now++
	}
	learn(1, 0.5)
	learn(2, 0.9)
	learn(3, 0.5)
	learn(4, 0.7)
	learn(5, 0.1)
	if _, err := w.Bus.Drain(w); err != nil {
		t.Fatal(err)
	}
	var kept []api.UID
	for _, f := range ecs.NewMap[comp.Knowledge](w.ECS).Get(knower).Facts {
		kept = append(kept, f.Subject)
	}
	if want := []api.UID{2, 3, 4}; !slices.Equal(kept, want) {
		t.Errorf(
			"kept subjects %v, want %v: the oldest of the least sure goes first, and a new fact surer of nothing goes itself",
			kept,
			want,
		)
	}
	if want := []api.UID{1, 2, 3, 4}; !slices.Equal(learned, want) {
		t.Errorf("FactLearned for subjects %v, want %v: none for a fact forgotten as it came", learned, want)
	}
	if !knowledge.Forget(w, knower, 3, "claim.at", "origin.seen") || len(knowledge.About(w, knower, 3)) != 0 {
		t.Error("Forget did not drop the fact")
	}
	if knowledge.Forget(w, knower, 3, "claim.at", "origin.seen") {
		t.Error("Forget reported a fact it no longer held")
	}
}

func TestThePartysFactsAreNotCapped(t *testing.T) {
	w := cappedWorld(t, 2)
	member := newKnower(w)
	party := ecs.NewMap1[comp.Controlled](w.ECS).NewEntity(&comp.Controlled{})
	ecs.NewMap[comp.MemberOf](w.ECS).Add(member, &comp.MemberOf{}, party)
	for subject := range api.UID(5) {
		knowledge.Learn(w, member, api.Fact{Subject: subject, Claim: "claim.at", Origin: "origin.seen"})
	}
	if got := len(ecs.NewMap[comp.Knowledge](w.ECS).Get(member).Facts); got != 5 {
		t.Errorf("a party member keeps %d facts, want all 5", got)
	}
}

func TestLearnPublishesOnlyWhatTheKnowerDidNotHold(t *testing.T) {
	w := sim.NewWorld(1)
	var learned int
	w.Bus.SubscribeAll("test", func(_ *sim.World, e api.Event) {
		if _, ok := e.(api.FactLearned); ok {
			learned++
		}
	})
	knower := newKnower(w)
	fact := api.Fact{Subject: 2, Claim: "claim.at", Value: "1,0,0", Origin: "origin.seen", Confidence: 0.8}
	moved, surer, otherSource := fact, fact, fact
	moved.Value = "2,0,0"
	surer.Value, surer.Confidence = moved.Value, 1
	otherSource.Origin = "origin.rumor"
	for _, test := range []struct {
		name string
		fact api.Fact
		want int
	}{
		{"a new fact", fact, 1},
		{"the same fact again", fact, 0},
		{"a new value", moved, 1},
		{"a new confidence", surer, 1},
		{"the same claim from another source", otherSource, 1},
		{"the new value again", surer, 0},
	} {
		learned = 0
		w.Now++
		knowledge.Learn(w, knower, test.fact)
		if _, err := w.Bus.Drain(w); err != nil {
			t.Fatal(err)
		}
		if learned != test.want {
			t.Errorf("%s: %d FactLearned, want %d", test.name, learned, test.want)
		}
	}
	if got := knowledge.About(w, knower, 2); got[1].Tick != w.Now {
		t.Errorf("a known fact learned again kept tick %d, want %d", got[1].Tick, w.Now)
	}
}

func TestThePartyNoticesAFactItAlreadyHeld(t *testing.T) {
	w := sim.NewWorld(1)
	var noticed int
	w.Bus.SubscribeAll("test", func(_ *sim.World, e api.Event) {
		if _, ok := e.(api.FactNoticed); ok {
			noticed++
		}
	})
	knower := newKnower(w)
	ecs.NewMap[comp.Controlled](w.ECS).Add(knower, &comp.Controlled{})
	fact := api.Fact{Subject: 2, Claim: "claim.material", Value: "material.clay", Origin: "origin.sweep"}
	knowledge.Learn(w, knower, fact)
	knowledge.Learn(w, knower, fact)
	if _, err := w.Bus.Drain(w); err != nil {
		t.Fatal(err)
	}
	if noticed != 2 {
		t.Errorf("FactNoticed %d times, want 2", noticed)
	}
}
