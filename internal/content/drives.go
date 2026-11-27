package content

import (
	"errors"
	"fmt"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// DriveKind says whether a drive is a need or a desire.
type DriveKind string

// Kinds of drive.
const (
	// DriveNeed rises with time and can kill.
	DriveNeed DriveKind = "need"
	// DriveDesire never kills: it carries no thresholds.
	DriveDesire DriveKind = "desire"
)

// driveKinds lists the kinds a drive may name, for bundles built in Go, which skip the schema.
func driveKinds() []DriveKind { return []DriveKind{DriveNeed, DriveDesire} }

// Drive is something an agent wants, such as water or pay. A body carries the drives its template
// lists, each with a level that is 0 when met and grows as it goes unmet.
type Drive struct {
	Data
	// Name is a text ID.
	Name api.ContentID `json:"name"`
	Kind DriveKind     `json:"kind"`
	// Rise is an Expr over DriveEnv: how much the level rises each phase. A need requires one; a desire
	// without one keeps its level until something changes it.
	Rise string `json:"rise,omitempty"`
	// Urge is an Expr over DriveEnv: how pressing the drive is at its level, from 0 to 1.
	Urge string `json:"urge"`
	// Thresholds apply effects each phase while the level is at or above them. Only a need has them.
	Thresholds []Threshold `json:"thresholds,omitempty"`
}

// Threshold is a level of a need from which its effects apply, once each phase it holds. The body
// carrying the need is both source and target of the effects.
type Threshold struct {
	At float64     `json:"at"`
	Do []EffectRef `json:"do"`
}

// DriveEnv is what a drive's rise and urge formulas see.
type DriveEnv struct {
	Level float64 `expr:"level"`
	// Weight is the individual's own weight on the drive.
	Weight float64 `expr:"weight"`
}

// WeightEnv is what agents.weight sees, for one drive of one individual.
type WeightEnv struct {
	// Typical is the template's weight on the drive.
	Typical float64 `expr:"typical"`
	// Roll is a fresh draw in [0,1) from the individual's own stream.
	Roll float64 `expr:"roll"`
}

type compiledDrive struct {
	// rise is nil for a desire without one.
	rise *vm.Program
	urge *vm.Program
}

// Drive returns the drive with the given ID. Its Thresholds slice is shared with the registry: don't
// change it.
func (r *Registry) Drive(id api.ContentID) (Drive, bool) {
	drive, ok := r.bundle.Drives[string(id)]
	return drive, ok
}

// DriveRise is how much the drive's level rises this phase; a drive without a rise formula gives 0.
func (r *Registry) DriveRise(id api.ContentID, env DriveEnv) (float64, error) {
	compiled, ok := r.drives[id]
	if !ok {
		return 0, fmt.Errorf("content: unknown drive %q", id)
	}
	if compiled.rise == nil {
		return 0, nil
	}
	return run[float64](compiled.rise, env, fmt.Sprintf("drive %q rise", id))
}

// DriveUrge is how pressing the drive is at env's level. A result outside [0, 1] is a content bug.
func (r *Registry) DriveUrge(id api.ContentID, env DriveEnv) (float64, error) {
	compiled, ok := r.drives[id]
	if !ok {
		return 0, fmt.Errorf("content: unknown drive %q", id)
	}
	urge, err := run[float64](compiled.urge, env, fmt.Sprintf("drive %q urge", id))
	if err != nil {
		return 0, err
	}
	if urge < 0 || urge > 1 {
		return 0, fmt.Errorf("content: drive %q urge is %v, want 0 to 1", id, urge)
	}
	return urge, nil
}

// DriveWeight draws an individual's weight on a drive by agents.weight. A negative result is a content bug.
func (r *Registry) DriveWeight(env WeightEnv) (float64, error) {
	if r.weight == nil {
		return 0, errors.New("content: agents.weight is empty")
	}
	weight, err := run[float64](r.weight, env, "agents.weight")
	if err != nil {
		return 0, err
	}
	if weight < 0 {
		return 0, fmt.Errorf("content: agents.weight is %v for typical %v, want at least 0", weight, env.Typical)
	}
	return weight, nil
}

// checkDrives checks and compiles the drives, the drives each template lists, and agents.weight.
func (r *Registry) checkDrives() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Drives) {
		drive := r.bundle.Drives[key]
		path := fmt.Sprintf("drives[%q]", key)
		p.text(path+".name", drive.Name)
		if !slices.Contains(driveKinds(), drive.Kind) {
			p.addf("%s.kind: unknown kind %q (want one of %v)", path, drive.Kind, driveKinds())
		}
		var compiled compiledDrive
		switch {
		case drive.Rise != "":
			program, err := r.Compile(drive.Rise, DriveEnv{}, expr.AsFloat64())
			if err != nil {
				p.addf("%s.rise: %w", path, err)
			}
			compiled.rise = program
		case drive.Kind == DriveNeed:
			p.addf("%s.rise: required for a need", path)
		}
		program, err := r.Compile(drive.Urge, DriveEnv{}, expr.AsFloat64())
		if err != nil {
			p.addf("%s.urge: %w", path, err)
		}
		compiled.urge = program
		if drive.Kind == DriveDesire && len(drive.Thresholds) > 0 {
			p.addf("%s.thresholds: a desire never kills, so it has none", path)
		}
		for i, threshold := range drive.Thresholds {
			if len(threshold.Do) == 0 {
				p.addf("%s.thresholds[%d].do: at least one effect is required", path, i)
			}
		}
		r.drives[api.ContentID(key)] = compiled
	}
	listed := false
	for _, key := range sortedKeys(r.bundle.Templates) {
		for _, id := range sortedKeys(r.bundle.Templates[key].Drives) {
			listed = true
			path := fmt.Sprintf("templates[%q].drives[%q]", key, id)
			ref(p, r.bundle.Drives, path, "drive", api.ContentID(id))
			if weight := r.bundle.Templates[key].Drives[id]; weight < 0 {
				p.addf("%s: a weight must not be negative, got %v", path, weight)
			}
		}
	}
	switch weight := r.bundle.Agents.Weight; {
	case weight != "":
		program, err := r.Compile(weight, WeightEnv{}, expr.AsFloat64())
		if err != nil {
			p.addf("agents.weight: %w", err)
		}
		r.weight = program
	case listed:
		p.addf("agents.weight: required when a template lists drives")
	}
	return p.errs
}

func driveSites(drives map[string]Drive) []EffectSite {
	var sites []EffectSite
	for _, id := range sortedKeys(drives) {
		for i, threshold := range drives[id].Thresholds {
			if len(threshold.Do) > 0 {
				sites = append(sites, EffectSite{
					Path: fmt.Sprintf("drives[%q].thresholds[%d].do", id, i), Refs: threshold.Do,
				})
			}
		}
	}
	return sites
}
