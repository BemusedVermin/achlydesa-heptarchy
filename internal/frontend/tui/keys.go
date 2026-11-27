package tui

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// binding ties keys to an option ID. Keys never build commands: a key does something only when
// the view offers its option.
type binding struct {
	keys   key.Binding
	option string
}

// KeyMap is every binding of the map screen. A plain key.Binding is one whose option depends on
// the view, or that drives the screen itself.
type KeyMap struct {
	moves []binding
	// up and down step NE or NW, and SE or SW, alternating, so the party climbs a screen column.
	up        key.Binding
	down      key.Binding
	waitEvent binding
	waitPhase binding
	// nextLead takes the next leader option, cycling.
	nextLead key.Binding
	// control takes whichever control option is offered.
	control   key.Binding
	prevTick  key.Binding
	nextTick  key.Binding
	selectTic key.Binding
	menu      key.Binding
	look      key.Binding
	quit      key.Binding
	advance   key.Binding
	skip      key.Binding
}

// DefaultKeyMap binds q e a d z c to the six steps, w and s to the zigzag, and . and , to waiting.
func DefaultKeyMap() KeyMap {
	move := func(k, option string) binding {
		return binding{key.NewBinding(key.WithKeys(k), key.WithHelp(k, "move")), option}
	}
	return KeyMap{
		moves: []binding{
			move("q", api.StepOption("nw")), move("e", api.StepOption("ne")), move("a", api.StepOption("w")),
			move("d", api.StepOption("e")), move("z", api.StepOption("sw")), move("c", api.StepOption("se")),
		},
		up:        key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "up")),
		down:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "down")),
		waitEvent: binding{key.NewBinding(key.WithKeys("."), key.WithHelp(".", "wait")), api.OptionWaitEvent},
		waitPhase: binding{
			key.NewBinding(key.WithKeys(","), key.WithHelp(",", "wait to next phase")),
			api.OptionWaitPhase,
		},
		nextLead:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next leader")),
		control:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "lead or follow")),
		prevTick:  key.NewBinding(key.WithKeys("["), key.WithHelp("[ ]", "pick tick")),
		nextTick:  key.NewBinding(key.WithKeys("]")),
		selectTic: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "list tick")),
		menu:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "actions")),
		look:      key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "look")),
		advance:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("any key", "continue")),
		skip:      key.NewBinding(key.WithKeys("x", "esc"), key.WithHelp("x", "skip")),
		quit:      key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit")),
	}
}

// optionByID is the view's option with the ID, if the view offers it.
func optionByID(v api.View, id string) (api.Option, bool) {
	i := slices.IndexFunc(v.Options, func(o api.Option) bool { return o.ID == id })
	if i < 0 {
		return api.Option{}, false
	}
	return v.Options[i], true
}

// hasOption reports whether the view offers an option.
func hasOption(v api.View, id string) bool {
	_, ok := optionByID(v, id)
	return ok
}

// OptionFor is the option ID a key press stands for in this view, if any. It does not check the
// option is offered, except for the bindings that choose among several.
func (k KeyMap) OptionFor(msg tea.KeyPressMsg, v api.View) (string, bool) {
	all := append(slices.Clone(k.moves), k.waitEvent, k.waitPhase)
	for _, b := range all {
		if key.Matches(msg, b.keys) {
			return b.option, true
		}
	}
	switch {
	case key.Matches(msg, k.control):
		for _, id := range []string{api.ControlOption(api.ControlIndividual), api.ControlOption(api.ControlFollow)} {
			if hasOption(v, id) {
				return id, true
			}
		}
	case key.Matches(msg, k.nextLead):
		return nextLeadOption(v)
	}
	return "", false
}

// VerticalOptions are the step options a vertical key tries, in order. A pointy-top grid has no
// straight up or down, so the party alternates two diagonals. The first press after another move
// (last empty) prefers the east diagonal; otherwise the diagonal that was not just used goes first.
func VerticalOptions(up bool, last string) [2]string {
	east, west := api.StepOption("se"), api.StepOption("sw")
	if up {
		east, west = api.StepOption("ne"), api.StepOption("nw")
	}
	if last == east {
		return [2]string{west, east}
	}
	return [2]string{east, west}
}

// VerticalOption picks the step a vertical key stands for: the first of [VerticalOptions] that is
// offered and enabled, else the first that is offered (so its reason shows), else none.
func VerticalOption(v api.View, up bool, last string) (string, bool) {
	offered := ""
	for _, id := range VerticalOptions(up, last) {
		o, ok := optionByID(v, id)
		switch {
		case !ok:
		case o.Enabled:
			return id, true
		case offered == "":
			offered = id
		}
	}
	return offered, offered != ""
}

// nextLeadOption picks the member after the leader in a fixed (UID) order, wrapping, so that
// repeated presses visit every member even though leading reorders the party.
func nextLeadOption(v api.View) (string, bool) {
	var uids []api.UID
	for _, o := range v.Options {
		if uid, ok := api.ParseLeadOption(o.ID); ok {
			uids = append(uids, uid)
		}
	}
	if len(uids) == 0 {
		return "", false
	}
	slices.Sort(uids)
	var leader api.UID
	if v.Party != nil {
		for _, m := range v.Party.Members {
			if m.Leader {
				leader = m.UID
			}
		}
	}
	for _, uid := range uids {
		if uid > leader {
			return api.LeadOption(uid), true
		}
	}
	return api.LeadOption(uids[0]), true
}

// HelpBindings are the bindings to list on the help line: those whose options are present.
func (k KeyMap) HelpBindings(v api.View) []key.Binding {
	var out []key.Binding
	if slices.ContainsFunc(k.moves, func(b binding) bool { return hasOption(v, b.option) }) {
		out = append(out, key.NewBinding(key.WithHelp("q e a d z c", "move")))
	}
	if hasOption(v, api.StepOption("ne")) || hasOption(v, api.StepOption("nw")) || hasOption(v, api.StepOption("se")) ||
		hasOption(v, api.StepOption("sw")) {
		out = append(out, key.NewBinding(key.WithHelp("w s", "up down")))
	}
	if hasOption(v, k.waitEvent.option) {
		out = append(out, k.waitEvent.keys)
	}
	if hasOption(v, k.waitPhase.option) {
		out = append(out, k.waitPhase.keys)
	}
	if _, ok := nextLeadOption(v); ok {
		out = append(out, k.nextLead)
	}
	if hasOption(v, api.ControlOption(api.ControlIndividual)) || hasOption(v, api.ControlOption(api.ControlFollow)) {
		out = append(out, k.control)
	}
	if len(buildMenu(v.Options).children) > 0 {
		out = append(out, k.menu)
	}
	if v.Bodies != nil {
		out = append(out, k.look)
	}
	if v.Timeline != nil {
		out = append(out, k.prevTick, k.selectTic)
	}
	return append(out, k.quit)
}
