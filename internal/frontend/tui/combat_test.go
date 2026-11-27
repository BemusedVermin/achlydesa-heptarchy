package tui_test

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
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

// startApp is the sample game at its start, with all of its actions and a standing moving clay: these
// tests browse menus, and a hunting creature would change what they see.
func startApp(t *testing.T) *tui.App {
	t.Helper()
	g, _, _ := gametest.New(t, 1, contenttest.StandingCrawler)
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
	golden(t, "menu_actions", tui.MenuPane(app))
	selectEntry(t, app, "Strike")
	pressCode(app, tea.KeyEnter)
	view := tui.MenuPane(app)
	golden(t, "menu_types", view)
	if !strings.Contains(view, "Actions > Strike") {
		t.Errorf("no path on the type level:\n%s", view)
	}
	selectEntry(t, app, "Cut")
	pressCode(app, tea.KeyEnter)
	view = tui.MenuPane(app)
	golden(t, "menu_parts", view)
	if !strings.Contains(view, "maw") {
		t.Errorf("the part level does not list the creature's parts:\n%s", view)
	}
	if !strings.Contains(view, "Actions > Strike > Cut > "+creature(app).Name) {
		t.Errorf("the single target's level was not skipped into the path:\n%s", view)
	}
	selectEntry(t, app, "maw")
	pressCode(app, tea.KeyEnter)
	golden(t, "menu_arms", tui.MenuPane(app))
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
	golden(t, "actor_free", tui.RenderBody(currentMember(t, app), app.State.Bodies, panelWidth))
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
	press(app, "x")
	if strings.Contains(app.View().Content, "Actions") {
		t.Error("the menu stayed open after a command")
	}
	body := bodyOf(t, app, app.State.Bodies.Bodies[0].UID)
	golden(t, "actor_windup", tui.RenderBody(body, app.State.Bodies, panelWidth))
	if !strings.Contains(app.View().Content, "~ Strike") {
		t.Errorf("the panel does not show the swing:\n%s", app.View().Content)
	}
}

func TestALongTargetNameWrapsTheSwingInsteadOfCuttingItsTick(t *testing.T) {
	windup := &api.WindupView{Action: "Strike", Target: 2, Lands: 3, When: "lands at tick 86"}
	bodies := &api.BodiesView{Bodies: []api.BodyView{
		{UID: 1, Name: "The Reaper", Channels: []api.ChannelView{{UID: 3, Slot: "arm", Name: "arm", Windup: windup}}},
		{UID: 2, Name: "The very long-named moving clay"},
	}}
	out := tui.RenderBody(bodies.Bodies[0], bodies, panelWidth)
	if !strings.Contains(strings.ReplaceAll(out, "\n", " "), "lands at tick 86") {
		t.Errorf("the swing line lost its tick:\n%s", out)
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
	golden(t, "actor_sweep", tui.RenderBody(bodyOf(t, app, mover), app.State.Bodies, panelWidth))
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

func TestLookingAtAnEmptyHexAndListingATickUseTheViewsText(t *testing.T) {
	app := startApp(t)
	if got := tui.RenderLook(app.State, api.Hex{Q: -100, R: -100}, panelWidth); got != app.State.Labels.LookEmpty ||
		got == "" {
		t.Errorf("look at an empty hex = %q, want the view's %q", got, app.State.Labels.LookEmpty)
	}
	press(app, "t")
	want := app.State.Timeline.Entries[0].TickLabel
	if want == "" || !strings.Contains(app.View().Content, want) {
		t.Errorf("the tick popup does not head with %q:\n%s", want, app.View().Content)
	}
}

func TestSeveredArmShowsOnThePanel(t *testing.T) {
	app := startApp(t)
	var victim api.BodyView
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
	golden(t, "actor_severed", tui.RenderBody(body, app.State.Bodies, panelWidth))
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
		{Tick: 11, Kind: api.EntryReady, Source: 2, Glyph: "b", Color: "red"},
		{Tick: 13, Kind: api.EntryLands, Source: 2, Target: 1, Glyph: "b", Color: "red"},
	}}
	golden(t, "timeline_marked", tui.RenderTimelineMarked(tl, bodies, 20, 0))
}

// huntingApp is the sample game as shipped: the moving clay hunts and fights back.
func huntingApp(t *testing.T) *tui.App {
	t.Helper()
	reg := contenttest.Sample(t)
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

// TestWholeFightEndsAndTheMapResumesFollow plays the hall's fight through the menu's options, from
// the alert to the creature's death, and checks the alert ends and the party follows again.
func TestWholeFightEndsAndTheMapResumesFollow(t *testing.T) {
	app := huntingApp(t)
	sawAlert, sawThreat := false, false
	for range 600 {
		if app.State.Party.Alert {
			sawAlert = true
		}
		if app.State.Bodies != nil && len(app.State.Bodies.Threats) > 0 {
			sawThreat = true
		}
		if sawAlert && !app.State.Party.Alert && creature(app) == nil {
			break
		}
		if foe := creature(app); foe != nil {
			// The last part listed is the core, which a severing kills the creature through.
			var sever []api.Option
			for _, o := range optionsWithPrefix(app, "sever/type.soak/") {
				if strings.Contains(o.ID, "/"+uidString(foe.UID)+"/") && o.Enabled {
					sever = append(sever, o)
				}
			}
			if len(sever) > 0 {
				app.Choose(sever[len(sever)-1].ID)
			}
		} else {
			press(app, "d")
		}
		press(app, ".")
	}
	if !sawAlert {
		t.Fatal("the party never came on alert")
	}
	if app.State.Party.Alert || creature(app) != nil {
		t.Fatalf("the fight never ended:\n%s", app.View().Content)
	}
	if strings.Contains(app.View().Content, "ALERT") {
		t.Errorf("the status line still says ALERT:\n%s", app.View().Content)
	}
	if app.State.Party.Control != api.ControlFollow {
		t.Errorf("control = %q, want follow", app.State.Party.Control)
	}
	if !sawThreat {
		t.Error("the hunting movingClay never showed a threat")
	}
}

func commitStrike(t *testing.T, app *tui.App) {
	t.Helper()
	pressCode(app, tea.KeyEnter)
	for _, label := range []string{"Strike", "Cut", "maw"} {
		selectEntry(t, app, label)
		pressCode(app, tea.KeyEnter)
	}
	pressCode(app, tea.KeyEnter)
}

func TestAFightCommandOpensAPlayByPlayThatStepsAndCloses(t *testing.T) {
	app := fightApp(t)
	commitStrike(t, app)
	var beats []api.LogLine
	for _, line := range app.State.Log {
		if line.Pause {
			beats = append(beats, line)
		}
	}
	if len(beats) == 0 {
		t.Fatal("a Strike committed produced no fight beat")
	}
	view := app.View().Content
	if !strings.Contains(view, "begins") || strings.Contains(view, "Actions") {
		t.Fatalf("the play-by-play is not showing the start line:\n%s", view)
	}
	if len(beats) > 1 {
		press(app, "n")
		if !strings.Contains(app.View().Content, beats[1].Text) {
			t.Errorf("one key did not show the second line:\n%s", app.View().Content)
		}
		if len(beats) > 2 && strings.Contains(app.View().Content, beats[2].Text) {
			t.Errorf("one key showed more than one line:\n%s", app.View().Content)
		}
	}
	press(app, "x")
	if !strings.Contains(app.View().Content, beats[len(beats)-1].Text) {
		t.Errorf("skip did not show every line:\n%s", app.View().Content)
	}
	press(app, "x")
	if strings.Contains(app.View().Content, "skip") {
		t.Errorf("the play-by-play stayed open after the last line:\n%s", app.View().Content)
	}
}

var (
	startLine  = regexp.MustCompile(`\S.* begins \S.* at \S.* \(lands at tick \d+\)\.`)
	damageLine = regexp.MustCompile(`\S.*'s \S.* hits \S.*'s \S.*: \d+ integrity lost, \d+ left\.`)
)

func TestAHitLogsTheStartAndTheDamage(t *testing.T) {
	app := fightApp(t)
	commitStrike(t, app)
	press(app, "x")
	press(app, "x")
	for range 20 {
		var text strings.Builder
		for _, line := range app.State.Log {
			text.WriteString(line.Text + "\n")
		}
		if strings.Contains(text.String(), "integrity lost") {
			if !startLine.MatchString(text.String()) || !damageLine.MatchString(text.String()) {
				t.Errorf("start or damage line missing or malformed:\n%s", text.String())
			}
			return
		}
		press(app, ".")
		press(app, "x")
		press(app, "x")
	}
	t.Error("no damage line ever appeared in the log")
}

func TestWalkingPausesNothing(t *testing.T) {
	app := startApp(t)
	press(app, "d")
	press(app, ".")
	for _, line := range app.State.Log {
		if line.Pause {
			t.Errorf("line %q pauses while walking", line.Text)
		}
	}
	if strings.Contains(app.View().Content, "skip") {
		t.Error("a play-by-play opened while walking")
	}
}
