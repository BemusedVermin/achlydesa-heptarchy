package agents

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const affordanceSpawner = "affordances"

func installLandmarks(w *sim.World) error {
	if err := w.RegisterSpawner(sim.Spawner{Name: affordanceSpawner, Apply: spawnAffordances}); err != nil {
		return err
	}
	return effects.Register(w, effects.Primitive{
		Name: "afford", Params: []string{"bodies", "affordance"},
		Kinds:         map[string]effects.Kind{"bodies": effects.KindBodies, "affordance": effects.KindText},
		VerifyContent: verifyAffordance, Apply: applyAfford,
	})
}

func spawnAffordances(w *sim.World, e ecs.Entity, t content.Template) error {
	for _, tag := range t.Affordances {
		afford(w, e, tag)
	}
	return nil
}

func verifyAffordance(reg *content.Registry, raw map[string]string) error {
	return effects.KnownIDs(raw, map[string]func(api.ContentID) bool{"affordance": effects.Known(reg.Affordance)})
}

// applyAfford makes each body a landmark that gives the affordance, as a corpse gives the dead.
func applyAfford(ctx *effects.Context, p effects.Params) error {
	bodies, err := p.Bodies("bodies")
	if err != nil {
		return err
	}
	tag, err := p.String("affordance")
	if err != nil {
		return err
	}
	if err := verifyAffordance(ctx.World.Content, map[string]string{"affordance": tag}); err != nil {
		return err
	}
	for _, body := range bodies {
		afford(ctx.World, body, api.ContentID(tag))
	}
	return nil
}

// afford adds the tag to what the entity gives, keeping the tags in ID order and each once.
func afford(w *sim.World, e ecs.Entity, tag api.ContentID) {
	affordances := ecs.NewMap[comp.Affordance](w.ECS)
	if !affordances.Has(e) {
		affordances.Add(e, &comp.Affordance{})
	}
	given := affordances.Get(e)
	if at, held := slices.BinarySearch(given.Tags, tag); !held {
		given.Tags = slices.Insert(given.Tags, at, tag)
	}
}

// Landmark is a remembered landmark and the cell its knower believes it stands in.
type Landmark struct {
	Subject api.UID
	At      api.Cell
}

// Known lists the landmarks the body remembers giving the affordance, nearest its own cell first by
// where it believes each is, then by UID. A landmark it remembers giving something but not where it is
// is left out, as is everything when the body has no position.
func Known(w *sim.World, body ecs.Entity, tag api.ContentID) ([]Landmark, error) {
	cfg := w.Content.Agents()
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if cfg.GivesClaim == "" || !w.Alive(body) || !positions.Has(body) {
		return nil, nil
	}
	from := positions.Get(body).At
	var out []Landmark
	for _, gives := range factsOf(w, body, cfg.GivesClaim) {
		if !slices.Contains(strings.Split(gives.Value, givesSeparator), string(tag)) ||
			slices.ContainsFunc(out, func(l Landmark) bool { return l.Subject == gives.Subject }) {
			continue
		}
		where, ok := knowledge.Best(w, body, gives.Subject, cfg.AtClaim)
		if !ok {
			continue
		}
		at, err := knowledge.ParseCell(where.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, Landmark{Subject: gives.Subject, At: at})
	}
	slices.SortFunc(out, func(a, b Landmark) int {
		return cmp.Or(
			cmp.Compare(hexmap.CellDistance(from, a.At), hexmap.CellDistance(from, b.At)),
			cmp.Compare(a.Subject, b.Subject),
		)
	})
	return out, nil
}

// Ground is a cell a knower remembers and how decayed it believed the ground there was.
type Ground struct {
	At    api.Cell
	Level int64
}

// KnownGround lists the cells the body remembers the decay of, least decayed first, then nearest its
// own cell, then in cell order, so that a mind needing undecayed ground heads for the first.
func KnownGround(w *sim.World, body ecs.Entity) ([]Ground, error) {
	cfg := w.Content.Agents()
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if cfg.GroundClaim == "" || !w.Alive(body) || !positions.Has(body) {
		return nil, nil
	}
	from := positions.Get(body).At
	var out []Ground
	for _, fact := range factsOf(w, body, cfg.GroundClaim) {
		if fact.Cell == nil || slices.ContainsFunc(out, func(g Ground) bool { return g.At == *fact.Cell }) {
			continue
		}
		level, err := strconv.ParseInt(fact.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, Ground{At: *fact.Cell, Level: level})
	}
	slices.SortFunc(out, func(a, b Ground) int {
		return cmp.Or(
			cmp.Compare(a.Level, b.Level),
			cmp.Compare(hexmap.CellDistance(from, a.At), hexmap.CellDistance(from, b.At)),
			hexmap.CompareCells(a.At, b.At),
		)
	})
	return out, nil
}

// factsOf lists the body's facts of the claim, in the order the knowledge store keeps them.
func factsOf(w *sim.World, body ecs.Entity, claim api.ContentID) []api.Fact {
	known := ecs.NewMap[comp.Knowledge](w.ECS)
	if !known.Has(body) {
		return nil
	}
	var out []api.Fact
	for _, fact := range known.Get(body).Facts {
		if fact.Claim == claim {
			out = append(out, fact)
		}
	}
	return out
}
