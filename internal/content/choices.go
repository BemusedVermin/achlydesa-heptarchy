package content

import (
	"fmt"

	"github.com/expr-lang/expr"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Choice returns the content choice with the given name.
func (r *Registry) Choice(name string) (ChoiceDef, bool) {
	def, ok := r.bundle.Choices[name]
	return def, ok
}

// ChoiceNames lists every content choice name in order.
func (r *Registry) ChoiceNames() []string { return sortedKeys(r.bundle.Choices) }

// Offered says whether the record is offered for the choice under env. A choice with no offer
// formula offers every record.
func (r *Registry) Offered(choice string, env OfferEnv) (bool, error) {
	program, ok := r.offers[choice]
	if !ok {
		return true, nil
	}
	return run[bool](program, env, fmt.Sprintf("choice %q offer", choice))
}

// checkChoices checks every content choice and compiles its offer.
func (r *Registry) checkChoices() []error {
	p := r.problems()
	for _, name := range sortedKeys(r.bundle.Choices) {
		def := r.bundle.Choices[name]
		path := fmt.Sprintf("choices[%q]", name)
		if !keyPattern.MatchString(name) {
			p.addf("%s: a choice name must be lower-case words", path)
		}
		if _, ok := r.recordCollection(def.From); !ok {
			p.addf("%s.from: %q is not a collection of records", path, def.From)
		}
		if def.Offer == "" {
			continue
		}
		program, err := r.Compile(def.Offer, OfferEnv{}, expr.AsBool())
		if err != nil {
			p.addf("%s.offer: %w", path, err)
			continue
		}
		r.offers[name] = program
	}
	return p.errs
}

// ChoiceRecords lists the IDs of the records a choice draws from, in ID order.
func (r *Registry) ChoiceRecords(name string) []api.ContentID {
	def, ok := r.bundle.Choices[name]
	if !ok {
		return nil
	}
	return r.RecordIDs(def.From)
}
