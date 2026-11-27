package content

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
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

// cueFiles lists every .cue file under dir, as paths relative to it in a stable order, so that
// subfolders such as places/ are part of the same data package.
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

// newRegistry validates a decoded bundle, compiles its formulas and templates, and hashes it.
func newRegistry(bundle Bundle) (*Registry, error) {
	registry := &Registry{
		bundle:      bundle,
		moveCosts:   map[api.ContentID]*vm.Program{},
		decays:      map[api.ContentID]*vm.Program{},
		templates:   map[api.ContentID]*template.Template{},
		actions:     map[api.ContentID]compiledAction{},
		offers:      map[string]*vm.Program{},
		mods:        map[api.ContentID]compiledMods{},
		drives:      map[api.ContentID]compiledDrive{},
		rulesUsed:   map[string]bool{},
		collections: collections(),
	}
	var problems []error
	problems = append(problems, registry.checkIDs()...)
	registry.buildRecords()
	problems = append(problems, registry.checkRules()...)
	problems = append(problems, registry.checkVocabulary()...)
	problems = append(problems, registry.compileText()...)
	problems = append(problems, registry.checkTextNames()...)
	problems = append(problems, registry.checkKnowledge()...)
	problems = append(problems, registry.checkCalendar()...)
	problems = append(problems, registry.checkTemplates()...)
	problems = append(problems, registry.checkBodies()...)
	problems = append(problems, registry.checkActions()...)
	problems = append(problems, registry.checkTypes()...)
	problems = append(problems, registry.checkChoices()...)
	problems = append(problems, registry.checkTemplateActions()...)
	problems = append(problems, registry.checkTriggers()...)
	problems = append(problems, registry.checkDrives()...)
	problems = append(problems, registry.checkParty()...)
	problems = append(problems, registry.checkWorld()...)
	problems = append(problems, registry.checkGeographies()...)
	problems = append(problems, registry.checkStart()...)
	problems = append(problems, registry.checkCombat()...)
	problems = append(problems, registry.checkTerrains()...)
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
func sortedKeys[V any](m map[string]V) []string { return slices.Sorted(maps.Keys(m)) }

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

// checkKnowledge checks that claims and origins name existing text, and that every claim has the
// text that reports it in the log, then the numbers and claims of memory and sight.
func (r *Registry) checkKnowledge() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Claims) {
		p.text(fmt.Sprintf("claims[%q].name", key), r.bundle.Claims[key].Name)
		p.text(fmt.Sprintf("claims[%q] log text", key), api.KnowledgeTextID(api.ContentID(key)))
	}
	for _, key := range sortedKeys(r.bundle.Origins) {
		p.text(fmt.Sprintf("origins[%q].name", key), r.bundle.Origins[key].Name)
	}
	if r.bundle.Knowledge.Cap < 1 {
		p.addf("knowledge.cap: must be at least 1, got %d", r.bundle.Knowledge.Cap)
	}
	agents := r.bundle.Agents
	if agents.SightRadius < 0 || agents.SightDepth < 0 {
		p.addf("agents: sight_radius and sight_depth must not be negative, got %d and %d",
			agents.SightRadius, agents.SightDepth)
	}
	if agents.SeenConfidence < 0 || agents.SeenConfidence > 1 {
		p.addf("agents.seen_confidence: must be from 0 to 1, got %v", agents.SeenConfidence)
	}
	if agents.AtClaim != "" {
		ref(p, r.bundle.Claims, "agents.at_claim", "claim", agents.AtClaim)
	}
	if agents.DeadClaim != "" {
		ref(p, r.bundle.Claims, "agents.dead_claim", "claim", agents.DeadClaim)
	}
	if agents.AtClaim != "" || agents.DeadClaim != "" {
		ref(p, r.bundle.Origins, "agents.seen_origin", "origin", agents.SeenOrigin)
	}
	return p.errs
}

// checkCrawlsAs checks the template a crawling part acts as. It must exist, and declare none of the
// fields spawn_from_part would ignore: the part is its whole body, the effect's parameters give its
// faction and controller, and only its identity and appearance come from the template.
func (r *Registry) checkCrawlsAs(path string, id api.ContentID) []error {
	p := r.problems()
	def, ok := ref(p, r.bundle.Templates, path+".crawls_as", "template", id)
	if !ok {
		return p.errs
	}
	ignored := map[string]bool{
		"body": def.Body != nil, "party": def.Party != nil, "actor": def.Actor != nil,
		"faction": def.Faction != nil, "protected": def.Protected, "inventory": len(def.Inventory) > 0,
		"triggers": len(def.Triggers) > 0, "drives": len(def.Drives) > 0,
	}
	for _, field := range sortedKeys(ignored) {
		if ignored[field] {
			p.addf("%s.crawls_as: template %q sets %s, which a crawling part ignores", path, id, field)
		}
	}
	return p.errs
}

// checkBodies checks every part template and body plan: references, integrity and grant kinds.
func (r *Registry) checkBodies() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Parts) {
		part := r.bundle.Parts[key]
		path := fmt.Sprintf("parts[%q]", key)
		p.text(path+".name", part.Name)
		if part.Slot == "" {
			p.addf("%s.slot: required", path)
		}
		ref(p, r.bundle.Materials, path+".material", "material", part.Material)
		if part.Integrity < 1 {
			p.addf("%s.integrity: must be at least 1, got %d", path, part.Integrity)
		}
		for i, grant := range part.Grants {
			if grant.Kind == "" {
				p.addf("%s.grants[%d].kind: required", path, i)
			}
		}
		if part.CrawlsAs != "" {
			p.add(r.checkCrawlsAs(path, part.CrawlsAs)...)
		}
	}
	for _, key := range sortedKeys(r.bundle.Bodies) {
		plan := r.bundle.Bodies[key]
		path := fmt.Sprintf("bodies[%q].parts", key)
		if len(plan.Parts) == 0 {
			p.addf("%s: a body needs at least one part", path)
		}
		for i, entry := range plan.Parts {
			ref(p, r.bundle.Parts, fmt.Sprintf("%s[%d].part", path, i), "part", entry.Part)
			if entry.Material != "" {
				ref(p, r.bundle.Materials, fmt.Sprintf("%s[%d].material", path, i), "material", entry.Material)
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Templates) {
		def := r.bundle.Templates[key]
		path := fmt.Sprintf("templates[%q]", key)
		if def.Body != nil {
			ref(p, r.bundle.Bodies, path+".body", "body plan", *def.Body)
		}
		if def.Name != nil {
			p.text(path+".name", *def.Name)
		}
	}
	return p.errs
}

func (r *Registry) compileText() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Text) {
		parsed, err := template.New(key).Option("missingkey=zero").Parse(r.bundle.Text[key])
		if err != nil {
			p.addf("text[%q]: bad template: %w", key, err)
			continue
		}
		r.templates[api.ContentID(key)] = parsed
	}
	return p.errs
}

func (r *Registry) checkCalendar() []error {
	p := r.problems()
	cal := r.bundle.Calendar
	if len(cal.Phases) == 0 {
		p.addf("calendar.phases: need at least one phase")
	}
	seen := map[api.ContentID]bool{}
	for i, phase := range cal.Phases {
		path := fmt.Sprintf("calendar.phases[%d]", i)
		if err := checkID(path, string(phase.ID), "phase"); err != nil {
			p.add(err)
		}
		if seen[phase.ID] {
			p.addf("%s: duplicate phase %q", path, phase.ID)
		}
		seen[phase.ID] = true
		if phase.Ticks < 1 {
			p.addf("%s.ticks: must be at least 1, got %d", path, phase.Ticks)
		}
		p.text(path+".name", phase.Name)
	}
	exports := cal.Exports
	if exports.FirstDay < 1 {
		p.addf("calendar.exports.first_day: must be at least 1, got %d", exports.FirstDay)
	}
	if exports.MinGap < 1 {
		p.addf("calendar.exports.min_gap: must be at least 1, got %d", exports.MinGap)
	}
	if exports.MaxGap < exports.MinGap {
		p.addf("calendar.exports.max_gap: must be at least min_gap (%d), got %d", exports.MinGap, exports.MaxGap)
	}
	return p.errs
}

func (r *Registry) checkTerrains() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Terrains) {
		terrain := r.bundle.Terrains[key]
		p.text(fmt.Sprintf("terrains[%q].name", key), terrain.Name)
		path := fmt.Sprintf("terrains[%q].move_cost", key)
		if terrain.MoveCost == "" {
			if terrain.Solid {
				p.addf("%s: required when the terrain is solid", path)
			}
			continue
		}
		program, err := r.Compile(terrain.MoveCost, MoveEnv{}, expr.AsInt64())
		if err != nil {
			p.addf("%s: %w", path, err)
			continue
		}
		r.moveCosts[api.ContentID(key)] = program
	}
	return p.errs
}

// compileTimeline compiles step_ticks and checks the timeline numbers.
func (r *Registry) compileTimeline() []error {
	cfg := r.bundle.Timeline
	p := r.problems()
	if cfg.DefaultSpeed < 0 {
		p.addf("timeline.default_speed: must not be negative, got %d", cfg.DefaultSpeed)
	}
	if cfg.WindowTicks < 1 {
		p.addf("timeline.window_ticks: must be at least 1, got %d", cfg.WindowTicks)
	}
	program, err := r.Compile(cfg.StepTicks, SpeedEnv{}, expr.AsInt64())
	if err != nil {
		p.addf("timeline.step_ticks: %w", err)
	} else {
		r.stepTicks = program
	}
	if speed, err := r.Compile(cfg.Speed, GrantEnv{}, expr.AsInt64()); err != nil {
		p.addf("timeline.speed: %w", err)
	} else {
		r.speed = speed
	}
	return p.errs
}

func (r *Registry) checkTemplates() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Templates) {
		if actor := r.bundle.Templates[key].Actor; actor != nil && actor.Controller == "" {
			p.addf("templates[%q].actor.controller: required", key)
		}
		def := r.bundle.Templates[key].Party
		if def == nil {
			continue
		}
		path := fmt.Sprintf("templates[%q].party.members", key)
		if len(def.Members) == 0 {
			p.addf("%s: a party needs at least one member", path)
		}
		for i, member := range def.Members {
			target, ok := ref(p, r.bundle.Templates, fmt.Sprintf("%s[%d]", path, i), "template", member)
			switch {
			case !ok:
			case target.Party != nil:
				p.addf("%s[%d]: member %q is itself a party", path, i, member)
			}
		}
	}
	return p.errs
}

func (r *Registry) checkParty() []error {
	p := r.problems()
	cfg := r.bundle.Party
	if cfg.Slots < 1 {
		p.addf("party.slots: must be at least 1, got %d", cfg.Slots)
	}
	if cfg.SightRadius < 1 {
		p.addf("party.sight_radius: must be at least 1, got %d", cfg.SightRadius)
	}
	counts := map[string]int{
		"sight_depth":   cfg.SightDepth,
		"plan_range":    cfg.PlanRange,
		"max_step_up":   cfg.MaxStepUp,
		"max_step_down": cfg.MaxStepDown,
		"max_drop":      cfg.MaxDrop,
	}
	for _, name := range sortedKeys(counts) {
		if counts[name] < 0 {
			p.addf("party.%s: must not be negative, got %d", name, counts[name])
		}
	}
	if follow, ok := ref(p, r.bundle.Actions, "party.follow_action", "action", cfg.FollowAction); ok &&
		follow.Target != TargetDirection {
		p.addf("party.follow_action: %q must target a direction", cfg.FollowAction)
	}
	if cfg.ClimbCost != "" {
		program, err := r.Compile(cfg.ClimbCost, SpeedEnv{}, expr.AsInt64())
		if err != nil {
			p.addf("party.climb_cost: %w", err)
			return p.errs
		}
		r.climbCost = program
	}
	return p.errs
}
