package sim

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// Spawner adds the components for one template field. Spawners run in registration order
// after the core components (Identity, Pos, Appearance).
type Spawner struct {
	Name  string
	Apply func(w *World, e ecs.Entity, t content.Template) error
}

// mappers caches the Ark component maps spawning uses.
type mappers struct {
	identity   *ecs.Map[comp.Identity]
	pos        *ecs.Map[comp.Pos]
	appearance *ecs.Map[comp.Appearance]
	party      *ecs.Map[comp.Party]
	memberOf   *ecs.Map[comp.MemberOf]
}

func newMappers(world *ecs.World) *mappers {
	return &mappers{
		identity:   ecs.NewMap[comp.Identity](world),
		pos:        ecs.NewMap[comp.Pos](world),
		appearance: ecs.NewMap[comp.Appearance](world),
		party:      ecs.NewMap[comp.Party](world),
		memberOf:   ecs.NewMap[comp.MemberOf](world),
	}
}

// RegisterSpawner adds a spawner to run after those already registered. Names must be unique.
func (w *World) RegisterSpawner(s Spawner) error {
	if s.Name == "" || s.Apply == nil {
		return errors.New("sim: a spawner needs a name and an Apply function")
	}
	for _, existing := range w.spawners {
		if existing.Name == s.Name {
			return fmt.Errorf("sim: spawner %q already registered", s.Name)
		}
	}
	w.spawners = append(w.spawners, s)
	return nil
}

// Spawn builds an entity from a template: Identity with a new UID, Pos when at is not nil,
// Appearance when the template has one, then whatever the spawners add.
// If a spawner fails, the entity is removed again and the error returned.
func (w *World) Spawn(template api.ContentID, at *comp.Pos) (ecs.Entity, error) {
	if w.Content == nil {
		return ecs.Entity{}, errors.New("sim: spawn: world has no content")
	}
	def, ok := w.Content.Template(template)
	if !ok {
		return ecs.Entity{}, fmt.Errorf("sim: spawn: unknown template %q", template)
	}
	uid := w.NewUID()
	e := w.maps.identity.NewEntity(&comp.Identity{UID: uid, Template: template})
	w.byUID[uid] = e
	if at != nil {
		w.maps.pos.Add(e, at)
	}
	if def.Appearance != nil {
		appearance := *def.Appearance
		w.maps.appearance.Add(e, &appearance)
	}
	for _, spawner := range w.spawners {
		if err := spawner.Apply(w, e, def); err != nil {
			w.Despawn(e)
			return ecs.Entity{}, fmt.Errorf("sim: spawn %q: spawner %q: %w", template, spawner.Name, err)
		}
	}
	return e, nil
}

// NewIdentified makes an entity that has only an Identity with a new UID, for things that are not
// spawned from a template of their own, such as body parts. The caller adds the rest.
func (w *World) NewIdentified(template api.ContentID) ecs.Entity {
	uid := w.NewUID()
	e := w.maps.identity.NewEntity(&comp.Identity{UID: uid, Template: template})
	w.byUID[uid] = e
	return e
}

// Despawn removes the entity and its UID index entry. The UID is never reused.
// Parts still attached to it (comp.PartOf) go too, children first; parts severed earlier are left alone.
// Despawning a dead entity does nothing.
func (w *World) Despawn(e ecs.Entity) {
	if !w.ECS.Alive(e) {
		return
	}
	// Collect before removing: Ark forbids changing the world during a query, and removing the
	// body first would reset each part's PartOf target to the zero entity.
	var attached []ecs.Entity
	query := ecs.NewFilter1[comp.PartOf](w.ECS).Query(ecs.Rel[comp.PartOf](e))
	for query.Next() {
		attached = append(attached, query.Entity())
	}
	for _, part := range attached {
		w.Despawn(part)
	}
	if w.maps.identity.Has(e) {
		delete(w.byUID, w.maps.identity.Get(e).UID)
	}
	w.ECS.RemoveEntity(e)
}

// ByUID finds the live entity with the UID.
func (w *World) ByUID(uid api.UID) (ecs.Entity, bool) {
	e, ok := w.byUID[uid]
	return e, ok
}

// UIDOf returns the entity's UID, or 0 when it has no Identity.
func (w *World) UIDOf(e ecs.Entity) api.UID {
	if !w.ECS.Alive(e) || !w.maps.identity.Has(e) {
		return 0
	}
	return w.maps.identity.Get(e).UID
}

// TemplateOf returns the entity's template ID, or "" when it is dead or has no Identity.
func (w *World) TemplateOf(e ecs.Entity) api.ContentID {
	if !w.Alive(e) || !w.maps.identity.Has(e) {
		return ""
	}
	return w.maps.identity.Get(e).Template
}

// CompareUID orders two entities by UID. An entity with no Identity has UID 0 and sorts first.
func (w *World) CompareUID(a, b ecs.Entity) int {
	return cmp.Compare(w.UIDOf(a), w.UIDOf(b))
}

// SortByUID sorts the entities by UID in place, keeping the order of equal ones, and returns them.
func (w *World) SortByUID(entities []ecs.Entity) []ecs.Entity {
	slices.SortStableFunc(entities, w.CompareUID)
	return entities
}

// Alive reports whether the entity exists. The zero entity is never alive.
func (w *World) Alive(e ecs.Entity) bool {
	return w.ECS.Alive(e)
}

// Put sets the entity's component to v, adding it if the entity has none.
func Put[T any](w *World, e ecs.Entity, v T) {
	components := ecs.NewMap[T](w.ECS)
	if components.Has(e) {
		*components.Get(e) = v
		return
	}
	components.Add(e, &v)
}

// registerCoreSpawners adds the spawners for the components this package owns.
func (w *World) registerCoreSpawners() {
	// The names are fixed and distinct, so registration cannot fail.
	_ = w.RegisterSpawner(Spawner{Name: "party", Apply: spawnParty})
	_ = w.RegisterSpawner(Spawner{Name: "party_members", Apply: spawnPartyMembers})
}

// spawnParty gives a party template its Party component, with the slot count copied from content.
func spawnParty(w *World, e ecs.Entity, t content.Template) error {
	if t.Party == nil {
		return nil
	}
	w.maps.party.Add(e, &comp.Party{
		Slots: w.Content.Party().Slots, Control: api.ControlFollow, ControlBy: api.ControlByPlayer,
	})
	return nil
}

// spawnPartyMembers spawns each member without a position, ties it to the party, and makes the first lead.
func spawnPartyMembers(w *World, e ecs.Entity, t content.Template) error {
	if t.Party == nil {
		return nil
	}
	members := make([]ecs.Entity, 0, len(t.Party.Members))
	for _, template := range t.Party.Members {
		member, err := w.Spawn(template, nil)
		if err != nil {
			for _, spawned := range members {
				w.Despawn(spawned)
			}
			return err
		}
		w.maps.memberOf.Add(member, &comp.MemberOf{}, e)
		members = append(members, member)
	}
	party := w.maps.party.Get(e)
	for _, member := range members {
		party.Order = append(party.Order, w.UIDOf(member))
	}
	if len(members) > 0 {
		party.Leader = party.Order[0]
	}
	return nil
}

// ReindexUIDs rebuilds the UID index from the Identity components. Loading a save restores entities
// without spawning them, so it calls this afterwards. It errors on a duplicate UID.
func (w *World) ReindexUIDs() error {
	w.byUID = map[api.UID]ecs.Entity{}
	query := ecs.NewFilter1[comp.Identity](w.ECS).Query()
	for query.Next() {
		e, identity := query.Entity(), query.Get()
		if _, dup := w.byUID[identity.UID]; dup {
			query.Close()
			return fmt.Errorf("sim: duplicate UID %d", identity.UID)
		}
		w.byUID[identity.UID] = e
	}
	return nil
}
