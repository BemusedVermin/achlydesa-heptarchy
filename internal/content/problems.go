package content

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

type problems struct {
	bundle *Bundle
	errs   []error
}

func (r *Registry) problems() *problems { return &problems{bundle: &r.bundle} }

func (p *problems) addf(format string, args ...any) {
	p.errs = append(p.errs, fmt.Errorf(format, args...))
}

func (p *problems) add(errs ...error) { p.errs = append(p.errs, errs...) }

func (p *problems) text(path string, ref api.ContentID) {
	if _, ok := p.bundle.Text[string(ref)]; !ok {
		p.addf("%s: text reference %q does not exist", path, ref)
	}
}

func ref[V any](p *problems, records map[string]V, path, kind string, id api.ContentID) (V, bool) {
	record, ok := records[string(id)]
	if !ok {
		p.addf("%s: unknown %s %q", path, kind, id)
	}
	return record, ok
}
