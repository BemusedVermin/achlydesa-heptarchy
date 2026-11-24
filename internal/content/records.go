package content

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

type collection struct {
	name        Collection
	prefix      string
	keys        func(*Bundle) []string
	data        func(*Bundle) map[string]Data
	recordName  func(b *Bundle, key string) (api.ContentID, bool)
	sharedNames bool
}

func collections() []collection {
	return []collection{
		described(CollectionMaterials, "material", func(b *Bundle) map[string]Material { return b.Materials },
			func(m Material) Data { return m.Data }, func(m Material) (api.ContentID, bool) { return m.Name, true }),
		described(CollectionTerrains, "terrain", func(b *Bundle) map[string]Terrain { return b.Terrains },
			nil, func(t Terrain) (api.ContentID, bool) { return t.Name, true }),
		described[string](CollectionText, "text", func(b *Bundle) map[string]string { return b.Text }, nil, nil),
		described(CollectionTemplates, "template", func(b *Bundle) map[string]Template { return b.Templates },
			func(t Template) Data { return t.Data }, func(t Template) (api.ContentID, bool) {
				if t.Name == nil {
					return "", false
				}
				return *t.Name, true
			}),
		described(CollectionActions, "action", func(b *Bundle) map[string]ActionDef { return b.Actions },
			nil, func(a ActionDef) (api.ContentID, bool) { return a.Name, true }),
		described(CollectionMaps, "map", func(b *Bundle) map[string]MapDef { return b.Maps },
			nil, func(m MapDef) (api.ContentID, bool) { return m.Name, true }),
		shared(described(
			CollectionParts,
			"part",
			func(b *Bundle) map[string]PartTemplate { return b.Parts },
			func(p PartTemplate) Data { return p.Data },
			func(p PartTemplate) (api.ContentID, bool) { return p.Name, true },
		)),
		described(CollectionClaims, "claim", func(b *Bundle) map[string]Claim { return b.Claims },
			nil, func(c Claim) (api.ContentID, bool) { return c.Name, true }),
		described(CollectionOrigins, "origin", func(b *Bundle) map[string]Origin { return b.Origins },
			nil, func(o Origin) (api.ContentID, bool) { return o.Name, true }),
		described[BodyPlan](CollectionBodies, "body", func(b *Bundle) map[string]BodyPlan { return b.Bodies }, nil, nil),
		described(CollectionTypes, "type", func(b *Bundle) map[string]Type { return b.Types },
			func(t Type) Data { return t.Data }, func(t Type) (api.ContentID, bool) { return t.Name, true }),
		described(CollectionItems, "item", func(b *Bundle) map[string]Item { return b.Items },
			func(i Item) Data { return i.Data }, func(i Item) (api.ContentID, bool) { return i.Name, true }),
		described(CollectionFactions, "faction", func(b *Bundle) map[string]Faction { return b.Factions },
			func(f Faction) Data { return f.Data }, func(f Faction) (api.ContentID, bool) { return f.Name, true }),
		described(CollectionStatuses, "status", func(b *Bundle) map[string]Status { return b.Statuses },
			func(s Status) Data { return s.Data }, func(s Status) (api.ContentID, bool) { return s.Name, true }),
		described(CollectionNeeds, "need", func(b *Bundle) map[string]NeedDef { return b.Needs },
			nil, func(n NeedDef) (api.ContentID, bool) { return n.Name, true }),
	}
}

func described[V any](
	name Collection,
	prefix string,
	records func(*Bundle) map[string]V,
	data func(V) Data,
	recordName func(V) (api.ContentID, bool),
) collection {
	c := collection{name: name, prefix: prefix, keys: func(b *Bundle) []string { return sortedKeys(records(b)) }}
	if data != nil {
		c.data = func(b *Bundle) map[string]Data {
			out := map[string]Data{}
			for id, rec := range records(b) {
				out[id] = data(rec)
			}
			return out
		}
	}
	if recordName != nil {
		c.recordName = func(b *Bundle, key string) (api.ContentID, bool) {
			rec, ok := records(b)[key]
			if !ok {
				return "", false
			}
			return recordName(rec)
		}
	}
	return c
}

func shared(c collection) collection {
	c.sharedNames = true
	return c
}

func (r *Registry) recordCollection(name Collection) (collection, bool) {
	for _, c := range r.collections {
		if c.name == name && c.data != nil {
			return c, true
		}
	}
	return collection{}, false
}

func (r *Registry) dataOf(name Collection) map[string]Data {
	c, ok := r.recordCollection(name)
	if !ok {
		return map[string]Data{}
	}
	return c.data(&r.bundle)
}

func (r *Registry) collectionIDs(name Collection) (map[string]bool, bool) {
	for _, c := range r.collections {
		if c.name != name {
			continue
		}
		keys := c.keys(&r.bundle)
		out := make(map[string]bool, len(keys))
		for _, key := range keys {
			out[key] = true
		}
		return out, true
	}
	return nil, false
}

func (r *Registry) checkIDs() []error {
	p := r.problems()
	for _, c := range r.collections {
		for _, key := range c.keys(&r.bundle) {
			if err := checkID(string(c.name), key, c.prefix); err != nil {
				p.add(err)
			}
		}
	}
	return p.errs
}

func (r *Registry) checkTextNames() []error {
	p := r.problems()
	check := func(collection, key string, name api.ContentID) {
		if want := api.ContentID("text." + key); name != want {
			p.addf("%s[%q].name: text ID is %q, want %q", collection, key, name, want)
		}
	}
	for _, c := range r.collections {
		if c.recordName == nil || c.sharedNames {
			continue
		}
		for _, key := range c.keys(&r.bundle) {
			if name, ok := c.recordName(&r.bundle, key); ok {
				check(string(c.name), key, name)
			}
		}
	}
	for _, phase := range r.bundle.Calendar.Phases {
		check("calendar.phases", string(phase.ID), phase.Name)
	}
	return p.errs
}

func recordOf(id string, data Data) Record {
	record := Record{ID: id, Tags: slices.Clone(data.Tags), Props: maps.Clone(data.Props), Lists: map[string][]string{}}
	if record.Tags == nil {
		record.Tags = []string{}
	}
	if record.Props == nil {
		record.Props = map[string]float64{}
	}
	for key, value := range data.Lists {
		record.Lists[key] = slices.Clone(value)
	}
	return record
}

func (r *Registry) buildRecords() {
	r.records = map[api.ContentID]Record{}
	for _, c := range r.collections {
		if c.data == nil {
			continue
		}
		for id, data := range c.data(&r.bundle) {
			r.records[api.ContentID(id)] = recordOf(id, data)
		}
	}
}

// Record finds a record by its ID, which says which collection it is in. For an ID it does not
// know it returns Record{ID: id} and false. The record's maps and slices are shared: don't change them.
func (r *Registry) Record(id api.ContentID) (Record, bool) {
	record, ok := r.records[id]
	if !ok {
		return emptyRecord(string(id)), false
	}
	return record, true
}

// EmptyRecord is the record of nothing: no ID, tags, props or lists.
func EmptyRecord() Record { return emptyRecord("") }

func emptyRecord(id string) Record {
	return Record{ID: id, Tags: []string{}, Props: map[string]float64{}, Lists: map[string][]string{}}
}

// RecordIDs lists the IDs of a record collection in order. An unknown collection has none.
func (r *Registry) RecordIDs(collection Collection) []api.ContentID {
	return sortedIDs(r.dataOf(collection))
}

func (r *Registry) checkVocabulary() []error {
	p := r.problems()
	vocab := r.bundle.Vocabulary
	for _, collection := range sortedKeys(vocab.Tags) {
		if _, ok := r.recordCollection(Collection(collection)); !ok {
			p.addf("vocabulary.tags[%q]: not a collection of records", collection)
			continue
		}
		data := r.dataOf(Collection(collection))
		for _, id := range sortedKeys(data) {
			for i, tag := range data[id].Tags {
				if !slices.Contains(vocab.Tags[collection], tag) {
					p.addf("%s[%q].tags[%d]: tag %q is not in vocabulary.tags (want one of %v)",
						collection, id, i, tag, vocab.Tags[collection])
				}
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Parts) {
		for i, grant := range r.bundle.Parts[key].Grants {
			path := fmt.Sprintf("parts[%q].grants[%d]", key, i)
			if grant.Value == 0 {
				p.addf("%s.value: a grant needs a value other than 0", path)
			}
			if len(vocab.Grants) > 0 && !slices.Contains(vocab.Grants, grant.Kind) {
				p.addf("%s.kind: unknown kind %q (want one of %v)", path, grant.Kind, vocab.Grants)
			}
		}
	}
	for _, key := range sortedKeys(vocab.Lists) {
		collection, list, ok := strings.Cut(key, ".")
		if _, record := r.recordCollection(Collection(collection)); !ok || !record {
			p.addf("vocabulary.lists[%q]: want <collection of records>.<key>", key)
			continue
		}
		targets, ok := r.collectionIDs(vocab.Lists[key])
		if !ok {
			p.addf("vocabulary.lists[%q]: unknown collection %q", key, vocab.Lists[key])
			continue
		}
		data := r.dataOf(Collection(collection))
		for _, id := range sortedKeys(data) {
			for i, value := range data[id].Lists[list] {
				if !targets[value] {
					p.addf("%s[%q].lists[%q][%d]: %q is not in %s", collection, id, list, i, value, vocab.Lists[key])
				}
			}
		}
	}
	return p.errs
}

// RecordName returns the text ID of a record's name, or false when the ID is not a record or the
// record has no name.
func (r *Registry) RecordName(id api.ContentID) (api.ContentID, bool) {
	for _, c := range r.collections {
		if c.data == nil || c.recordName == nil {
			continue
		}
		if name, ok := c.recordName(&r.bundle, string(id)); ok {
			return name, name != ""
		}
	}
	return "", false
}
