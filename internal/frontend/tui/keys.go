package tui

import (
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// binding ties keys to an option ID. Keys never build commands: a key does something only when
// the view offers its option.
type binding struct {
	keys   key.Binding
	option string // the option ID; empty for the two special bindings below
}

// KeyMap is every binding of the map screen.
type KeyMap struct {
	moves     []binding
	up        key.Binding // w: NE or NW, alternating, so the party climbs a screen column
	down      key.Binding // s: SE or SW, alternating
	waitEvent binding
	waitPhase binding
	nextLead  key.Binding // tab: the next leader option, cycling
	control   key.Binding // f: whichever control option is offered
	prevTick  key.Binding
	nextTick  key.Binding
	selectTic key.Binding // t: list the selected tick
	menu      key.Binding // enter: the action menu
	look      key.Binding // l: look mode
	quit      key.Binding
}

// DefaultKeyMap is the keymap of the spec.
func DefaultKeyMap() KeyMap {
	move := func(k, option string) binding {
		return binding{key.NewBinding(key.WithKeys(k), key.WithHelp(k, "move")), option}
	}
	return KeyMap{
		moves: []binding{
			move("q", "step/nw"), move("e", "step/ne"), move("a", "step/w"),
			move("d", "step/e"), move("z", "step/sw"), move("c", "step/se"),
		},
		up:        key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "up")),
		down:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "down")),
		waitEvent: binding{key.NewBinding(key.WithKeys("."), key.WithHelp(".", "wait")), "wait/event"},
		waitPhase: binding{key.NewBinding(key.WithKeys(","), key.WithHelp(",", "wait to next phase")), "wait/phase"},
		nextLead:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next leader")),
		control:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "lead or follow")),
		prevTick:  key.NewBinding(key.WithKeys("["), key.WithHelp("[ ]", "pick tick")),
		nextTick:  key.NewBinding(key.WithKeys("]")),
		selectTic: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "list tick")),
		menu:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "actions")),
		look:      key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "look")),
		quit:      key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit")),
	}
}

// hasOption reports whether the view offers an option.
func hasOption(v api.View, id string) bool {
	return slices.ContainsFunc(v.Options, func(o api.Option) bool { return o.ID == id })
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
		for _, id := range []string{"control/" + api.ControlIndividual, "control/" + api.ControlFollow} {
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
	east, west := "step/se", "step/sw"
	if up {
		east, west = "step/ne", "step/nw"
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
		for _, o := range v.Options {
			if o.ID != id {
				continue
			}
			if o.Enabled {
				return id, true
			}
			if offered == "" {
				offered = id
			}
		}
	}
	return offered, offered != ""
}

// nextLeadOption picks the member after the leader in a fixed (UID) order, wrapping, so that
// repeated presses visit every member even though leading reorders the party.
func nextLeadOption(v api.View) (string, bool) {
	var uids []uint64
	for _, o := range v.Options {
		if rest, ok := strings.CutPrefix(o.ID, "lead/"); ok {
			if uid, err := strconv.ParseUint(rest, 10, 64); err == nil {
				uids = append(uids, uid)
			}
		}
	}
	if len(uids) == 0 {
		return "", false
	}
	slices.Sort(uids)
	var leader uint64
	if v.Party != nil {
		for _, m := range v.Party.Members {
			if m.Leader {
				leader = uint64(m.UID)
			}
		}
	}
	for _, uid := range uids {
		if uid > leader {
			return "lead/" + strconv.FormatUint(uid, 10), true
		}
	}
	return "lead/" + strconv.FormatUint(uids[0], 10), true
}

// HelpBindings are the bindings to list on the help line: those whose options are present.
func (k KeyMap) HelpBindings(v api.View) []key.Binding {
	var out []key.Binding
	if slices.ContainsFunc(k.moves, func(b binding) bool { return hasOption(v, b.option) }) {
		out = append(out, key.NewBinding(key.WithHelp("q e a d z c", "move")))
	}
	if hasOption(v, "step/ne") || hasOption(v, "step/nw") || hasOption(v, "step/se") || hasOption(v, "step/sw") {
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
	if hasOption(v, "control/"+api.ControlIndividual) || hasOption(v, "control/"+api.ControlFollow) {
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
