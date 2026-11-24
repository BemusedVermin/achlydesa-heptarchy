package actions

import (
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// StepRules reads the party config's limits on how far a body can climb or drop in one step.
func StepRules(w *sim.World) hexmap.StepRules {
	cfg := w.Content.Party()
	return hexmap.StepRules{MaxStepUp: cfg.MaxStepUp, MaxStepDown: cfg.MaxStepDown, MaxDrop: cfg.MaxDrop}
}

// landKind marks the pending actions this package resolves.
const landKind = "actions.land"

// Intent is one body's wish to start one action, with every choice made.
type Intent struct {
	Actor      ecs.Entity
	Action     api.ContentID
	Channels   []timeline.Channel // chosen; empty lets Validate pick the lowest-UID matches
	Target     ecs.Entity
	TargetPart ecs.Entity
	Dir        *api.Dir
	Choices    map[string]string
}

// Modifier is what a hook adds to an intent: wind-up and recovery ticks, and extra effects.
type Modifier struct {
	Windup   int64
	Recovery int64
	OnStart  []content.EffectRef
	OnLand   []content.EffectRef
}

// Validator is an extra check on an intent: ok, or why not.
type Validator func(w *sim.World, in Intent) (ok bool, reason api.Text)

// ModifierFunc is a hook that says what an intent's choices add to it.
type ModifierFunc func(w *sim.World, in Intent) Modifier

// why is the reason text for one of this package's own refusals.
func why(id string) api.Text { return api.Text{ID: api.ContentID("text.actions." + id)} }

// plan is an intent that has been worked out: the content, the channels it takes, and what it costs.
type plan struct {
	id       api.ContentID
	def      content.ActionDef
	actor    ecs.Entity
	channels []timeline.Channel
	target   ecs.Entity
	part     ecs.Entity
	dir      *api.Dir
	choices  map[string]string
	env      content.ActionEnv
	windup   int64
	recovery int64
	onStart  []content.EffectRef
	onLand   []content.EffectRef
	// refusal is why a chosen record's requires formula says the plan may not start, or empty.
	refusal api.Text
}

func (p *plan) intent() Intent {
	return Intent{
		Actor: p.actor, Action: p.id, Channels: p.channels, Target: p.target, TargetPart: p.part,
		Dir: p.dir, Choices: p.choices,
	}
}

func (p *plan) context(w *sim.World) *effects.Context {
	return &effects.Context{
		World: w, Source: p.actor, Used: p.channels, Target: p.target, TargetPart: p.part, Dir: p.dir,
		Choices: p.choices,
	}
}

// evaluator works out intents. It caches what one operation asks again and again, such as who sees
// whom, what targets an action has and which values a choice offers, so it must not outlive the
// operation: the world may change between operations.
type evaluator struct {
	w          *sim.World
	st         *state
	identities *ecs.Map[comp.Identity]
	positions  *ecs.Map[comp.Pos]
	actors     *ecs.Map[comp.Actor]
	sight      map[ecs.Entity]map[ecs.Entity]bool
	inventory  map[ecs.Entity]map[string]int // each actor's whole inventory, for choice offers
	targetMemo map[targetKey][]Intent        // body targets, by what decides them
	valueMemo  map[valueKey][]string         // offered choice values, by choice and actor
}

func newEvaluator(w *sim.World) *evaluator {
	return &evaluator{
		w: w, st: stateOf(w),
		identities: ecs.NewMap[comp.Identity](w.ECS),
		positions:  ecs.NewMap[comp.Pos](w.ECS),
		actors:     ecs.NewMap[comp.Actor](w.ECS),
		sight:      map[ecs.Entity]map[ecs.Entity]bool{},
	}
}

// known lists the actions the body knows, from its template.
func (ev *evaluator) known(actor ecs.Entity) []api.ContentID {
	if !ev.w.Alive(actor) || !ev.identities.Has(actor) {
		return nil
	}
	template, ok := ev.w.Content.Template(ev.identities.Get(actor).Template)
	if !ok {
		return nil
	}
	return template.Actions
}

// sees reports whether source sees target, by the same sight the selectors use.
func (ev *evaluator) sees(source, target ecs.Entity) (bool, error) {
	seen, ok := ev.sight[source]
	if !ok {
		selector, err := effects.ParseBodies("others")
		if err != nil {
			return false, err
		}
		others, err := selector.Resolve(&effects.Context{World: ev.w, Source: source})
		if err != nil {
			return false, err
		}
		seen = make(map[ecs.Entity]bool, len(others))
		for _, body := range others {
			seen[body] = true
		}
		ev.sight[source] = seen
	}
	return seen[target], nil
}

// distance is the hex distance between two bodies on the same map, or 0 when it has no meaning.
func (ev *evaluator) distance(a, b ecs.Entity) int {
	if !ev.w.Alive(a) || !ev.w.Alive(b) || !ev.positions.Has(a) || !ev.positions.Has(b) {
		return 0
	}
	pa, pb := ev.positions.Get(a), ev.positions.Get(b)
	if pa.Map != pb.Map {
		return 0
	}
	return hexmap.Distance(pa.At.Hex, pb.At.Hex)
}

// reach says whether the actor can reach the target body with the action: on the same map, within
// range, and in sight when the action needs it. The empty Text means yes.
func (ev *evaluator) reach(actor, target ecs.Entity, def content.ActionDef) (api.Text, error) {
	w := ev.w
	if !w.Alive(actor) || !w.Alive(target) || !ev.positions.Has(actor) || !ev.positions.Has(target) ||
		ev.positions.Get(actor).Map != ev.positions.Get(target).Map {
		return why("bad_target"), nil
	}
	if ev.distance(actor, target) > def.Range {
		return why("out_of_range"), nil
	}
	if def.Sight {
		seen, err := ev.sees(actor, target)
		if err != nil {
			return api.Text{}, err
		}
		if !seen {
			return why("not_visible"), nil
		}
	}
	return api.Text{}, nil
}

// hasPart reports whether part is one of target's own parts: a channel that is not the body.
func (ev *evaluator) hasPart(target, part ecs.Entity) bool {
	if !ev.w.Alive(part) || part == target {
		return false
	}
	for _, c := range timeline.Channels(ev.w, target) {
		if c.Entity == part {
			return true
		}
	}
	return false
}

// checkTarget checks the intent's target, part and direction against the action's target kind.
func (ev *evaluator) checkTarget(in Intent, def content.ActionDef) (p plan, reason api.Text, err error) {
	bad := func() (plan, api.Text, error) { return plan{}, why("bad_target"), nil }
	switch def.Target {
	case content.TargetNone:
		if !in.Target.IsZero() || !in.TargetPart.IsZero() || in.Dir != nil {
			return bad()
		}
	case content.TargetSelf:
		if (!in.Target.IsZero() && in.Target != in.Actor) || !in.TargetPart.IsZero() || in.Dir != nil {
			return bad()
		}
		p.target = in.Actor
	case content.TargetDirection:
		if in.Dir == nil || !in.Dir.Valid() || !in.Target.IsZero() || !in.TargetPart.IsZero() {
			return bad()
		}
		dir := *in.Dir
		p.dir = &dir
	case content.TargetBody, content.TargetAlly, content.TargetPart:
		if in.Target == in.Actor || in.Dir != nil || !ev.isBody(in.Target) {
			return bad()
		}
		if def.Target == content.TargetPart {
			if !ev.hasPart(in.Target, in.TargetPart) {
				return bad()
			}
			p.part = in.TargetPart
		} else if !in.TargetPart.IsZero() {
			return bad()
		}
		if def.Target == content.TargetAlly && effects.Hostile(ev.w, in.Actor, in.Target) {
			return bad()
		}
		if reason, err := ev.reach(in.Actor, in.Target, def); err != nil || reason.ID != "" {
			return plan{}, reason, err
		}
		p.target = in.Target
	default:
		return bad() // "cell" targets need a cell in the intent, which nothing carries yet
	}
	return p, api.Text{}, nil
}

// isBody reports whether e is a body: a living entity on the timeline.
func (ev *evaluator) isBody(e ecs.Entity) bool {
	return ev.w.Alive(e) && ev.actors.Has(e)
}

// checkChoices checks that the intent made exactly the choices the action asks for, each one offered.
func (ev *evaluator) checkChoices(in Intent, def content.ActionDef) api.Text {
	if len(in.Choices) != len(def.Choices) {
		return why("bad_choice")
	}
	for _, name := range def.Choices {
		value, ok := in.Choices[name]
		if !ok {
			return why("bad_choice")
		}
		values, err := ev.choiceValues(name, in.Actor)
		if err != nil || !slices.Contains(values, value) {
			return why("bad_choice")
		}
	}
	return api.Text{}
}

// env builds the formula environment for the plan.
func (ev *evaluator) env(p *plan) (content.ActionEnv, error) {
	stepTicks, err := timeline.StepTicks(ev.w, p.actor)
	if err != nil {
		return content.ActionEnv{}, err
	}
	choices := p.choices
	if choices == nil {
		choices = map[string]string{}
	}
	source := effects.BodyEnvOf(ev.w, p.actor, timeline.FirstPart(p.channels))
	return content.ActionEnv{
		Speed: timeline.Speed(ev.w, p.actor), StepTicks: stepTicks,
		Distance: ev.distance(p.actor, p.target), Used: len(p.channels),
		Source: source, Target: effects.BodyEnvOf(ev.w, p.target, p.part),
		Chosen: effects.ChosenRecords(ev.w, choices), Inventory: inventory.Usable(ev.w, p.actor, p.target),
	}, nil
}

// modify folds the mods of the chosen records, in the order of the action's choices, and then the
// modifier hooks, in registration order, into the plan.
func (ev *evaluator) modify(p *plan) error {
	p.onStart, p.onLand = slices.Clone(p.def.OnStart), slices.Clone(p.def.OnLand)
	reg := ev.w.Content
	for _, name := range p.def.Choices {
		id := api.ContentID(p.choices[name])
		mods, ok := reg.Mods(id)
		if !ok {
			continue
		}
		windup, recovery, err := reg.ModsTicks(id, p.env)
		if err != nil {
			return err
		}
		p.windup += windup
		p.recovery += recovery
		p.onStart = append(p.onStart, mods.OnStart...)
		p.onLand = append(p.onLand, mods.OnLand...)
		holds, refusal, err := reg.ModsRequires(id, p.env)
		if err != nil {
			return err
		}
		if !holds && p.refusal.ID == "" {
			p.refusal = api.Text{ID: refusal}
		}
	}
	for _, hook := range ev.st.modifiers {
		m := hook(ev.w, p.intent())
		p.windup += m.Windup
		p.recovery += m.Recovery
		p.onStart = append(p.onStart, m.OnStart...)
		p.onLand = append(p.onLand, m.OnLand...)
	}
	return nil
}

// evaluate works an intent out. A nil plan means it is structurally invalid, with the reason. A plan
// with a reason is well-formed but not allowed right now (a check failed); its timing is still
// filled in, so the player can see a disabled option's cost. An error is a content bug.
// requireFree says whether the channels must be free; the usable hook asks without it.
func (ev *evaluator) evaluate(in Intent, requireFree bool) (*plan, api.Text, error) {
	w := ev.w
	if !w.Alive(in.Actor) {
		return nil, why("not_known"), nil
	}
	def, ok := w.Content.Action(in.Action)
	if !ok || !slices.Contains(ev.known(in.Actor), in.Action) {
		return nil, why("not_known"), nil
	}
	if reason := ev.checkChoices(in, def); reason.ID != "" {
		return nil, reason, nil
	}
	p, reason, err := ev.checkTarget(in, def)
	if err != nil || reason.ID != "" {
		return nil, reason, err
	}
	p.id, p.def, p.actor, p.choices = in.Action, def, in.Actor, in.Choices
	if p.channels, reason = ev.assign(in, def, requireFree); reason.ID != "" {
		return nil, reason, nil
	}
	if p.env, err = ev.env(&p); err != nil {
		return nil, api.Text{}, err
	}
	if ok, err := w.Content.ActionRequires(in.Action, p.env); err != nil {
		return nil, api.Text{}, err
	} else if !ok {
		return nil, why("requires"), nil
	}
	if p.windup, err = w.Content.ActionWindup(in.Action, p.env); err != nil {
		return nil, api.Text{}, err
	}
	if p.recovery, err = w.Content.ActionRecovery(in.Action, p.env); err != nil {
		return nil, api.Text{}, err
	}
	if err := ev.modify(&p); err != nil {
		return nil, api.Text{}, err
	}
	if p.windup < 0 || p.recovery < 0 {
		return nil, api.Text{}, fmt.Errorf("actions: %s: wind-up %d and recovery %d must not be negative",
			in.Action, p.windup, p.recovery)
	}
	if p.windup+p.recovery < 1 {
		return nil, why("too_quick"), nil
	}
	return &p, ev.verdict(&p), nil
}

// verdict asks the hooks and the effects whether the plan may start now.
func (ev *evaluator) verdict(p *plan) api.Text {
	if p.refusal.ID != "" {
		return p.refusal
	}
	for _, validate := range ev.st.validators {
		if ok, reason := validate(ev.w, p.intent()); !ok {
			return reason
		}
	}
	ctx := p.context(ev.w)
	if ok, reason := effects.Check(ctx, p.onStart); !ok {
		return reason
	}
	if p.windup == 0 {
		if ok, reason := effects.Check(ctx, p.onLand); !ok {
			return reason
		}
	}
	return api.Text{}
}

// Validate says whether the body may start the intent right now, and if not, why. A content bug
// (a formula that fails) counts as a refusal here and is an error from Start.
func Validate(w *sim.World, in Intent) (ok bool, reason api.Text) {
	p, reason := newEvaluator(w).validate(in)
	return p != nil, reason
}

// validate works out the intent for starting now. A nil plan comes with the reason it may not
// start, a content bug included.
func (ev *evaluator) validate(in Intent) (*plan, api.Text) {
	p, reason, err := ev.evaluate(in, true)
	switch {
	case err != nil:
		return nil, api.Text{ID: why("broken").ID, Args: map[string]string{"error": err.Error()}}
	case p == nil || reason.ID != "":
		return nil, reason
	}
	return p, api.Text{}
}

// Start validates the intent, applies the action's start effects, occupies its channels for the
// wind-up plus recovery, and lands it at once or schedules the landing.
func Start(w *sim.World, in Intent) error { return startIntent(w, in, false) }

// StartAutonomous is Start for an action a graft takes on its own: the same checks and wind-up,
// and the ActionStarted event says it was autonomous.
func StartAutonomous(w *sim.World, in Intent) error { return startIntent(w, in, true) }

// startIntent is Start's own check: unlike validate, it reports a content bug as an error.
func startIntent(w *sim.World, in Intent, autonomous bool) error {
	ev := newEvaluator(w)
	p, reason, err := ev.evaluate(in, true)
	switch {
	case err != nil:
		return err
	case p == nil || reason.ID != "":
		return fmt.Errorf("actions: cannot start %s: %s", in.Action, w.Content.Render(reason))
	}
	return ev.start(p, autonomous)
}

func (ev *evaluator) start(p *plan, autonomous bool) error {
	w := ev.w
	if err := effects.Apply(p.context(w), p.onStart); err != nil {
		return fmt.Errorf("actions: %s: %w", p.id, err)
	}
	for _, c := range p.channels {
		if err := timeline.Occupy(w, c, p.windup+p.recovery); err != nil {
			return fmt.Errorf("actions: %s: %w", p.id, err)
		}
	}
	if p.windup == 0 {
		return ev.land(p)
	}
	uids := timeline.UIDs(w, p.channels)
	pending, err := timeline.Schedule(w, comp.Pending{
		Kind: landKind, Actor: w.UIDOf(p.actor), Channel: uids[0], Target: w.UIDOf(p.target), Lands: w.Now + p.windup,
	})
	if err != nil {
		return fmt.Errorf("actions: %s: %w", p.id, err)
	}
	landing := &comp.Landing{Action: p.id, Channels: uids, TargetPart: w.UIDOf(p.part), Choices: p.choices}
	if p.dir != nil {
		landing.HasDir, landing.Dir = true, *p.dir
	}
	ecs.NewMap[comp.Landing](w.ECS).Add(pending, landing)
	w.Bus.Publish(api.ActionStarted{
		Source: w.UIDOf(p.actor), Action: p.id, Target: w.UIDOf(p.target), Tick: w.Now, Lands: w.Now + p.windup,
		Autonomous: autonomous,
	})
	return nil
}

// connects says whether the landing action reaches its target, and why not when it does not.
func (ev *evaluator) connects(p *plan) (api.Text, error) {
	env, err := ev.env(p)
	if err != nil {
		return api.Text{}, err
	}
	connects, present, err := ev.w.Content.ActionConnects(p.id, env)
	switch {
	case err != nil:
		return api.Text{}, err
	case present && !connects:
		return why("no_connect"), nil
	case present:
		return api.Text{}, nil
	}
	switch p.def.Target {
	case content.TargetBody, content.TargetAlly, content.TargetPart:
		if reason, err := ev.reach(p.actor, p.target, p.def); err != nil || reason.ID != "" {
			return reason, err
		}
		if p.def.Target == content.TargetPart && !ev.hasPart(p.target, p.part) {
			return why("bad_target"), nil
		}
	}
	return api.Text{}, nil
}

// land applies the action's landing effects when it connects, and publishes ActionMissed when not.
func (ev *evaluator) land(p *plan) error {
	w := ev.w
	reason, err := ev.connects(p)
	if err != nil {
		return err
	}
	ctx := p.context(w)
	if reason.ID == "" {
		if ok, refusal := effects.Check(ctx, p.onLand); !ok {
			reason = refusal
		}
	}
	if reason.ID != "" {
		ev.missed(p.actor, w.UIDOf(p.actor), w.UIDOf(p.target), p.id, reason)
		return nil
	}
	if err := effects.Apply(ctx, p.onLand); err != nil {
		return fmt.Errorf("actions: %s: %w", p.id, err)
	}
	return nil
}

// missed publishes ActionMissed for an action that did not land.
func (ev *evaluator) missed(actor ecs.Entity, source, target api.UID, action api.ContentID, reason api.Text) {
	ev.w.Bus.Publish(api.ActionMissed{
		Source: source, Action: action, Target: target, Tick: ev.w.Now, Reason: reason,
		Line: api.Text{ID: "text.actions.missed", Args: ev.landArgs(actor, action)},
	})
}

// landArgs are the text arguments naming a landing's actor and action.
func (ev *evaluator) landArgs(actor ecs.Entity, action api.ContentID) map[string]string {
	def, _ := ev.w.Content.Action(action)
	return map[string]string{"actor": ev.nameOf(actor), "action": ev.w.Content.Render(api.Text{ID: def.Name})}
}

// nameOf is the rendered name of a body, from its template.
func (ev *evaluator) nameOf(e ecs.Entity) string {
	return ev.w.Content.Render(ev.w.Content.TemplateName(ev.w.TemplateOf(e)))
}

// landingPlan rebuilds the plan of a pending action. The reason is set when a channel it used is gone.
func (ev *evaluator) landingPlan(pending ecs.Entity) (*plan, api.Text, error) {
	w := ev.w
	pend := ecs.NewMap[comp.Pending](w.ECS).Get(pending)
	landing := ecs.NewMap[comp.Landing](w.ECS).Get(pending)
	def, ok := w.Content.Action(landing.Action)
	actor, actorOK := w.ByUID(pend.Actor)
	if !ok || !actorOK {
		return nil, why("channel_gone"), nil
	}
	p := &plan{id: landing.Action, def: def, actor: actor, choices: landing.Choices}
	p.target, _ = w.ByUID(pend.Target)
	p.part, _ = w.ByUID(landing.TargetPart)
	if landing.HasDir {
		dir := landing.Dir
		p.dir = &dir
	}
	channels := timeline.Channels(w, actor)
	for _, uid := range landing.Channels {
		e, ok := w.ByUID(uid)
		i := slices.IndexFunc(channels, func(c timeline.Channel) bool { return ok && c.Entity == e })
		if i < 0 {
			return nil, why("channel_gone"), nil
		}
		p.channels = append(p.channels, channels[i])
	}
	var err error
	if p.env, err = ev.env(p); err != nil {
		return nil, api.Text{}, err
	}
	if err := ev.modify(p); err != nil {
		return nil, api.Text{}, err
	}
	return p, api.Text{}, nil
}

// resolveLanding lands a pending action, or publishes its miss when a channel it used is gone.
func resolveLanding(w *sim.World, pending ecs.Entity) error {
	ev := newEvaluator(w)
	p, reason, err := ev.landingPlan(pending)
	if err != nil {
		return err
	}
	if p == nil {
		pend := ecs.NewMap[comp.Pending](w.ECS).Get(pending)
		landing := ecs.NewMap[comp.Landing](w.ECS).Get(pending)
		actor, _ := w.ByUID(pend.Actor)
		ev.missed(actor, pend.Actor, pend.Target, landing.Action, reason)
		return nil
	}
	return ev.land(p)
}

// labelLanding describes a pending action for the timeline view.
func labelLanding(w *sim.World, pending ecs.Entity) string {
	ev := newEvaluator(w)
	pend := ecs.NewMap[comp.Pending](w.ECS).Get(pending)
	landing := ecs.NewMap[comp.Landing](w.ECS).Get(pending)
	actor, _ := w.ByUID(pend.Actor)
	args := ev.landArgs(actor, landing.Action)
	id := api.ContentID("text.actions.lands")
	if target, ok := w.ByUID(pend.Target); ok && target != actor {
		id, args["target"] = "text.actions.lands_on", ev.nameOf(target)
	}
	return w.Content.Render(api.Text{ID: id, Args: args})
}

// targetKey is what decides the body targets an action offers: the actor, the kind of target, the
// range and whether the target must be in sight.
type targetKey struct {
	actor  ecs.Entity
	target string
	rng    int
	sight  bool
}

// valueKey names one choice's offered values for one actor.
type valueKey struct {
	actor ecs.Entity
	name  string
}
