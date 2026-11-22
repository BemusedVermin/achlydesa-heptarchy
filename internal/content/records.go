package content

import (
	"fmt"
	"slices"
	"strings"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// recordCollections maps every collection whose records carry [Data] to its ID prefix. Collection
// names are the Bundle's json keys.
var recordCollections = map[string]string{
	"materials": "material", "types": "type", "items": "item", "parts": "part",
	"statuses": "status", "factions": "faction", "templates": "template",
}

// dataOf lists the Data of every record of a collection that has one, by ID.
func (r *Registry) dataOf(collection string) map[string]Data {
	out := map[string]Data{}
	switch collection {
	case "materials":
		for id, rec := range r.bundle.Materials {
			out[id] = rec.Data
		}
	case "types":
		for id, rec := range r.bundle.Types {
			out[id] = rec.Data
		}
	case "items":
		for id, rec := range r.bundle.Items {
			out[id] = rec.Data
		}
	case "parts":
		for id, rec := range r.bundle.Parts {
			out[id] = rec.Data
		}
	case "statuses":
		for id, rec := range r.bundle.Statuses {
			out[id] = rec.Data
		}
	case "factions":
		for id, rec := range r.bundle.Factions {
			out[id] = rec.Data
		}
	case "templates":
		for id, rec := range r.bundle.Templates {
			out[id] = rec.Data
		}
	}
	return out
}

// collectionIDs lists the IDs of any collection of the bundle, or false when there is none by that name.
func (r *Registry) collectionIDs(collection string) (map[string]bool, bool) {
	var keys []string
	switch collection {
	case "materials":
		keys = sortedKeys(r.bundle.Materials)
	case "terrains":
		keys = sortedKeys(r.bundle.Terrains)
	case "text":
		keys = sortedKeys(r.bundle.Text)
	case "templates":
		keys = sortedKeys(r.bundle.Templates)
	case "actions":
		keys = sortedKeys(r.bundle.Actions)
	case "maps":
		keys = sortedKeys(r.bundle.Maps)
	case "claims":
		keys = sortedKeys(r.bundle.Claims)
	case "sources":
		keys = sortedKeys(r.bundle.Sources)
	case "parts":
		keys = sortedKeys(r.bundle.Parts)
	case "bodies":
		keys = sortedKeys(r.bundle.Bodies)
	case "types":
		keys = sortedKeys(r.bundle.Types)
	case "items":
		keys = sortedKeys(r.bundle.Items)
	case "factions":
		keys = sortedKeys(r.bundle.Factions)
	case "statuses":
		keys = sortedKeys(r.bundle.Statuses)
	default:
		return nil, false
	}
	out := make(map[string]bool, len(keys))
	for _, key := range keys {
		out[key] = true
	}
	return out, true
}

func recordOf(id string, data Data) Record {
	record := Record{
		ID: id, Tags: slices.Clone(data.Tags), Props: map[string]float64{}, Lists: map[string][]string{},
	}
	if record.Tags == nil {
		record.Tags = []string{}
	}
	for key, value := range data.Props {
		record.Props[key] = value
	}
	for key, value := range data.Lists {
		record.Lists[key] = slices.Clone(value)
	}
	return record
}

// buildRecords makes the formula view of every record, once per registry.
func (r *Registry) buildRecords() {
	r.records = map[api.ContentID]Record{}
	for collection := range recordCollections {
		for id, data := range r.dataOf(collection) {
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
func (r *Registry) RecordIDs(collection string) []api.ContentID {
	return sortedIDs(r.dataOf(collection))
}

// checkVocabulary checks tags, grant kinds and list values against what content declared closed.
func (r *Registry) checkVocabulary() []error {
	var problems []error
	vocab := r.bundle.Vocabulary
	for _, collection := range sortedKeys(vocab.Tags) {
		if _, ok := recordCollections[collection]; !ok {
			problems = append(problems, fmt.Errorf("vocabulary.tags[%q]: not a collection of records", collection))
			continue
		}
		data := r.dataOf(collection)
		for _, id := range sortedKeys(data) {
			for i, tag := range data[id].Tags {
				if !slices.Contains(vocab.Tags[collection], tag) {
					problems = append(problems, fmt.Errorf(
						"%s[%q].tags[%d]: tag %q is not in vocabulary.tags (want one of %v)",
						collection, id, i, tag, vocab.Tags[collection]))
				}
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Parts) {
		for i, grant := range r.bundle.Parts[key].Grants {
			path := fmt.Sprintf("parts[%q].grants[%d]", key, i)
			if grant.Value == 0 {
				problems = append(problems, fmt.Errorf("%s.value: a grant needs a value other than 0", path))
			}
			if len(vocab.Grants) > 0 && !slices.Contains(vocab.Grants, grant.Kind) {
				problems = append(problems, fmt.Errorf(
					"%s.kind: unknown kind %q (want one of %v)", path, grant.Kind, vocab.Grants))
			}
		}
	}
	for _, key := range sortedKeys(vocab.Lists) {
		collection, list, ok := strings.Cut(key, ".")
		if _, record := recordCollections[collection]; !ok || !record {
			problems = append(problems, fmt.Errorf("vocabulary.lists[%q]: want <collection of records>.<key>", key))
			continue
		}
		targets, ok := r.collectionIDs(vocab.Lists[key])
		if !ok {
			problems = append(
				problems,
				fmt.Errorf("vocabulary.lists[%q]: unknown collection %q", key, vocab.Lists[key]),
			)
			continue
		}
		data := r.dataOf(collection)
		for _, id := range sortedKeys(data) {
			for i, value := range data[id].Lists[list] {
				if !targets[value] {
					problems = append(problems, fmt.Errorf(
						"%s[%q].lists[%q][%d]: %q is not in %s", collection, id, list, i, value, vocab.Lists[key]))
				}
			}
		}
	}
	return problems
}
