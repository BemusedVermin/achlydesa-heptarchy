package content

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// idPattern is the shape of every content ID: dot-separated lower-case words.
var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

// Load reads the CUE package in dir (normally content/data), validates it and builds the registry.
// Every problem found is reported, each naming its path.
func Load(dir string) (*Registry, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("content: resolve %q: %w", dir, err)
	}
	instances := load.Instances([]string{"."}, &load.Config{Dir: abs})
	if len(instances) != 1 {
		return nil, fmt.Errorf("content: expected one CUE package in %q, found %d", dir, len(instances))
	}
	if instances[0].Err != nil {
		return nil, fmt.Errorf("content: load %q: %w", dir, instances[0].Err)
	}
	value := cuecontext.New().BuildInstance(instances[0])
	if err := value.Validate(cue.Concrete(true)); err != nil {
		return nil, fmt.Errorf("content: validate %q: %w", dir, err)
	}
	var bundle Bundle
	if err := value.Decode(&bundle); err != nil {
		return nil, fmt.Errorf("content: decode %q: %w", dir, err)
	}
	return newRegistry(bundle)
}

// newRegistry validates a decoded bundle, compiles its formulas and templates, and hashes it.
func newRegistry(bundle Bundle) (*Registry, error) {
	registry := &Registry{
		bundle:    bundle,
		moveCosts: map[api.ContentID]*vm.Program{},
		templates: map[api.ContentID]*template.Template{},
	}
	var problems []error
	problems = append(problems, registry.checkIDs()...)
	problems = append(problems, registry.compileText()...)
	problems = append(problems, registry.checkTextRefs()...)
	problems = append(problems, registry.checkCalendar()...)
	problems = append(problems, registry.compileFormulas()...)
	if err := errors.Join(problems...); err != nil {
		return nil, fmt.Errorf("content: invalid: %w", err)
	}
	encoded, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("content: hash: %w", err)
	}
	sum := sha256.Sum256(encoded)
	registry.hash = hex.EncodeToString(sum[:])
	return registry, nil
}

// sortedKeys returns a map's keys in order so problems are reported deterministically.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// checkID reports a key that is not a well-formed ID of the given kind ("material", "terrain", "text").
func checkID(collection, key, kind string) error {
	if !idPattern.MatchString(key) {
		return fmt.Errorf("%s[%q]: id does not match %s", collection, key, idPattern)
	}
	if !strings.HasPrefix(key, kind+".") {
		return fmt.Errorf("%s[%q]: id must start with %q", collection, key, kind+".")
	}
	return nil
}

func (r *Registry) checkIDs() []error {
	var problems []error
	for _, key := range sortedKeys(r.bundle.Materials) {
		if err := checkID("materials", key, "material"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Terrains) {
		if err := checkID("terrains", key, "terrain"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Text) {
		if err := checkID("text", key, "text"); err != nil {
			problems = append(problems, err)
		}
	}
	return problems
}

func (r *Registry) compileText() []error {
	var problems []error
	for _, key := range sortedKeys(r.bundle.Text) {
		parsed, err := template.New(key).Option("missingkey=zero").Parse(r.bundle.Text[key])
		if err != nil {
			problems = append(problems, fmt.Errorf("text[%q]: bad template: %w", key, err))
			continue
		}
		r.templates[api.ContentID(key)] = parsed
	}
	return problems
}

func (r *Registry) checkTextRefs() []error {
	var problems []error
	check := func(path string, ref api.ContentID) {
		if _, ok := r.bundle.Text[string(ref)]; !ok {
			problems = append(problems, fmt.Errorf("%s: text reference %q does not exist", path, ref))
		}
	}
	for _, key := range sortedKeys(r.bundle.Materials) {
		check(fmt.Sprintf("materials[%q].name", key), r.bundle.Materials[key].Name)
	}
	for _, key := range sortedKeys(r.bundle.Terrains) {
		check(fmt.Sprintf("terrains[%q].name", key), r.bundle.Terrains[key].Name)
	}
	return problems
}

func (r *Registry) checkCalendar() []error {
	var problems []error
	cal := r.bundle.Calendar
	if cal.TicksPerPhase < 1 {
		problems = append(
			problems,
			fmt.Errorf("calendar.ticks_per_phase: must be at least 1, got %d", cal.TicksPerPhase),
		)
	}
	if len(cal.Phases) == 0 {
		problems = append(problems, errors.New("calendar.phases: need at least one phase"))
	}
	seen := map[api.ContentID]bool{}
	for i, phase := range cal.Phases {
		path := fmt.Sprintf("calendar.phases[%d]", i)
		if err := checkID(path, string(phase.ID), "phase"); err != nil {
			problems = append(problems, err)
		}
		if seen[phase.ID] {
			problems = append(problems, fmt.Errorf("%s: duplicate phase %q", path, phase.ID))
		}
		seen[phase.ID] = true
		if _, ok := r.bundle.Text[string(phase.Name)]; !ok {
			problems = append(problems, fmt.Errorf("%s.name: text reference %q does not exist", path, phase.Name))
		}
	}
	return problems
}

func (r *Registry) compileFormulas() []error {
	var problems []error
	for _, key := range sortedKeys(r.bundle.Terrains) {
		terrain := r.bundle.Terrains[key]
		path := fmt.Sprintf("terrains[%q].move_cost", key)
		if terrain.MoveCost == "" {
			if terrain.Solid {
				problems = append(problems, fmt.Errorf("%s: required when the terrain is solid", path))
			}
			continue
		}
		program, err := expr.Compile(terrain.MoveCost, expr.Env(MoveEnv{}), expr.AsInt64())
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", path, err))
			continue
		}
		r.moveCosts[api.ContentID(key)] = program
	}
	return problems
}
