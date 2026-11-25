package agents

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// hireView is the view builder that adds the hire options.
const hireView = "agents.hire"

// Text IDs of the reasons a hire is refused.
const (
	badPartyReason     api.ContentID = "text.agents.bad_party"
	notHirableReason   api.ContentID = "text.agents.not_hirable"
	tooFarReason       api.ContentID = "text.agents.hire_too_far"
	hostileNearReason  api.ContentID = "text.agents.hire_hostile_near"
	partyFullReason    api.ContentID = "text.agents.hire_party_full"
	cannotPayReason    api.ContentID = "text.agents.hire_cannot_pay"
	leaderDownedReason api.ContentID = "text.agents.hire_leader_out"
)

func installRecruits(w *sim.World) error {
	if err := sim.Handle(w, handleHire); err != nil {
		return err
	}
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.Hired],
		api.RegisterEvent[api.Grumbled],
		api.RegisterEvent[api.Deserted],
		api.RegisterEvent[api.Fled],
	} {
		if err := register(w.Codec); err != nil {
			return err
		}
	}
	return w.Views.Register(sim.ViewBuilder{
		Name: hireView, After: []string{"timeline", "actions"}, Build: buildHireOptions,
	})
}

// recruitPrice is what the recruit asks: each want's count times its worth, rounded up, merged
// per item and sorted by item.
func recruitPrice(w *sim.World, recruit ecs.Entity) ([]api.ItemCount, error) {
	template := w.TemplateOf(recruit)
	def, _ := w.Content.Template(template)
	worth, err := w.Content.RecruitWorth(template, effects.BodyEnvOf(w, recruit, ecs.Entity{}))
	if err != nil {
		return nil, err
	}
	totals := map[api.ContentID]int{}
	for _, stack := range content.RecruitPrice(def.Recruit.Wants, worth) {
		totals[stack.Item] += stack.Count
	}
	price := make([]api.ItemCount, 0, len(totals))
	for item, count := range totals {
		price = append(price, api.ItemCount{Item: item, Count: count})
	}
	slices.SortFunc(price, func(a, b api.ItemCount) int { return cmp.Compare(a.Item, b.Item) })
	return price, nil
}

// hirable reports whether the body is a living generic whose template is for hire.
func hirable(w *sim.World, body ecs.Entity) bool {
	if !w.Alive(body) || !ecs.NewMap[comp.Generic](w.ECS).Has(body) || anatomy.OutOfFight(w, body) {
		return false
	}
	def, ok := w.Content.Template(w.TemplateOf(body))
	return ok && def.Recruit != nil
}

// hireRefusal is the text ID of the reason the party cannot hire the recruit now, or "" when it can.
func hireRefusal(w *sim.World, party, recruit ecs.Entity, price []api.ItemCount) api.ContentID {
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	positions := ecs.NewMap[comp.Pos](w.ECS)
	leader, found := w.ByUID(state.Leader)
	if !found || anatomy.OutOfFight(w, leader) || !positions.Has(leader) {
		return leaderDownedReason
	}
	here, there := positions.Get(leader), positions.Get(recruit)
	if here.Map != there.Map || hexmap.CellDistance(here.At, there.At) > 1 {
		return tooFarReason
	}
	if _, isTown := TownOf(w, here.Map); !isTown {
		return notHirableReason
	}
	if state.Alert {
		return hostileNearReason
	}
	if membersAlive(w, party) >= w.Content.Party().Slots {
		return partyFullReason
	}
	for _, want := range price {
		if inventory.Count(w, party, want.Item, 0) < want.Count {
			return cannotPayReason
		}
	}
	return ""
}

func membersAlive(w *sim.World, party ecs.Entity) int {
	n := 0
	for _, member := range anatomy.Members(w, party) {
		if !ecs.NewMap[comp.Dead](w.ECS).Has(member) {
			n++
		}
	}
	return n
}

// handleHire is the agents.hire command.
func handleHire(w *sim.World, cmd api.Hire) (int64, error) {
	party, ok := anatomy.ControlledParty(w)
	if !ok || cmd.Party == 0 || w.UIDOf(party) != cmd.Party {
		w.Bus.Publish(api.Reject(cmd.CommandKind(), badPartyReason))
		return 0, nil
	}
	recruit, found := w.ByUID(cmd.Recruit)
	if !found || !hirable(w, recruit) {
		w.Bus.Publish(api.Reject(cmd.CommandKind(), notHirableReason))
		return 0, nil
	}
	price, err := recruitPrice(w, recruit)
	if err != nil {
		return 0, err
	}
	if reason := hireRefusal(w, party, recruit, price); reason != "" {
		w.Bus.Publish(api.Reject(cmd.CommandKind(), reason))
		return 0, nil
	}
	for _, want := range price {
		inventory.Take(w, party, want.Item, 0, want.Count)
	}
	join(w, party, recruit)
	w.Bus.Publish(api.Hired{Party: cmd.Party, Recruit: cmd.Recruit, Paid: price})
	return 0, nil
}

// join makes the recruit the last member of the party. It stops being a generic, so the town's
// replacement timer starts. The explore package gives it the follow controller on the next timeline entry.
func join(w *sim.World, party, recruit ecs.Entity) {
	ecs.NewMap[comp.Generic](w.ECS).Remove(recruit)
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	if leader, ok := w.ByUID(state.Leader); ok {
		if allegiances := ecs.NewMap[comp.Allegiance](w.ECS); allegiances.Has(leader) {
			sim.Put(w, recruit, *allegiances.Get(leader))
		}
	}
	ecs.NewMap[comp.MemberOf](w.ECS).Add(recruit, &comp.MemberOf{}, party)
	state.Order = append(state.Order, w.UIDOf(recruit))
}

// buildHireOptions offers a hire option for each hirable generic next to the party's leader, while
// a player-controlled member of the party is deciding.
func buildHireOptions(w *sim.World, v *api.View) error {
	party, ok := anatomy.ControlledParty(w)
	if !ok {
		return nil
	}
	current, found := timeline.Current(w)
	if !found || !anatomy.InControlledParty(w, current) {
		return nil
	}
	if _, player := timeline.CurrentPlayer(w, w.UIDOf(current)); !player {
		return nil
	}
	leader, found := w.ByUID(ecs.NewMap[comp.Party](w.ECS).Get(party).Leader)
	if !found || !ecs.NewMap[comp.Pos](w.ECS).Has(leader) {
		return nil
	}
	here := *ecs.NewMap[comp.Pos](w.ECS).Get(leader)
	if _, isTown := TownOf(w, here.Map); !isTown {
		return nil
	}
	var candidates []ecs.Entity
	query := ecs.NewFilter2[comp.Generic, comp.Pos](w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		if pos.Map == here.Map && hexmap.CellDistance(pos.At, here.At) <= 1 && hirable(w, query.Entity()) {
			candidates = append(candidates, query.Entity())
		}
	}
	for _, recruit := range w.SortByUID(candidates) {
		price, err := recruitPrice(w, recruit)
		if err != nil {
			return err
		}
		option := api.Option{
			ID: api.HireOption(w.UIDOf(recruit)),
			Label: w.Content.Render(api.Text{ID: "text.agents.hire", Args: map[string]string{
				"name":  w.Content.Render(w.Content.TemplateName(w.TemplateOf(recruit))),
				"price": priceText(w, price),
			}}),
			Command: api.Hire{Party: w.UIDOf(party), Recruit: w.UIDOf(recruit)},
			Enabled: true,
		}
		if reason := hireRefusal(w, party, recruit, price); reason != "" {
			option.Enabled = false
			option.Reason = w.Content.Render(api.Text{ID: reason})
		}
		v.Options = append(v.Options, option)
	}
	return nil
}

func priceText(w *sim.World, price []api.ItemCount) string {
	parts := make([]string, 0, len(price))
	for _, want := range price {
		item, _ := w.Content.Item(want.Item)
		parts = append(parts, w.Content.Render(api.Text{ID: "text.agents.price_part", Args: map[string]string{
			"count": strconv.Itoa(want.Count),
			"item":  w.Content.Render(api.Text{ID: item.Name}),
		}}))
	}
	return strings.Join(parts, w.Content.Render(api.Text{ID: "text.agents.price_separator"}))
}
