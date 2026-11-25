package sim

import (
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

func newSampleWorld(t *testing.T, seed uint64) *World {
	t.Helper()
	w := NewWorld(seed)
	w.Content = contenttest.Sample(t)
	return w
}

func TestSpawnParty(t *testing.T) {
	w := newSampleWorld(t, 1)
	party, err := w.Spawn("template.party.main", nil)
	if err != nil {
		t.Fatal(err)
	}
	id := ecs.NewMap[comp.Identity](w.ECS)
	pos := ecs.NewMap[comp.Pos](w.ECS)
	appearance := ecs.NewMap[comp.Appearance](w.ECS)
	partyMap := ecs.NewMap[comp.Party](w.ECS)
	memberOf := ecs.NewMap[comp.MemberOf](w.ECS)

	if !id.Has(party) || pos.Has(party) {
		t.Fatal("party needs Identity and no Pos")
	}
	members := ecs.NewFilter1[comp.MemberOf](w.ECS).Relations(ecs.Rel[comp.MemberOf](party))
	var got []api.ContentID
	var colors []string
	var uids []api.UID
	query := members.Query()
	for query.Next() {
		e := query.Entity()
		got = append(got, id.Get(e).Template)
		uids = append(uids, id.Get(e).UID)
		if pos.Has(e) {
			t.Error("member has a position")
		}
		if memberOf.GetRelation(e) != party {
			t.Error("member not MemberOf the party")
		}
		look := appearance.Get(e)
		if look.Glyph != "@" {
			t.Errorf("glyph = %q", look.Glyph)
		}
		colors = append(colors, look.Color)
	}
	if len(got) != 2 || got[0] != "template.person.reaper" || got[1] != "template.person.salken" {
		t.Fatalf("members = %v", got)
	}
	if colors[0] == colors[1] {
		t.Errorf("members share a color: %v", colors)
	}
	if uids[0] > uids[1] {
		t.Errorf("members out of order: %v", uids)
	}
	if p := *partyMap.Get(party); p.Slots != 4 || p.Leader != uids[0] {
		t.Errorf("Party = %+v, want slots 4 leader %d", p, uids[0])
	}
}

func TestDespawnRemovesAttachedParts(t *testing.T) {
	w := newSampleWorld(t, 1)
	body, err := w.Spawn("template.person.reaper", nil)
	if err != nil {
		t.Fatal(err)
	}
	other, err := w.Spawn("template.person.reaper", nil)
	if err != nil {
		t.Fatal(err)
	}
	partOf := ecs.NewMap[comp.PartOf](w.ECS)
	attached := w.NewIdentified("part.test")
	partOf.Add(attached, &comp.PartOf{}, body)
	nested := w.NewIdentified("part.test")
	partOf.Add(nested, &comp.PartOf{}, attached)
	severed := w.NewIdentified("part.test")
	kept := w.NewIdentified("part.test")
	partOf.Add(kept, &comp.PartOf{}, other)
	attachedUID := w.UIDOf(attached)

	w.Despawn(body)

	if w.ECS.Alive(attached) || w.ECS.Alive(nested) {
		t.Error("a part survived its body")
	}
	if _, ok := w.ByUID(attachedUID); ok {
		t.Error("attached part still in the UID index")
	}
	if !w.ECS.Alive(severed) || !w.ECS.Alive(kept) {
		t.Error("an unrelated part was removed")
	}
	query := ecs.NewFilter1[comp.PartOf](w.ECS).Query()
	for query.Next() {
		if partOf.GetRelation(query.Entity()).IsZero() {
			t.Error("a part with a zero PartOf target remains")
		}
	}
}

func TestByUIDAndDespawn(t *testing.T) {
	w := newSampleWorld(t, 1)
	e, err := w.Spawn("template.person.reaper", nil)
	if err != nil {
		t.Fatal(err)
	}
	uid := w.UIDOf(e)
	if found, ok := w.ByUID(uid); !ok || found != e {
		t.Fatalf("ByUID(%d) = %v, %v", uid, found, ok)
	}
	w.Despawn(e)
	if _, ok := w.ByUID(uid); ok {
		t.Error("ByUID finds a despawned entity")
	}
	w.Despawn(e)
	next, err := w.Spawn("template.person.reaper", nil)
	if err != nil {
		t.Fatal(err)
	}
	if w.UIDOf(next) == uid {
		t.Error("UID was reused")
	}
}

func TestSpawnPositionsEntity(t *testing.T) {
	w := newSampleWorld(t, 1)
	at := comp.Pos{At: api.Cell{}}
	e, err := w.Spawn("template.person.salken", &at)
	if err != nil {
		t.Fatal(err)
	}
	if got := ecs.NewMap[comp.Pos](w.ECS).Get(e); *got != at {
		t.Errorf("Pos = %+v, want %+v", *got, at)
	}
}

func TestSpawnersRunInOrder(t *testing.T) {
	w := newSampleWorld(t, 1)
	var calls []string
	for _, name := range []string{"first", "second"} {
		err := w.RegisterSpawner(Spawner{Name: name, Apply: func(*World, ecs.Entity, content.Template) error {
			calls = append(calls, name)
			return nil
		}})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := w.Spawn("template.person.reaper", nil); err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "first,second" {
		t.Errorf("calls = %v", calls)
	}
	if err := w.RegisterSpawner(
		Spawner{Name: "first", Apply: func(*World, ecs.Entity, content.Template) error { return nil }},
	); err == nil {
		t.Error("duplicate spawner name accepted")
	}
}

func TestSpawnFailureRemovesEntity(t *testing.T) {
	w := newSampleWorld(t, 1)
	if err := w.RegisterSpawner(Spawner{Name: "boom", Apply: func(*World, ecs.Entity, content.Template) error {
		return errBoom
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Spawn("template.party.main", nil); err == nil {
		t.Fatal("want an error")
	}
	query := ecs.NewFilter1[comp.Identity](w.ECS).Query()
	if n := query.Count(); n != 0 {
		t.Errorf("%d entities left behind", n)
	}
}

type testError string

func (e testError) Error() string { return string(e) }

const errBoom = testError("boom")

func TestSpawnErrors(t *testing.T) {
	w := newSampleWorld(t, 1)
	if _, err := w.Spawn("template.nope", nil); err == nil {
		t.Error("unknown template accepted")
	}
	if _, err := NewWorld(1).Spawn("template.person.reaper", nil); err == nil {
		t.Error("spawn without content accepted")
	}
}

func TestSpawnDeterministicUIDs(t *testing.T) {
	var runs [2][]api.UID
	for i := range runs {
		w := newSampleWorld(t, 42)
		for _, template := range []api.ContentID{"template.party.main", "template.person.salken", "template.party.main"} {
			e, err := w.Spawn(template, nil)
			if err != nil {
				t.Fatal(err)
			}
			runs[i] = append(runs[i], w.UIDOf(e))
		}
	}
	if len(runs[0]) != 3 || runs[0][0] != runs[1][0] || runs[0][1] != runs[1][1] || runs[0][2] != runs[1][2] {
		t.Errorf("UIDs differ: %v vs %v", runs[0], runs[1])
	}
}
