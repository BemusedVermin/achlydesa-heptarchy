package game_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// badActions loads the fixture of actions whose effects are each wrong in one way.
func badActions(t *testing.T) map[string]content.ActionDef {
	t.Helper()
	data, err := os.ReadFile("testdata/bad_actions.json")
	if err != nil {
		t.Fatal(err)
	}
	var actions map[string]content.ActionDef
	if err := json.Unmarshal(data, &actions); err != nil {
		t.Fatal(err)
	}
	return actions
}

func TestNewRejectsBadActionEffects(t *testing.T) {
	bundle, err := content.LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Start = content.Start{}
	for id, def := range badActions(t) {
		bundle.Actions[id] = def
	}
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatalf("the fixture should be valid content with only its effects wrong: %v", err)
	}
	_, err = game.New(game.Config{Seed: 1, Content: reg})
	if err == nil {
		t.Fatal("game.New accepted actions with bad effects")
	}
	for _, want := range []string{
		`actions["action.bad_primitive"].on_land[0]: unknown primitive "no_such_primitive"`,
		`actions["action.bad_parameter"].on_start[0] (lock): missing parameter "ticks"`,
		`actions["action.bad_selector"].on_land[0] (damage).parts: bad channel selector "bogus"`,
		`actions["action.bad_status"].on_land[0] (add_status): unknown status "status.nope"`,
		`actions["action.bad_item"].on_land[0] (consume): unknown item "item.nope"`,
		`actions["action.bad_faction"].on_land[0] (spawn_from_part): unknown faction "faction.nope"`,
		`actions["action.bad_piece"].on_land[0] (attach_part): item "item.coal" is not a piece`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not list %q:\n%v", want, err)
		}
	}
}
