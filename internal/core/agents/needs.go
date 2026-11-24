package agents

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// spawnNeeds gives a body the needs its template starts it with.
func spawnNeeds(w *sim.World, e ecs.Entity, t content.Template) error {
	if len(t.Needs) == 0 {
		return nil
	}
	levels := make([]comp.NeedLevel, 0, len(t.Needs))
	for _, need := range slices.Sorted(maps.Keys(t.Needs)) {
		levels = append(levels, comp.NeedLevel{Need: api.ContentID(need), Level: t.Needs[need]})
	}
	ecs.NewMap[comp.Needs](w.ECS).Add(e, &comp.Needs{Levels: levels})
	return nil
}

// tickNeeds runs at a phase boundary: for every living body with needs, in UID order, and each of
// its needs in ID order, it raises the level by the need's rise and applies the thresholds that hold.
func tickNeeds(w *sim.World) error {
	calendar := w.Content.Calendar()
	phase := string(calendar.Phases[clock.New(calendar).Phase(w.Now)].ID)
	needs := ecs.NewMap[comp.Needs](w.ECS)
	var bodies []ecs.Entity
	query := ecs.NewFilter1[comp.Needs](w.ECS).Without(ecs.C[comp.Dead]()).Query()
	for query.Next() {
		bodies = append(bodies, query.Entity())
	}
	slices.SortFunc(bodies, func(a, b ecs.Entity) int { return cmp.Compare(w.UIDOf(a), w.UIDOf(b)) })
	dead := ecs.NewMap[comp.Dead](w.ECS)
	for _, body := range bodies {
		ids := make([]api.ContentID, 0, len(needs.Get(body).Levels))
		for _, level := range needs.Get(body).Levels {
			ids = append(ids, level.Need)
		}
		for _, id := range ids {
			// A threshold's effects may have killed the body or taken its needs.
			if !w.Alive(body) || dead.Has(body) || !needs.Has(body) {
				break
			}
			if err := tickNeed(w, body, id, phase); err != nil {
				return fmt.Errorf("agents: body %d: %w", w.UIDOf(body), err)
			}
		}
	}
	return nil
}

func tickNeed(w *sim.World, body ecs.Entity, id api.ContentID, phase string) error {
	needs := ecs.NewMap[comp.Needs](w.ECS)
	at, ok := find(needs.Get(body), id)
	if !ok {
		return nil
	}
	def, ok := w.Content.Need(id)
	// TODO(M3-02.F1): a need dropped by hot-reloaded content should not stop the game.
	if !ok {
		return fmt.Errorf("unknown need %q", id)
	}
	env := content.NeedEnv{
		Level:  needs.Get(body).Levels[at].Level,
		Phase:  phase,
		Source: effects.BodyEnvOf(w, body, ecs.Entity{}),
	}
	rise, err := w.Content.NeedRise(id, env)
	if err != nil {
		return err
	}
	level := clamp(env.Level + rise)
	needs.Get(body).Levels[at].Level = level
	ctx := &effects.Context{World: w, Source: body, Target: body}
	for i, threshold := range def.Thresholds {
		if level < threshold.At {
			break
		}
		if err := effects.Apply(ctx, threshold.Do); err != nil {
			return fmt.Errorf("need %q threshold %d: %w", id, i, err)
		}
	}
	return nil
}

func find(needs *comp.Needs, id api.ContentID) (int, bool) {
	return slices.BinarySearchFunc(needs.Levels, id, func(level comp.NeedLevel, id api.ContentID) int {
		return cmp.Compare(level.Need, id)
	})
}

func clamp(level float64) float64 { return min(max(level, 0), 1) }

func satisfyPrimitive() effects.Primitive {
	return effects.Primitive{
		Name: "satisfy", Params: []string{"need", "amount"},
		Kinds:         map[string]effects.Kind{"need": effects.KindText},
		VerifyContent: verifyNeed, Apply: applySatisfy,
	}
}

// applySatisfy lowers the target's level of the need by amount; a negative amount raises it.
func applySatisfy(ctx *effects.Context, p effects.Params) error {
	need, err := p.String("need")
	if err != nil {
		return err
	}
	amount, err := p.Float("amount")
	if err != nil {
		return err
	}
	w := ctx.World
	needs := ecs.NewMap[comp.Needs](w.ECS)
	if !w.Alive(ctx.Target) || !needs.Has(ctx.Target) {
		return nil
	}
	held := needs.Get(ctx.Target)
	if at, ok := find(held, api.ContentID(need)); ok {
		held.Levels[at].Level = clamp(held.Levels[at].Level - amount)
	}
	return nil
}

func verifyNeed(reg *content.Registry, raw map[string]string) error {
	need, ok := raw["need"]
	if !ok {
		return nil
	}
	if _, known := reg.Need(api.ContentID(need)); !known {
		return fmt.Errorf("unknown need %q", need)
	}
	return nil
}
