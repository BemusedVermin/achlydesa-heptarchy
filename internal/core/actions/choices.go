package actions

import (
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// MaxChannelSets caps how many ways one action may pick its channels, so a body with many limbs does
// not flood the options. Sets come in channel order, so the cap keeps the lowest ones.
const MaxChannelSets = 32

// directionChoice is the built-in choice provider: the six directions.
const directionChoice = "direction"

// dirNames names the six directions, by api.Dir value.
var dirNames = [api.DirCount]string{"e", "ne", "nw", "w", "sw", "se"}

// ChoiceProvider offers the values a choice may take for an actor, in a stable order. Only choices
// that are not records need one, such as the directions; the rest are content choices (see
// content.ChoiceDef).
type ChoiceProvider func(w *sim.World, actor ecs.Entity) []string

// RegisterChoice makes a choice provider available to actions by name. Names are unique.
func RegisterChoice(w *sim.World, name string, p ChoiceProvider) error {
	if name == "" || p == nil {
		return errors.New("actions: a choice provider needs a name and a function")
	}
	st := stateOf(w)
	if _, dup := st.providers[name]; dup {
		return fmt.Errorf("actions: choice provider %q already registered", name)
	}
	st.providers[name] = p
	return nil
}

// choiceValues lists the values the actor may pick for a choice, in order: the Go provider's, or
// else the records the content choice offers, in ID order.
func (ev *evaluator) choiceValues(name string, actor ecs.Entity) ([]string, error) {
	key := valueKey{actor: actor, name: name}
	if values, ok := ev.valueMemo[key]; ok {
		return values, nil
	}
	values, err := ev.offeredValues(name, actor)
	if err != nil {
		return nil, err
	}
	if ev.valueMemo == nil {
		ev.valueMemo = map[valueKey][]string{}
	}
	ev.valueMemo[key] = values
	return values, nil
}

// offeredValues works out the values choiceValues remembers.
func (ev *evaluator) offeredValues(name string, actor ecs.Entity) ([]string, error) {
	if provider, ok := ev.st.providers[name]; ok {
		return provider(ev.w, actor), nil
	}
	reg := ev.w.Content
	if _, ok := reg.Choice(name); !ok {
		return nil, nil
	}
	if ev.inventory == nil {
		ev.inventory = map[ecs.Entity]map[string]int{}
	}
	held, ok := ev.inventory[actor]
	if !ok {
		held = inventory.Usable(ev.w, actor, ecs.Entity{})
		ev.inventory[actor] = held
	}
	source := effects.BodyEnvOf(ev.w, actor, ecs.Entity{})
	values := []string{}
	for _, id := range reg.ChoiceRecords(name) {
		record, _ := reg.Record(id)
		offered, err := reg.Offered(name, content.OfferEnv{Record: record, Source: source, Inventory: held})
		if err != nil {
			return nil, err
		}
		if offered {
			values = append(values, string(id))
		}
	}
	return values, nil
}

func directionValues(*sim.World, ecs.Entity) []string { return dirNames[:] }

func hasTag(c timeline.Channel, tag string) bool { return slices.Contains(c.Tags, tag) }

// channelSets lists the ways the actor can meet the action's requirements, as sets of channels in the
// actor's channel order, up to limit. With requireFree, busy channels do not count.
func (ev *evaluator) channelSets(actor ecs.Entity, def content.ActionDef, requireFree bool, limit int,
) [][]timeline.Channel {
	all := timeline.Channels(ev.w, actor)
	var sets [][]timeline.Channel
	var walk func(i int, chosen []timeline.Channel)
	walk = func(i int, chosen []timeline.Channel) {
		if len(sets) >= limit {
			return
		}
		if i == len(def.Uses) {
			set := slices.Clone(chosen)
			slices.SortStableFunc(set, func(a, b timeline.Channel) int { return indexOf(all, a) - indexOf(all, b) })
			sets = append(sets, set)
			return
		}
		req := def.Uses[i]
		var matching []timeline.Channel
		total := 0
		for _, c := range all {
			if !hasTag(c, req.Tag) {
				continue
			}
			total++
			taken := slices.ContainsFunc(chosen, func(o timeline.Channel) bool { return o.Entity == c.Entity })
			if !taken && (!requireFree || timeline.Free(ev.w, c)) {
				matching = append(matching, c)
			}
		}
		need := req.Count
		if need == 0 {
			need = total // every one the body has, and all of them must be available
			if len(matching) != total {
				return
			}
		}
		if need == 0 || len(matching) < need {
			return
		}
		combinations(matching, need, func(combo []timeline.Channel) bool {
			walk(i+1, append(slices.Clone(chosen), combo...))
			return len(sets) < limit
		})
	}
	walk(0, nil)
	return sets
}

func indexOf(all []timeline.Channel, c timeline.Channel) int {
	return slices.IndexFunc(all, func(o timeline.Channel) bool { return o.Entity == c.Entity })
}

// combinations calls fn with each way to pick k of items, in order, until fn returns false.
func combinations(items []timeline.Channel, k int, fn func([]timeline.Channel) bool) {
	combo := make([]timeline.Channel, 0, k)
	var pick func(start int) bool
	pick = func(start int) bool {
		if len(combo) == k {
			return fn(slices.Clone(combo))
		}
		for i := start; i <= len(items)-(k-len(combo)); i++ {
			combo = append(combo, items[i])
			if !pick(i + 1) {
				return false
			}
			combo = combo[:len(combo)-1]
		}
		return true
	}
	pick(0)
}

// assign settles which channels the intent uses: the ones it names, if they fit the requirements, or
// else the first set that does.
func (ev *evaluator) assign(in Intent, def content.ActionDef, requireFree bool) ([]timeline.Channel, api.Text) {
	if len(in.Channels) == 0 {
		sets := ev.channelSets(in.Actor, def, requireFree, 1)
		if len(sets) == 0 {
			return nil, why("channels_busy")
		}
		return sets[0], api.Text{}
	}
	all := timeline.Channels(ev.w, in.Actor)
	var given []timeline.Channel
	for _, named := range in.Channels {
		i := indexOf(all, named)
		if i < 0 || slices.ContainsFunc(given, func(c timeline.Channel) bool { return c.Entity == named.Entity }) {
			return nil, why("bad_channels")
		}
		if requireFree && !timeline.Free(ev.w, all[i]) {
			return nil, why("channels_busy")
		}
		given = append(given, all[i])
	}
	remaining := slices.Clone(given)
	for _, req := range def.Uses {
		var matching []timeline.Channel
		for _, c := range remaining {
			if hasTag(c, req.Tag) {
				matching = append(matching, c)
			}
		}
		need := req.Count
		if need == 0 {
			need = len(
				slices.DeleteFunc(slices.Clone(all), func(c timeline.Channel) bool { return !hasTag(c, req.Tag) }),
			)
		}
		if need == 0 || len(matching) != need {
			return nil, why("bad_channels")
		}
		remaining = slices.DeleteFunc(remaining, func(c timeline.Channel) bool {
			return slices.ContainsFunc(matching, func(m timeline.Channel) bool { return m.Entity == c.Entity })
		})
	}
	if len(remaining) != 0 {
		return nil, why("bad_channels")
	}
	slices.SortStableFunc(given, func(a, b timeline.Channel) int { return indexOf(all, a) - indexOf(all, b) })
	return given, api.Text{}
}
