package actions_test

import (
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

func TestAWellOffersATakeOfWaterAndNoneOfRation(t *testing.T) {
	g, w, _ := gametest.New(t, 1, nil)
	body, ok := timeline.Current(w)
	if !ok {
		t.Fatal("no body is deciding")
	}
	well, err := w.Spawn("template.source.well", &comp.Pos{At: ecs.NewMap[comp.Pos](w.ECS).Get(body).At})
	if err != nil {
		t.Fatal(err)
	}
	view, err := g.View()
	if err != nil {
		t.Fatal(err)
	}
	wellUID := w.UIDOf(well)
	takes := map[string]bool{}
	for _, option := range view.Options {
		if do, isDo := option.Command.(api.Do); isDo && do.Action == "action.take" && do.Target == wellUID {
			takes[do.Choices["ware"]] = option.Enabled
		}
	}
	if enabled, offered := takes["item.water"]; !offered || !enabled {
		t.Errorf("take water offered %v, enabled %v; want an enabled option", offered, enabled)
	}
	if _, offered := takes["item.ration"]; offered {
		t.Error("a well offers a take of ration")
	}

	ration := actions.Intent{
		Actor: body, Action: "action.take", Target: well, Choices: map[string]string{"ware": "item.ration"},
	}
	if ok, reason := actions.Validate(w, ration); ok || !strings.HasSuffix(string(reason.ID), "none_there") {
		t.Errorf("Validate(take ration from a well) = %v, %q; want refused as none_there", ok, reason.ID)
	}
	for _, in := range actions.Startable(w, body) {
		if in.Action == "action.take" && in.Target == well && in.Choices["ware"] == "item.ration" {
			t.Error("an agent may start a take of ration from a well")
		}
	}
}
