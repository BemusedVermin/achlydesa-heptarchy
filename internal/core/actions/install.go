package actions

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// state is the hooks and providers actions keep in the world.
type state struct {
	providers  map[string]ChoiceProvider
	validators []Validator
	modifiers  []modifierFunc
}

func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{providers: map[string]ChoiceProvider{directionChoice: directionValues}}
	res.Add(st)
	return st
}

// AddValidator adds a check to Validate. Combat uses it for immune materials and type conditions.
// Validators run in registration order, after the action's own checks.
func AddValidator(w *sim.World, fn Validator) {
	st := stateOf(w)
	st.validators = append(st.validators, fn)
}

// AddModifier lets an intent's choices add wind-up ticks, recovery ticks and extra start or landing
// effects. Combat uses it so a type changes any form it is loaded into. Modifiers run in
// registration order.
func AddModifier(w *sim.World, fn func(w *sim.World, in Intent) Modifier) {
	st := stateOf(w)
	st.modifiers = append(st.modifiers, fn)
}

// Install registers the events, the actions.do command, the landing resolver, the option builder and
// the usable hook. It needs effects and timeline installed first.
func Install(w *sim.World) error {
	if w.Content == nil {
		return errors.New("actions: install: world has no content")
	}
	stateOf(w)
	if err := api.RegisterEvent[api.ActionStarted](w.Codec); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.ActionMissed](w.Codec); err != nil {
		return err
	}
	if err := sim.Handle(w, handleDo); err != nil {
		return err
	}
	err := timeline.RegisterResolver(w, landKind, timeline.Resolver{Resolve: resolveLanding, Label: labelLanding})
	if err != nil {
		return err
	}
	timeline.SetUsableChannels(w, usableChannels)
	// Options come after the timeline's own, which are the waits.
	err = w.Views.Register(sim.ViewBuilder{Name: "actions", After: []string{"timeline"}, Build: buildOptions})
	if err != nil {
		return err
	}
	return nil
}

// ValidateContent checks what the loader cannot know: that every choice an action asks for is served,
// by a Go provider or a content choice, and that no content choice shadows a provider. It reports
// every problem. game.New calls it after the packages are installed.
func ValidateContent(w *sim.World) error {
	st := stateOf(w)
	var problems []error
	for _, name := range w.Content.ChoiceNames() {
		if _, ok := st.providers[name]; ok {
			problems = append(problems, fmt.Errorf("choices[%q]: shadows the choice provider of the same name", name))
		}
	}
	for _, id := range w.Content.ActionIDs() {
		def, _ := w.Content.Action(id)
		for i, name := range def.Choices {
			_, provided := st.providers[name]
			if _, served := w.Content.Choice(name); !provided && !served {
				problems = append(
					problems,
					fmt.Errorf("actions[%q].choices[%d]: nothing serves choice %q", id, i, name),
				)
			}
		}
	}
	if err := errors.Join(problems...); err != nil {
		return fmt.Errorf("actions: content asks for choices nothing provides: %w", err)
	}
	return nil
}

// handleDo is the actions.do command: the current player body starts one of its actions.
func handleDo(w *sim.World, cmd api.Do) (int64, error) {
	reject := func(reason api.Text) (int64, error) {
		w.Bus.Publish(api.Rejected{Command: cmd.CommandKind(), Reason: reason})
		return 0, nil
	}
	actor, ok := timeline.CurrentPlayer(w, cmd.Source)
	if !ok {
		return reject(api.Text{ID: "text.timeline.not_current"})
	}
	in, ok := intentOf(w, actor, cmd)
	if !ok {
		return reject(why("bad_target"))
	}
	if ok, reason := Validate(w, in); !ok {
		return reject(reason)
	}
	return 0, Start(w, in)
}

// intentOf turns a command's UIDs into entities. Channels are looked up among the actor's own; one
// that is not its own makes the intent unusable.
func intentOf(w *sim.World, actor ecs.Entity, cmd api.Do) (Intent, bool) {
	in := Intent{Actor: actor, Action: cmd.Action, Dir: cmd.Dir, Choices: cmd.Choices}
	lookup := func(uid api.UID) (ecs.Entity, bool) {
		if uid == 0 {
			return ecs.Entity{}, true
		}
		e, ok := w.ByUID(uid)
		return e, ok
	}
	var ok bool
	if in.Target, ok = lookup(cmd.Target); !ok {
		return Intent{}, false
	}
	if in.TargetPart, ok = lookup(cmd.TargetPart); !ok {
		return Intent{}, false
	}
	all := timeline.Channels(w, actor)
	for _, uid := range cmd.Channels {
		e, ok := w.ByUID(uid)
		if !ok {
			return Intent{}, false
		}
		i := indexOf(all, timeline.Channel{Entity: e})
		if i < 0 {
			return Intent{}, false
		}
		in.Channels = append(in.Channels, all[i])
	}
	return in, true
}
