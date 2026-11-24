package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// notMenu are the first segments of option IDs that have their own keys, so they are not in the
// action menu.
var notMenu = map[string]bool{"step": true, "wait": true, "lead": true, "control": true}

// menuNode is one entry of the drill-down: a segment of option IDs, with the options below it.
type menuNode struct {
	segment  string
	prefix   string      // the option ID segments up to and including this one
	option   *api.Option // set when an option ends here
	children []*menuNode
}

// buildMenu groups the options that belong in the action menu by splitting their IDs on "/". The
// core's order is kept; the frontend decides nothing about which options exist or come first.
func buildMenu(options []api.Option) *menuNode {
	root := &menuNode{}
	for i := range options {
		option := &options[i]
		segments := strings.Split(option.ID, "/")
		if notMenu[segments[0]] {
			continue
		}
		node := root
		for depth, segment := range segments {
			var next *menuNode
			for _, child := range node.children {
				if child.segment == segment {
					next = child
					break
				}
			}
			if next == nil {
				next = &menuNode{segment: segment, prefix: strings.Join(segments[:depth+1], "/")}
				node.children = append(node.children, next)
			}
			node = next
		}
		node.option = option
	}
	return root
}

// resolve skips levels that hold a single entry which itself has a choice below it.
func (n *menuNode) resolve() *menuNode {
	for len(n.children) == 1 && n.children[0].option == nil && len(n.children[0].children) > 0 {
		n = n.children[0]
	}
	return n
}

// menuItem is one row of a level.
type menuItem struct {
	label string
	node  *menuNode
}

// FilterValue implements [list.Item].
func (i menuItem) FilterValue() string { return i.label }

// menuDelegate draws one row: the selected one reversed, a disabled option dimmed.
type menuDelegate struct{}

func (menuDelegate) Height() int                         { return 1 }
func (menuDelegate) Spacing() int                        { return 0 }
func (menuDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (menuDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	row, ok := item.(menuItem)
	if !ok {
		return
	}
	style := lipgloss.NewStyle().MaxWidth(max(1, m.Width()))
	if row.node.option != nil && !row.node.option.Enabled {
		style = style.Faint(true)
	}
	if index == m.Index() {
		_, _ = io.WriteString(w, style.Reverse(true).Render("> "+row.label))
		return
	}
	_, _ = io.WriteString(w, style.Render("  "+row.label))
}

// menuState is the open action menu: the levels entered so far and the list of the current one.
type menuState struct {
	levels []*menuNode // displayed levels, outermost first
	list   list.Model
	names  map[api.UID]string
	width  int
	height int
}

// newMenu opens the menu on the view's options, or returns nil when there is nothing to choose.
func newMenu(v api.View, width, height int) *menuState {
	root := buildMenu(v.Options)
	if len(root.children) == 0 {
		return nil
	}
	m := &menuState{names: bodyNames(v.Bodies), width: width, height: height}
	m.push(root.resolve())
	return m
}

// segmentLabel is how a group entry reads: the name of the body or part a number stands for, else
// the last word of a content ID.
func (m *menuState) segmentLabel(segment string) string {
	if uid, err := strconv.ParseUint(segment, 10, 64); err == nil {
		if name, ok := m.names[api.UID(uid)]; ok {
			return name
		}
		return segment
	}
	if i := strings.LastIndexByte(segment, '.'); i >= 0 {
		segment = segment[i+1:]
	}
	if segment == "" {
		return segment
	}
	return strings.ToUpper(segment[:1]) + segment[1:]
}

func (m *menuState) push(node *menuNode) {
	m.levels = append(m.levels, node)
	items := make([]list.Item, len(node.children))
	labels := map[string]int{}
	leaves := map[string]int{}
	for _, child := range node.children {
		labels[m.segmentLabel(child.segment)]++
		if child.option != nil {
			leaves[child.option.Label]++
		}
	}
	for i, child := range node.children {
		label := m.segmentLabel(child.segment)
		switch {
		case child.option != nil:
			label = child.option.Label
			if leaves[label] > 1 {
				// Options that differ only in which parts do it read the same: tell them apart.
				label = m.segmentLabel(child.segment) + " #" + child.segment + ": " + label
			}
			if !child.option.Enabled && child.option.Reason != "" {
				label += " (" + child.option.Reason + ")"
			}
		case labels[label] > 1:
			label += " #" + child.segment
		}
		items[i] = menuItem{label: label, node: child}
	}
	l := list.New(items, menuDelegate{}, m.width, max(1, m.height-1))
	l.SetShowTitle(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	m.list = l
}

// back goes up a level, and reports false when the menu should close.
func (m *menuState) back() bool {
	if len(m.levels) <= 1 {
		return false
	}
	m.levels = m.levels[:len(m.levels)-1]
	upper := m.levels[len(m.levels)-1]
	m.levels = m.levels[:len(m.levels)-1]
	m.push(upper)
	return true
}

// selected is the highlighted node, or nil.
func (m *menuState) selected() *menuNode {
	if item, ok := m.list.SelectedItem().(menuItem); ok {
		return item.node
	}
	return nil
}

// targetUID is the body the highlighted entry aims at: the last number in its option ID that is a
// body in sight.
func (m *menuState) targetUID(bodies *api.BodiesView) api.UID {
	node := m.selected()
	if node == nil || bodies == nil {
		return 0
	}
	var found api.UID
	for segment := range strings.SplitSeq(node.prefix, "/") {
		uid, err := strconv.ParseUint(segment, 10, 64)
		if err != nil {
			continue
		}
		for _, body := range bodies.Bodies {
			if body.UID == api.UID(uid) {
				found = body.UID
			}
		}
	}
	return found
}

// Render draws the open level: the path so far, then the list.
func (m *menuState) Render() string {
	path := []string{"Actions"}
	if prefix := m.levels[len(m.levels)-1].prefix; prefix != "" {
		for segment := range strings.SplitSeq(prefix, "/") {
			path = append(path, m.segmentLabel(segment))
		}
	}
	return fmt.Sprintf("%s\n%s", strings.Join(path, " > "), m.list.View())
}

// update handles a key in the menu. It returns the option to send when a leaf is chosen, whether
// the menu should close without one, and the key to hand the list.
func (m *menuState) update(msg tea.KeyPressMsg) (choose string, closed bool) {
	switch msg.String() {
	case "enter":
		node := m.selected()
		switch {
		case node == nil:
		case node.option != nil:
			return node.option.ID, false
		default:
			m.push(node.resolve())
		}
	case "esc":
		return "", !m.back()
	case "up", "down", "k", "j", "pgup", "pgdown", "home", "end":
		m.list, _ = m.list.Update(msg)
	}
	return "", false
}
