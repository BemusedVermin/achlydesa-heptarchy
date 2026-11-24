package actions

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// MaxOptions caps the intents listed for one body. The sample content stays well under it.
const MaxOptions = 500

// candidate is an intent worth evaluating, with how it names itself as an option.
type candidate struct {
	intent Intent
	// withChannels says the action has more than one way to pick channels, so the channels are part
	// of the option's ID.
	withChannels bool
	key          string
	// In the action's choice order.
	choiceValues []string
}

// candidates lists the intents the actor could try, in option order, up to MaxOptions. They are
// well-formed guesses; evaluate decides which are allowed. With requireFree, busy channels are left
// out of the sets.
func (ev *evaluator) candidates(actor ecs.Entity, requireFree bool) ([]candidate, error) {
	var out []candidate
	for _, id := range ev.known(actor) {
		def, ok := ev.w.Content.Action(id)
		if !ok {
			continue
		}
		sets := ev.channelSets(actor, def, requireFree, MaxChannelSets)
		if len(sets) == 0 {
			continue
		}
		targets, err := ev.targets(actor, def)
		if err != nil {
			return nil, err
		}
		for _, values := range ev.choiceCombos(actor, def) {
			choices := map[string]string{}
			for i, name := range def.Choices {
				choices[name] = values[i]
			}
			for _, target := range targets {
				for _, set := range sets {
					in := target
					in.Actor, in.Action, in.Channels = actor, id, set
					if len(choices) > 0 {
						in.Choices = choices
					}
					out = append(out, candidate{
						intent: in, withChannels: len(sets) > 1, key: def.Key, choiceValues: values,
					})
				}
			}
		}
	}
	orderOptions(ev.w, out)
	if len(out) > MaxOptions {
		out = out[:MaxOptions]
	}
	return out, nil
}

// choiceCombos lists every combination of the action's choices, the first choice varying slowest. An
// action with no choices has one empty combination; a choice with no values leaves none.
func (ev *evaluator) choiceCombos(actor ecs.Entity, def content.ActionDef) [][]string {
	combos := [][]string{nil}
	for _, name := range def.Choices {
		values, err := ev.choiceValues(name, actor)
		if err != nil || values == nil {
			return nil
		}
		var next [][]string
		for _, prefix := range combos {
			for _, value := range values {
				next = append(next, append(slices.Clone(prefix), value))
			}
		}
		combos = next
	}
	return combos
}

// targets lists the targets an action could aim at, as intents carrying only the target fields.
func (ev *evaluator) targets(actor ecs.Entity, def content.ActionDef) ([]Intent, error) {
	switch def.Target {
	case content.TargetNone:
		return []Intent{{}}, nil
	case content.TargetSelf:
		return []Intent{{Target: actor}}, nil
	case content.TargetDirection:
		out := make([]Intent, api.DirCount)
		for i := range out {
			dir := api.Dir(i)
			out[i] = Intent{Dir: &dir}
		}
		return out, nil
	case content.TargetBody, content.TargetAlly, content.TargetPart:
		return ev.bodyTargets(actor, def)
	default:
		// Cell targets wait for a cell in the intent.
		return nil, nil
	}
}

func (ev *evaluator) bodyTargets(actor ecs.Entity, def content.ActionDef) ([]Intent, error) {
	key := targetKey{actor: actor, target: def.Target, rng: def.Range, sight: def.Sight}
	if out, ok := ev.targetMemo[key]; ok {
		return out, nil
	}
	out, err := ev.findBodyTargets(actor, def)
	if err != nil {
		return nil, err
	}
	if ev.targetMemo == nil {
		ev.targetMemo = map[targetKey][]Intent{}
	}
	ev.targetMemo[key] = out
	return out, nil
}

func (ev *evaluator) findBodyTargets(actor ecs.Entity, def content.ActionDef) ([]Intent, error) {
	var bodies []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Actor](ev.w.ECS).Query()
	for query.Next() {
		bodies = append(bodies, query.Entity())
	}
	ev.w.SortByUID(bodies)
	var out []Intent
	for _, body := range bodies {
		if body == actor {
			continue
		}
		if reason, err := ev.reach(actor, body, def); err != nil {
			return nil, err
		} else if reason.ID != "" {
			continue
		}
		if def.Target != content.TargetPart {
			out = append(out, Intent{Target: body})
			continue
		}
		for _, c := range timeline.Channels(ev.w, body) {
			if c.Entity != body {
				out = append(out, Intent{Target: body, TargetPart: c.Entity})
			}
		}
	}
	return out, nil
}

// orderOptions sorts candidates into option order: the action's key, then its choices, then the
// target, the part, the direction and the channels, by UID. Every list of intents and options goes
// through it, so they always agree.
func orderOptions(w *sim.World, list []candidate) {
	dirOf := func(in Intent) int {
		if in.Dir == nil {
			return -1
		}
		return int(*in.Dir)
	}
	slices.SortStableFunc(list, func(a, b candidate) int {
		if c := strings.Compare(a.key, b.key); c != 0 {
			return c
		}
		if c := strings.Compare(string(a.intent.Action), string(b.intent.Action)); c != 0 {
			return c
		}
		if c := slices.Compare(a.choiceValues, b.choiceValues); c != 0 {
			return c
		}
		if c := w.CompareUID(a.intent.Target, b.intent.Target); c != 0 {
			return c
		}
		if c := w.CompareUID(a.intent.TargetPart, b.intent.TargetPart); c != 0 {
			return c
		}
		if c := dirOf(a.intent) - dirOf(b.intent); c != 0 {
			return c
		}
		return slices.Compare(timeline.UIDs(w, a.intent.Channels), timeline.UIDs(w, b.intent.Channels))
	})
}

// Startable lists every intent the actor may start right now, in option order.
func Startable(w *sim.World, actor ecs.Entity) []Intent {
	ev := newEvaluator(w)
	cands, err := ev.candidates(actor, true)
	if err != nil {
		return nil
	}
	var out []Intent
	for _, cand := range cands {
		if p, reason, err := ev.evaluate(cand.intent, true); err == nil && p != nil && reason.ID == "" {
			out = append(out, p.intent())
		}
	}
	return out
}

// worthWaking reports whether starting the intent is reason enough to hand its body control. An
// action aimed at another body counts only against an enemy, so a free arm beside a friend does
// not stop a body that is still recovering from a step.
func (ev *evaluator) worthWaking(in Intent) bool {
	def, ok := ev.w.Content.Action(in.Action)
	if !ok {
		return false
	}
	switch def.Target {
	case content.TargetAlly:
		return false
	case content.TargetBody, content.TargetPart:
		return effects.Hostile(ev.w, in.Actor, in.Target)
	default:
		return true
	}
}

// usableChannels is the timeline's usable hook: a channel is usable when an action worth stopping
// for, such as one at an enemy in reach, could start with it were it free.
func usableChannels(w *sim.World, body ecs.Entity, channels []timeline.Channel) []bool {
	flags := make([]bool, len(channels))
	if len(channels) == 0 {
		return flags
	}
	ev := newEvaluator(w)
	cands, err := ev.candidates(body, false)
	if err != nil {
		return flags
	}
	index := make(map[ecs.Entity]int, len(channels))
	for i, c := range channels {
		index[c.Entity] = i
	}
	open := len(channels)
	for _, cand := range cands {
		if open == 0 {
			break
		}
		wanted := false
		for _, used := range cand.intent.Channels {
			if i, ok := index[used.Entity]; ok && !flags[i] {
				wanted = true
			}
		}
		if !wanted {
			continue
		}
		if !ev.worthWaking(cand.intent) {
			continue
		}
		p, reason, err := ev.evaluate(cand.intent, false)
		if err != nil || p == nil || reason.ID != "" {
			continue
		}
		for _, used := range cand.intent.Channels {
			if i, ok := index[used.Entity]; ok && !flags[i] {
				flags[i] = true
				open--
			}
		}
	}
	return flags
}

// optionID is the option's ID: the action's key, each choice, then the target, the part or direction,
// and the channels when there is more than one way to pick them. path names each segment.
func (ev *evaluator) optionID(cand candidate, def content.ActionDef) (id string, path []string) {
	w := ev.w
	in := cand.intent
	parts := []string{cand.key}
	path = []string{w.Content.Render(api.Text{ID: def.Name})}
	for _, value := range cand.choiceValues {
		parts, path = append(parts, value), append(path, ev.valueName(value))
	}
	if in.Dir != nil {
		parts = append(parts, dirNames[*in.Dir])
		path = append(path, w.Content.Render(api.Text{ID: api.ContentID("text.actions.dir." + dirNames[*in.Dir])}))
	}
	entity := func(e ecs.Entity) {
		parts, path = append(parts, uidSegment(w.UIDOf(e))), append(path, ev.segmentName(e))
	}
	if in.Target != in.Actor && !in.Target.IsZero() {
		entity(in.Target)
	}
	if !in.TargetPart.IsZero() {
		entity(in.TargetPart)
	}
	if cand.withChannels {
		for _, channel := range in.Channels {
			entity(channel.Entity)
		}
	}
	return strings.Join(parts, "/"), path
}

// segmentName names a body, or a part by its part template, since parts have no template name.
func (ev *evaluator) segmentName(e ecs.Entity) string {
	if parts := ecs.NewMap[comp.Part](ev.w.ECS); parts.Has(e) {
		if def, ok := ev.w.Content.Part(parts.Get(e).Template); ok {
			return ev.w.Content.Render(api.Text{ID: def.Name})
		}
	}
	return ev.nameOf(e)
}

func uidSegment(uid api.UID) string { return strconv.FormatUint(uint64(uid), 10) }

// label names the action with its choices and target, and what it costs in time.
func (ev *evaluator) label(cand candidate, p *plan) string {
	w := ev.w
	in := cand.intent
	var detail []string
	if in.Dir != nil {
		detail = append(detail, w.Content.Render(api.Text{ID: api.ContentID("text.actions.dir." + dirNames[*in.Dir])}))
	}
	for _, value := range cand.choiceValues {
		detail = append(detail, ev.valueName(value))
	}
	if in.Target != in.Actor && !in.Target.IsZero() {
		detail = append(detail, ev.nameOf(in.Target))
	}
	args := map[string]string{
		"name": w.Content.Render(api.Text{ID: p.def.Name}), "detail": strings.Join(detail, ", "),
		"lands": strconv.FormatInt(p.windup, 10), "busy": strconv.FormatInt(p.windup+p.recovery, 10),
	}
	id := api.ContentID("text.actions.option")
	if len(detail) > 0 {
		id = "text.actions.option_with"
	}
	return w.Content.Render(api.Text{ID: id, Args: args})
}

// valueName is how a chosen value reads in a label: the record's name when the value names a record
// with one, else the value as it is.
func (ev *evaluator) valueName(value string) string {
	if name, ok := ev.w.Content.RecordName(api.ContentID(value)); ok {
		return ev.w.Content.Render(api.Text{ID: name})
	}
	return value
}

// command builds the Do that starts the intent.
func (ev *evaluator) command(in Intent) api.Do {
	w := ev.w
	cmd := api.Do{
		Source: w.UIDOf(in.Actor), Action: in.Action, Channels: timeline.UIDs(w, in.Channels),
		Target: w.UIDOf(in.Target), TargetPart: w.UIDOf(in.TargetPart),
	}
	if in.Dir != nil {
		dir := *in.Dir
		cmd.Dir = &dir
	}
	if len(in.Choices) > 0 {
		cmd.Choices = maps.Clone(in.Choices)
	}
	return cmd
}

// buildOptions adds one option per startable intent of the player body that is deciding now. An
// intent an effect refuses comes out disabled, with the effect's reason.
func buildOptions(w *sim.World, v *api.View) error {
	body, ok := timeline.Current(w)
	if !ok {
		return nil
	}
	if _, player := timeline.CurrentPlayer(w, w.UIDOf(body)); !player {
		return nil
	}
	ev := newEvaluator(w)
	cands, err := ev.candidates(body, true)
	if err != nil {
		return err
	}
	for _, cand := range cands {
		p, reason, err := ev.evaluate(cand.intent, true)
		if err != nil {
			return err
		}
		if p == nil {
			continue
		}
		id, path := ev.optionID(cand, p.def)
		option := api.Option{
			ID: id, Path: path, Label: ev.label(cand, p), Command: ev.command(p.intent()), Enabled: reason.ID == "",
		}
		if !option.Enabled {
			option.Reason = w.Content.Render(reason)
		}
		v.Options = append(v.Options, option)
	}
	return nil
}

// Windup is the ticks the intent would take to land, or 0 when it cannot be worked out. A chooser
// scores by it.
func Windup(w *sim.World, in Intent) int64 {
	p, _, err := newEvaluator(w).evaluate(in, false)
	if err != nil || p == nil {
		return 0
	}
	return p.windup
}
