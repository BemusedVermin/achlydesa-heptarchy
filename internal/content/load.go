package content

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
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
	bundle, err := loadBundle(dir)
	if err != nil {
		return nil, err
	}
	return newRegistry(bundle)
}

func loadBundle(dir string) (Bundle, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Bundle{}, fmt.Errorf("content: resolve %q: %w", dir, err)
	}
	files, err := cueFiles(abs)
	if err != nil {
		return Bundle{}, err
	}
	instances := load.Instances(files, &load.Config{Dir: abs})
	if len(instances) != 1 {
		return Bundle{}, fmt.Errorf("content: expected one CUE package in %q, found %d", dir, len(instances))
	}
	if instances[0].Err != nil {
		return Bundle{}, fmt.Errorf("content: load %q: %w", dir, instances[0].Err)
	}
	value := cuecontext.New().BuildInstance(instances[0])
	if err := value.Validate(cue.Concrete(true)); err != nil {
		return Bundle{}, fmt.Errorf("content: validate %q: %w", dir, err)
	}
	var bundle Bundle
	if err := value.Decode(&bundle); err != nil {
		return Bundle{}, fmt.Errorf("content: decode %q: %w", dir, err)
	}
	return bundle, nil
}

// newRegistry validates a decoded bundle, compiles its formulas and templates, and hashes it.
// cueFiles lists every .cue file under dir, as paths relative to it in a stable order, so that
// subfolders such as maps/ are part of the same data package.
func cueFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".cue" {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, "./"+filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("content: list %q: %w", dir, err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("content: no CUE files in %q", dir)
	}
	return files, nil
}

func newRegistry(bundle Bundle) (*Registry, error) {
	registry := &Registry{
		bundle:    bundle,
		moveCosts: map[api.ContentID]*vm.Program{},
		templates: map[api.ContentID]*template.Template{},
		actions:   map[api.ContentID]compiledAction{},
		offers:    map[string]*vm.Program{},
		mods:      map[api.ContentID]compiledMods{},
		rulesUsed: map[string]bool{},
	}
	var problems []error
	problems = append(problems, registry.checkIDs()...)
	registry.buildRecords()
	problems = append(problems, registry.checkRules()...)
	problems = append(problems, registry.checkVocabulary()...)
	problems = append(problems, registry.compileText()...)
	problems = append(problems, registry.checkTextRefs()...)
	problems = append(problems, registry.checkKnowledge()...)
	problems = append(problems, registry.checkCalendar()...)
	problems = append(problems, registry.checkTemplates()...)
	problems = append(problems, registry.checkBodies()...)
	problems = append(problems, registry.checkActions()...)
	problems = append(problems, registry.checkTypes()...)
	problems = append(problems, registry.checkChoices()...)
	problems = append(problems, registry.checkTemplateActions()...)
	problems = append(problems, registry.checkTriggers()...)
	problems = append(problems, registry.checkParty()...)
	problems = append(problems, registry.checkMaps()...)
	problems = append(problems, registry.checkStart()...)
	problems = append(problems, registry.checkCombat()...)
	problems = append(problems, registry.compileFormulas()...)
	problems = append(problems, registry.compileTimeline()...)
	if err := errors.Join(problems...); err != nil {
		return nil, fmt.Errorf("content: invalid: %w", err)
	}
	registry.warnUnusedRules()
	registry.rulesUsed = nil
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
	for _, key := range sortedKeys(r.bundle.Templates) {
		if err := checkID("templates", key, "template"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Actions) {
		if err := checkID("actions", key, "action"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Maps) {
		if err := checkID("maps", key, "map"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Parts) {
		if err := checkID("parts", key, "part"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Claims) {
		if err := checkID("claims", key, "claim"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Sources) {
		if err := checkID("sources", key, "source"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Bodies) {
		if err := checkID("bodies", key, "body"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Types) {
		if err := checkID("types", key, "type"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Items) {
		if err := checkID("items", key, "item"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Factions) {
		if err := checkID("factions", key, "faction"); err != nil {
			problems = append(problems, err)
		}
	}
	for _, key := range sortedKeys(r.bundle.Statuses) {
		if err := checkID("statuses", key, "status"); err != nil {
			problems = append(problems, err)
		}
	}
	return problems
}

// checkKnowledge checks that claims and sources name existing text, and that every claim has the
// text that reports it in the log.
func (r *Registry) checkKnowledge() []error {
	var problems []error
	check := func(path string, ref api.ContentID) {
		if _, ok := r.bundle.Text[string(ref)]; !ok {
			problems = append(problems, fmt.Errorf("%s: text reference %q does not exist", path, ref))
		}
	}
	for _, key := range sortedKeys(r.bundle.Claims) {
		check(fmt.Sprintf("claims[%q].name", key), r.bundle.Claims[key].Name)
		check(fmt.Sprintf("claims[%q] log text", key), api.KnowledgeTextID(api.ContentID(key)))
		if kind := r.bundle.Claims[key].Value; kind != "" && !slices.Contains(ValueKinds, kind) {
			problems = append(problems, fmt.Errorf("claims[%q].value: unknown value kind %q", key, kind))
		}
	}
	for _, key := range sortedKeys(r.bundle.Sources) {
		check(fmt.Sprintf("sources[%q].name", key), r.bundle.Sources[key].Name)
	}
	return problems
}

// checkBodies checks every part template and body plan: references, integrity and grant kinds.
func (r *Registry) checkBodies() []error {
	var problems []error
	for _, key := range sortedKeys(r.bundle.Parts) {
		part := r.bundle.Parts[key]
		path := fmt.Sprintf("parts[%q]", key)
		if _, ok := r.bundle.Text[string(part.Name)]; !ok {
			problems = append(problems, fmt.Errorf("%s.name: text reference %q does not exist", path, part.Name))
		}
		if part.Slot == "" {
			problems = append(problems, fmt.Errorf("%s.slot: required", path))
		}
		if _, ok := r.bundle.Materials[string(part.Material)]; !ok {
			problems = append(problems, fmt.Errorf("%s.material: unknown material %q", path, part.Material))
		}
		if part.Integrity < 1 {
			problems = append(problems, fmt.Errorf("%s.integrity: must be at least 1, got %d", path, part.Integrity))
		}
		for i, grant := range part.Grants {
			if grant.Kind == "" {
				problems = append(problems, fmt.Errorf("%s.grants[%d].kind: required", path, i))
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Bodies) {
		plan := r.bundle.Bodies[key]
		path := fmt.Sprintf("bodies[%q].parts", key)
		if len(plan.Parts) == 0 {
			problems = append(problems, fmt.Errorf("%s: a body needs at least one part", path))
		}
		for i, entry := range plan.Parts {
			if _, ok := r.bundle.Parts[string(entry.Part)]; !ok {
				problems = append(problems, fmt.Errorf("%s[%d].part: unknown part %q", path, i, entry.Part))
			}
			if entry.Material == "" {
				continue
			}
			if _, ok := r.bundle.Materials[string(entry.Material)]; !ok {
				problems = append(problems, fmt.Errorf("%s[%d].material: unknown material %q", path, i, entry.Material))
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Templates) {
		def := r.bundle.Templates[key]
		path := fmt.Sprintf("templates[%q]", key)
		if def.Body != nil {
			if _, ok := r.bundle.Bodies[string(*def.Body)]; !ok {
				problems = append(problems, fmt.Errorf("%s.body: unknown body plan %q", path, *def.Body))
			}
		}
		if def.Name != nil {
			if _, ok := r.bundle.Text[string(*def.Name)]; !ok {
				problems = append(problems, fmt.Errorf("%s.name: text reference %q does not exist", path, *def.Name))
			}
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
		if phase.Ticks < 1 {
			problems = append(problems, fmt.Errorf("%s.ticks: must be at least 1, got %d", path, phase.Ticks))
		}
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
		program, err := r.Compile(terrain.MoveCost, MoveEnv{}, expr.AsInt64())
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", path, err))
			continue
		}
		r.moveCosts[api.ContentID(key)] = program
	}
	return problems
}

// compileTimeline compiles step_ticks and checks the timeline numbers.
func (r *Registry) compileTimeline() []error {
	cfg := r.bundle.Timeline
	var problems []error
	if cfg.DefaultSpeed < 0 {
		problems = append(
			problems,
			fmt.Errorf("timeline.default_speed: must not be negative, got %d", cfg.DefaultSpeed),
		)
	}
	if cfg.WindowTicks < 1 {
		problems = append(problems, fmt.Errorf("timeline.window_ticks: must be at least 1, got %d", cfg.WindowTicks))
	}
	program, err := r.Compile(cfg.StepTicks, SpeedEnv{}, expr.AsInt64())
	if err != nil {
		problems = append(problems, fmt.Errorf("timeline.step_ticks: %w", err))
	} else {
		r.stepTicks = program
	}
	if speed, err := r.Compile(cfg.Speed, GrantEnv{}, expr.AsInt64()); err != nil {
		problems = append(problems, fmt.Errorf("timeline.speed: %w", err))
	} else {
		r.speed = speed
	}
	return problems
}

func (r *Registry) checkTemplates() []error {
	var problems []error
	for _, key := range sortedKeys(r.bundle.Templates) {
		if actor := r.bundle.Templates[key].Actor; actor != nil && actor.Controller == "" {
			problems = append(problems, fmt.Errorf("templates[%q].actor.controller: required", key))
		}
		def := r.bundle.Templates[key].Party
		if def == nil {
			continue
		}
		path := fmt.Sprintf("templates[%q].party.members", key)
		if len(def.Members) == 0 {
			problems = append(problems, fmt.Errorf("%s: a party needs at least one member", path))
		}
		for i, member := range def.Members {
			target, ok := r.bundle.Templates[string(member)]
			switch {
			case !ok:
				problems = append(problems, fmt.Errorf("%s[%d]: unknown template %q", path, i, member))
			case target.Party != nil:
				problems = append(problems, fmt.Errorf("%s[%d]: member %q is itself a party", path, i, member))
			}
		}
	}
	return problems
}

func (r *Registry) checkParty() []error {
	var problems []error
	cfg := r.bundle.Party
	if cfg.Slots < 1 {
		problems = append(problems, fmt.Errorf("party.slots: must be at least 1, got %d", cfg.Slots))
	}
	if cfg.SightRadius < 1 {
		problems = append(problems, fmt.Errorf("party.sight_radius: must be at least 1, got %d", cfg.SightRadius))
	}
	for name, value := range map[string]int{
		"sight_depth": cfg.SightDepth, "plan_range": cfg.PlanRange, "max_step_up": cfg.MaxStepUp, "max_step_down": cfg.MaxStepDown,
		"max_drop": cfg.MaxDrop,
	} {
		if value < 0 {
			problems = append(problems, fmt.Errorf("party.%s: must not be negative, got %d", name, value))
		}
	}
	follow, ok := r.bundle.Actions[string(cfg.FollowAction)]
	switch {
	case !ok:
		problems = append(problems, fmt.Errorf("party.follow_action: unknown action %q", cfg.FollowAction))
	case follow.Target != TargetDirection:
		problems = append(problems, fmt.Errorf("party.follow_action: %q must target a direction", cfg.FollowAction))
	}
	if cfg.ClimbCost != "" {
		program, err := r.Compile(cfg.ClimbCost, SpeedEnv{}, expr.AsInt64())
		if err != nil {
			return append(problems, fmt.Errorf("party.climb_cost: %w", err))
		}
		r.climbCost = program
	}
	return problems
}
