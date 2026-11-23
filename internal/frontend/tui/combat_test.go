package tui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/frontend/tui"
)

const panelWidth = 44

// fightApp plays the sample game, with all of its actions, to the point where the deciding member
// can strike the hall's creature.
func fightApp(t *testing.T) *tui.App {
	t.Helper()
	app := startApp(t)
	for range 80 {
		if creature(app) != nil {
			if _, ok := findOption(app, "strike/", "/"+creatureUID(app)+"/"); ok {
				return app
			}
		}
		press(app, "d")
		press(app, ".")
	}
	t.Fatal("never came within reach of the creature")
	return nil
}

// startApp is the sample game at its start, with all of its actions and a standing crawler: these
// tests browse menus, and a hunting creature would change what they see.
func startApp(t *testing.T) *tui.App {
	t.Helper()
	bundle, err := content.LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	crawler := bundle.Templates["template.creature.crawler"]
	crawler.Actor = &content.ActorDef{Controller: "rest"}
	crawler.Actions = nil
	crawler.Lists = nil
	bundle.Templates["template.creature.crawler"] = crawler
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(game.Config{Seed: 1, Content: reg})
	if err != nil {
		t.Fatal(err)
	}
	app, err := tui.NewApp(g)
	if err != nil {
		t.Fatal(err)
	}
	app.Width, app.Height = 180, 30
	return app
}

func creature(app *tui.App) *api.BodyView {
	for _, b := range app.State.Bodies.Bodies {
		if b.Hostile {
			return &b
		}
	}
	return nil
}

func creatureUID(app *tui.App) string {
	return uidString(creature(app).UID)
}

func uidString(uid api.UID) string {
	digits := ""
	for n := uint64(uid); n > 0; n /= 10 {
		digits = string(rune('0'+n%10)) + digits
	}
	return digits
}

func optionsWithPrefix(app *tui.App, prefix string) []api.Option {
	var out []api.Option
	for _, o := range app.State.Options {
		if strings.HasPrefix(o.ID, prefix) {
			out = append(out, o)
		}
	}
	return out
}

func findOption(app *tui.App, prefix, contains string) (api.Option, bool) {
	for _, o := range optionsWithPrefix(app, prefix) {
		if strings.Contains(o.ID, contains) {
			return o, true
		}
	}
	return api.Option{}, false
}

func pressCode(app *tui.App, code rune) {
	app.Update(tea.KeyPressMsg{Code: code})
}

// selectEntry moves the menu cursor down to the entry with this label.
func selectEntry(t *testing.T, app *tui.App, label string) {
	t.Helper()
	for range 400 {
		if strings.Contains(app.View().Content, "> "+label) {
			return
		}
		pressCode(app, tea.KeyDown)
	}
	t.Fatalf("no menu entry %q:\n%s", label, app.View().Content)
}

func TestAlertShowsOnTheStatusLine(t *testing.T) {
	app := fightApp(t)
	if !app.State.Party.Alert {
		t.Fatal("no alert near the creature")
	}
	if !strings.Contains(app.View().Content, "ALERT") {
		t.Error("the status line does not say the party is on alert")
	}
}

func TestMenuDrillsDownSkipsSingleLevelsAndBacksOut(t *testing.T) {
	app := fightApp(t)
	history := len(app.Game.History())
	pressCode(app, tea.KeyEnter)
	view := app.View().Content
	golden(t, "menu_actions", view)
	selectEntry(t, app, "Strike")
	pressCode(app, tea.KeyEnter)
	view = app.View().Content
	golden(t, "menu_types", view)
	if !strings.Contains(view, "Actions > Strike") {
		t.Errorf("no path on the type level:\n%s", view)
	}
	selectEntry(t, app, "Cut")
	pressCode(app, tea.KeyEnter)
	// One target is in reach, so its level is skipped and the parts come next.
	view = app.View().Content
	golden(t, "menu_parts", view)
	if !strings.Contains(view, "maw") {
		t.Errorf("the part level does not list the creature's parts:\n%s", view)
	}
	if !strings.Contains(view, "Actions > Strike > Cut > "+creature(app).Name) {
		t.Errorf("the skipped target level is missing from the path:\n%s", view)
	}
	selectEntry(t, app, "maw")
	pressCode(app, tea.KeyEnter)
	golden(t, "menu_arms", app.View().Content)
	// Esc goes up one level at a time and then closes the menu.
	pressCode(app, tea.KeyEscape)
	if !strings.Contains(app.View().Content, "maw") {
		t.Error("esc from the arms did not return to the parts")
	}
	for range 3 {
		pressCode(app, tea.KeyEscape)
	}
	if strings.Contains(app.View().Content, "Actions") {
		t.Error("the menu is still open")
	}
	if got := len(app.Game.History()); got != history {
		t.Errorf("%d commands sent by browsing, want none", got-history)
	}
}

func TestMenuLeafSendsTheOptionAndTheActorPanelShowsTheSwing(t *testing.T) {
	app := fightApp(t)
	golden(t, "actor_free", tui.RenderBody(currentMember(t, app), app.State.Bodies, app.State.Tick, panelWidth))
	pressCode(app, tea.KeyEnter)
	selectEntry(t, app, "Strike")
	pressCode(app, tea.KeyEnter)
	selectEntry(t, app, "Cut")
	pressCode(app, tea.KeyEnter)
	selectEntry(t, app, "maw")
	pressCode(app, tea.KeyEnter)
	history := len(app.Game.History())
	pressCode(app, tea.KeyEnter)
	if got := len(app.Game.History()); got != history+1 {
		t.Fatalf("choosing a leaf sent %d commands, want 1", got-history)
	}
	if strings.Contains(app.View().Content, "Actions") {
		t.Error("the menu stayed open after a command")
	}
	body := bodyOf(t, app, app.State.Bodies.Bodies[0].UID)
	golden(t, "actor_windup", tui.RenderBody(body, app.State.Bodies, app.State.Tick, panelWidth))
	if !strings.Contains(app.View().Content, "~ Strike") {
		t.Errorf("the panel does not show the swing:\n%s", app.View().Content)
	}
}

func currentMember(t *testing.T, app *tui.App) api.BodyView {
	t.Helper()
	return bodyOf(t, app, app.State.Bodies.Current)
}

func bodyOf(t *testing.T, app *tui.App, uid api.UID) api.BodyView {
	t.Helper()
	for _, b := range app.State.Bodies.Bodies {
		if b.UID == uid {
			return b
		}
	}
	t.Fatalf("no body %d in sight", uid)
	return api.BodyView{}
}

func TestSweepThenLookShowsMaterials(t *testing.T) {
	app := fightApp(t)
	enemy := *creature(app)
	var enemyHex api.Hex
	for _, e := range app.State.Map.Entities {
		if e.UID == enemy.UID {
			enemyHex = e.Hex
		}
	}
	before := tui.RenderLook(app.State, enemyHex, panelWidth)
	golden(t, "look_before", before)
	if !strings.Contains(before, "?") {
		t.Errorf("an unknown body shows its materials before a Sweep:\n%s", before)
	}
	sweep, ok := findOption(app, "sweep/", "/"+uidString(enemy.UID))
	if !ok {
		t.Fatal("no sweep option")
	}
	mover := app.State.Bodies.Current
	app.Choose(sweep.ID)
	golden(t, "actor_sweep", tui.RenderBody(bodyOf(t, app, mover), app.State.Bodies, app.State.Tick, panelWidth))
	for range 40 {
		if !strings.Contains(tui.RenderLook(app.State, enemyHex, panelWidth), "?") {
			break
		}
		press(app, ".")
	}
	after := tui.RenderLook(app.State, enemyHex, panelWidth)
	golden(t, "look_after", after)
	if strings.Contains(after, "?") {
		t.Errorf("the Sweep taught nothing about the body's materials:\n%s", after)
	}
}

func TestLookModeMovesACursorAndLeaves(t *testing.T) {
	app := fightApp(t)
	press(app, "l")
	if !strings.Contains(app.View().Content, "Looking") {
		t.Fatal("look mode did not open")
	}
	for range 20 {
		if strings.Contains(app.View().Content, creature(app).Name+" (") {
			break
		}
		press(app, "d")
	}
	view := app.View().Content
	if !strings.Contains(view, "maw") {
		t.Errorf("the cursor never reached the creature:\n%s", view)
	}
	history := len(app.Game.History())
	press(app, "d")
	if len(app.Game.History()) != history {
		t.Error("a step key moved the party in look mode")
	}
	pressCode(app, tea.KeyEscape)
	if strings.Contains(app.View().Content, "Looking") {
		t.Error("esc did not leave look mode")
	}
}

func TestSeveredArmShowsOnThePanel(t *testing.T) {
	app := startApp(t)
	var victim api.BodyView // the member who is not deciding
	for _, b := range app.State.Bodies.Bodies {
		if b.Member && b.UID != app.State.Bodies.Current {
			victim = b
		}
	}
	var arm api.UID
	for _, c := range victim.Channels {
		if c.Slot == "arm" {
			arm = c.UID
			break
		}
	}
	sever, ok := findOption(app, "sever/", "/"+uidString(victim.UID)+"/"+uidString(arm))
	if !ok {
		t.Fatalf(
			"no sever option against the other member's arm %d of %d: %v",
			arm,
			victim.UID,
			optionsWithPrefix(app, "sever/")[:3],
		)
	}
	app.Choose(sever.ID)
	for range 60 {
		if got := bodyOf(t, app, victim.UID); armIntegrity(got, arm) < 0 {
			break
		}
		press(app, ".")
	}
	body := bodyOf(t, app, victim.UID)
	if armIntegrity(body, arm) >= 0 {
		t.Fatalf("the arm was not severed: %+v", body.Channels)
	}
	golden(t, "actor_severed", tui.RenderBody(body, app.State.Bodies, app.State.Tick, panelWidth))
}

func armIntegrity(body api.BodyView, arm api.UID) int {
	for _, c := range body.Channels {
		if c.UID == arm {
			return c.Integrity
		}
	}
	return -1
}

func TestOverlayMarksThreatsAndTimeline(t *testing.T) {
	bodies := &api.BodiesView{
		Bodies: []api.BodyView{
			{UID: 1, Name: "Hero", Member: true},
			{UID: 2, Name: "Beast", Hostile: true},
		},
		Threats: []api.ThreatView{{Attacker: 2, Target: 1, Action: "Bite", Lands: 3, Tick: 13}},
	}
	m := &api.MapView{
		Cells: []api.CellView{
			{Hex: api.Hex{Q: 0}, Glyph: ".", Color: "gray", Visible: true},
			{Hex: api.Hex{Q: 1}, Glyph: ".", Color: "gray", Visible: true},
			{Hex: api.Hex{Q: 2}, Glyph: ".", Color: "gray", Visible: true},
		},
		Entities: []api.EntityView{
			{UID: 1, Hex: api.Hex{Q: 0}, Glyph: "@", Color: "cyan"},
			{UID: 2, Hex: api.Hex{Q: 2}, Glyph: "b", Color: "red"},
		},
	}
	party := &api.PartyView{Members: []api.MemberView{{UID: 1}}, Hex: api.Hex{}}
	hex := api.Hex{Q: 1}
	out := tui.RenderMapOverlay(m, party, tui.OverlayFor(bodies, 2, &hex), 8, 1)
	golden(t, "map_overlay", out)
	if !strings.Contains(out, "3") {
		t.Errorf("no landing count next to the threatened member:\n%s", out)
	}
	tl := &api.TimelineView{Now: 10, Window: 5, Entries: []api.TimelineEntry{
		{Tick: 11, Kind: "ready", Actor: 2, Glyph: "b", Color: "red"},
		{Tick: 13, Kind: "lands", Actor: 2, Target: 1, Glyph: "b", Color: "red"},
	}}
	golden(t, "timeline_marked", tui.RenderTimelineMarked(tl, bodies, 20, 0))
}
