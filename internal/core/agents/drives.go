package agents

import (
	"fmt"
	"hash/fnv"
	"maps"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	rngStream    = "agents"
	driveSystem  = "agents.drives"
	driveSpawner = "drives"
)

// Install registers the drive, affordance and mind spawners, the satisfy and afford effects, the
// per-phase drive tick, and the mind controller with its self picker. It needs effects and the
// timeline installed first.
func Install(w *sim.World) error {
	if err := w.RegisterSpawner(sim.Spawner{Name: driveSpawner, Apply: spawnDrives}); err != nil {
		return err
	}
	if err := installLandmarks(w); err != nil {
		return err
	}
	if err := installMovement(w); err != nil {
		return err
	}
	if err := installMind(w); err != nil {
		return err
	}
	if err := installLOD(w); err != nil {
		return err
	}
	err := effects.Register(w, effects.Primitive{
		Name: "satisfy", Params: []string{"bodies", "drive", "amount"},
		Kinds:         map[string]effects.Kind{"bodies": effects.KindBodies, "drive": effects.KindText},
		VerifyContent: verifyDrive, Apply: applySatisfy,
	})
	if err != nil {
		return err
	}
	return w.Systems.Register(sim.System{Name: driveSystem, Cadence: sim.PerPhase, Run: tickDrives})
}

// spawnDrives gives a body the drives its template lists, each with a weight of its own drawn
// around the template's.
func spawnDrives(w *sim.World, e ecs.Entity, t content.Template) error {
	if len(t.Drives) == 0 {
		return nil
	}
	gen := w.RNG.Derived(rngStream, spawnSalt(w, e))
	drives := comp.Drives{Levels: map[api.ContentID]float64{}, Weights: map[api.ContentID]float64{}}
	for _, id := range slices.Sorted(maps.Keys(t.Drives)) {
		weight, err := w.Content.DriveWeight(content.WeightEnv{Typical: t.Drives[id], Roll: gen.Float64()})
		if err != nil {
			return fmt.Errorf("agents: weight on %q: %w", id, err)
		}
		drives.Levels[api.ContentID(id)] = 0
		drives.Weights[api.ContentID(id)] = weight
	}
	ecs.NewMap[comp.Drives](w.ECS).Add(e, &drives)
	return nil
}

// spawnSalt picks an individual's stream by its template and where it spawned, so the order chunks
// load in never decides a weight. A body spawned nowhere, such as a party member, has only its UID,
// which game start hands out in a fixed order.
func spawnSalt(w *sim.World, e ecs.Entity) uint64 {
	hasher := fnv.New64a()
	_, _ = fmt.Fprint(hasher, w.TemplateOf(e))
	if positions := ecs.NewMap[comp.Pos](w.ECS); positions.Has(e) {
		at := positions.Get(e).At
		_, _ = fmt.Fprintf(hasher, "@%d,%d,%d", at.Hex.Q, at.Hex.R, at.Depth)
	} else {
		_, _ = fmt.Fprintf(hasher, "#%d", w.UIDOf(e))
	}
	return hasher.Sum64()
}

// tickDrives raises every living body's drives by their rise formulas, then applies the thresholds
// each level has reached. Bodies go in UID order and drives in ID order.
func tickDrives(w *sim.World) error {
	carriers := ecs.NewMap[comp.Drives](w.ECS)
	dead := ecs.NewMap[comp.Dead](w.ECS)
	var bodies []ecs.Entity
	query := ecs.NewFilter1[comp.Drives](w.ECS).Query()
	for query.Next() {
		bodies = append(bodies, query.Entity())
	}
	for _, body := range w.SortByUID(bodies) {
		if !w.Alive(body) || dead.Has(body) {
			continue
		}
		levels := carriers.Get(body).Levels
		weights := carriers.Get(body).Weights
		ids := slices.Sorted(maps.Keys(levels))
		for _, id := range ids {
			rise, err := w.Content.DriveRise(id, content.DriveEnv{Level: levels[id], Weight: weights[id]})
			if err != nil {
				return fmt.Errorf("agents: body %d: %w", w.UIDOf(body), err)
			}
			levels[id] = max(0, levels[id]+rise)
		}
		for _, id := range ids {
			if err := applyThresholds(w, body, id, levels[id]); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyThresholds runs the effects of each threshold the level has reached, until the body dies.
func applyThresholds(w *sim.World, body ecs.Entity, id api.ContentID, level float64) error {
	drive, ok := w.Content.Drive(id)
	if !ok {
		return fmt.Errorf("agents: body %d carries unknown drive %q", w.UIDOf(body), id)
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	for i, threshold := range drive.Thresholds {
		if level < threshold.At {
			continue
		}
		if !w.Alive(body) || dead.Has(body) {
			return nil
		}
		ctx := &effects.Context{World: w, Source: body, Target: body}
		if err := effects.Apply(ctx, threshold.Do); err != nil {
			return fmt.Errorf("agents: drive %q threshold %d on body %d: %w", id, i, w.UIDOf(body), err)
		}
	}
	return nil
}

// applySatisfy lowers each body's level of the drive by amount, to no less than 0. A negative amount
// raises it. A body that does not carry the drive is left alone.
func applySatisfy(ctx *effects.Context, p effects.Params) error {
	bodies, err := p.Bodies("bodies")
	if err != nil {
		return err
	}
	drive, err := p.String("drive")
	if err != nil {
		return err
	}
	amount, err := p.Float("amount")
	if err != nil {
		return err
	}
	if err := verifyDrive(ctx.World.Content, map[string]string{"drive": drive}); err != nil {
		return err
	}
	carriers := ecs.NewMap[comp.Drives](ctx.World.ECS)
	for _, body := range bodies {
		if !carriers.Has(body) {
			continue
		}
		levels := carriers.Get(body).Levels
		if level, carried := levels[api.ContentID(drive)]; carried {
			levels[api.ContentID(drive)] = max(0, level-amount)
		}
	}
	return nil
}

func verifyDrive(reg *content.Registry, raw map[string]string) error {
	return effects.KnownIDs(raw, map[string]func(api.ContentID) bool{"drive": effects.Known(reg.Drive)})
}
