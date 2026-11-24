package explore

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

func installControl(w *sim.World) error {
	if err := timeline.RegisterController(w, followController, follower{}); err != nil {
		return err
	}
	if err := sim.Handle(w, handleControl); err != nil {
		return err
	}
	if err := sim.Handle(w, handleLead); err != nil {
		return err
	}
	err := w.Systems.Register(
		sim.System{
			Name:    "explore.party_control",
			Cadence: sim.PerEntry,
			After:   []string{"combat.alert"},
			Run:     partyControl,
		},
	)
	if err != nil {
		return err
	}
	return w.Views.Register(sim.ViewBuilder{
		Name: "explore.control", After: []string{"timeline", "actions"}, Build: buildControlOptions,
	})
}

// commandedParty returns the party the player commands when it has the UID.
func commandedParty(w *sim.World, uid api.UID) (ecs.Entity, bool) {
	party, ok := anatomy.ControlledParty(w)
	if !ok || uid == 0 || w.UIDOf(party) != uid {
		return ecs.Entity{}, false
	}
	return party, true
}

// setControl changes the party's control mode and who chose it, publishing ControlChanged when the
// mode itself changes.
func setControl(w *sim.World, party ecs.Entity, mode, by string) {
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	changed := state.Control != mode
	state.Control, state.ControlBy = mode, by
	if changed {
		w.Bus.Publish(api.ControlChanged{Party: w.UIDOf(party), Mode: mode, By: by})
	}
}

// handleControl is the explore.control command: the player picks the mode, which an alert's end
// does not override.
func handleControl(w *sim.World, cmd api.Control) (int64, error) {
	party, ok := commandedParty(w, cmd.Party)
	switch {
	case !ok:
		w.Bus.Publish(api.Reject(cmd.CommandKind(), "text.explore.bad_party"))
		return 0, nil
	case cmd.Mode != api.ControlFollow && cmd.Mode != api.ControlIndividual:
		w.Bus.Publish(api.Reject(cmd.CommandKind(), "text.explore.bad_mode"))
		return 0, nil
	}
	setControl(w, party, cmd.Mode, api.ControlByPlayer)
	applyControllers(w, party)
	return 0, nil
}

// handleLead is the explore.lead command: the member moves to the front of the party order.
func handleLead(w *sim.World, cmd api.Lead) (int64, error) {
	party, ok := commandedParty(w, cmd.Party)
	if !ok {
		w.Bus.Publish(api.Reject(cmd.CommandKind(), "text.explore.bad_party"))
		return 0, nil
	}
	member, found := w.ByUID(cmd.Member)
	if !found || !slices.Contains(membersOf(w, party), member) {
		w.Bus.Publish(api.Reject(cmd.CommandKind(), "text.explore.not_member"))
		return 0, nil
	}
	if anatomy.OutOfFight(w, member) {
		w.Bus.Publish(api.Reject(cmd.CommandKind(), "text.explore.out_of_fight"))
		return 0, nil
	}
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	if state.Leader == cmd.Member {
		w.Bus.Publish(api.Reject(cmd.CommandKind(), "text.explore.already_leads"))
		return 0, nil
	}
	anatomy.SetLeader(w, party, cmd.Member)
	applyControllers(w, party)
	return 0, nil
}

// partyControl is the explore.party_control system. When the alert changed since it last looked, it
// switches the mode, then it sets every member's controller from the mode.
func partyControl(w *sim.World) error {
	var parties []ecs.Entity
	query := ecs.NewFilter1[comp.Party](w.ECS).Query()
	for query.Next() {
		parties = append(parties, query.Entity())
	}
	w.SortByUID(parties)
	for _, party := range parties {
		state := ecs.NewMap[comp.Party](w.ECS).Get(party)
		if state.Alert != state.AlertSeen {
			state.AlertSeen = state.Alert
			if w.Content.Party().IndividualOnAlert {
				switch {
				case state.Alert && state.Control == api.ControlFollow:
					setControl(w, party, api.ControlIndividual, api.ControlByAlert)
				case !state.Alert && state.ControlBy == api.ControlByAlert:
					setControl(w, party, api.ControlFollow, api.ControlByAlert)
				}
			}
		}
		applyControllers(w, party)
	}
	return nil
}

// applyControllers sets the controller of every member from the party's mode: the leader is the
// player's in follow mode and the others follow; in individual mode every member is the player's.
// A member the player was deciding for that is no longer the player's goes back on the timeline to
// be decided for.
func applyControllers(w *sim.World, party ecs.Entity) {
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	actors := ecs.NewMap[comp.Actor](w.ECS)
	current, hasCurrent := timeline.Current(w)
	for _, member := range membersOf(w, party) {
		if !actors.Has(member) {
			continue
		}
		want := timeline.PlayerController
		if state.Control == api.ControlFollow && w.UIDOf(member) != state.Leader {
			want = followController
		}
		actor := actors.Get(member)
		if actor.Controller == want {
			continue
		}
		wasPlayer := actor.Controller == timeline.PlayerController
		actor.Controller = want
		if wasPlayer && hasCurrent && member == current {
			timeline.Reschedule(w, member)
		}
	}
}

// playerDecides reports whether the body deciding now is one of members and the player controls it.
func playerDecides(w *sim.World, members []ecs.Entity) bool {
	current, found := timeline.Current(w)
	if !found || !slices.Contains(members, current) {
		return false
	}
	_, player := timeline.CurrentPlayer(w, w.UIDOf(current))
	return player
}

// buildControlOptions adds the options to switch control mode and to change the leader, while a
// player-controlled member of the party is deciding.
func buildControlOptions(w *sim.World, v *api.View) error {
	party, members, _, ok := focusOf(w)
	if !ok || !playerDecides(w, members) {
		return nil
	}
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	partyUID := w.UIDOf(party)
	if state.Control == api.ControlFollow {
		v.Options = append(v.Options, api.Option{
			ID:      api.ControlOption(api.ControlIndividual),
			Label:   w.Content.Render(api.Text{ID: "text.explore.control_individual"}),
			Command: api.Control{Party: partyUID, Mode: api.ControlIndividual},
			Enabled: true,
		})
	} else {
		v.Options = append(v.Options, api.Option{
			ID:      api.ControlOption(api.ControlFollow),
			Label:   w.Content.Render(api.Text{ID: "text.explore.control_follow"}),
			Command: api.Control{Party: partyUID, Mode: api.ControlFollow},
			Enabled: true,
		})
	}
	for _, member := range members {
		uid := w.UIDOf(member)
		if uid == state.Leader || anatomy.OutOfFight(w, member) {
			continue
		}
		v.Options = append(v.Options, api.Option{
			ID: api.LeadOption(uid),
			Label: w.Content.Render(api.Text{
				ID:   "text.explore.lead_with",
				Args: map[string]string{"name": w.Content.Render(templateName(w, member))},
			}),
			Command: api.Lead{Party: partyUID, Member: uid},
			Enabled: true,
		})
	}
	return nil
}
