package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
)

type Pit struct {
	baseRun
}

func (a Pit) Name() string {
	return "Pit"
}

func (a Pit) BuildActions() (actions []action.Action) {
	actions = append(actions,
		a.builder.WayPoint(area.BlackMarsh),      // Moving to starting point (OuterCloister)
		a.builder.MoveToArea(area.TamoeHighland), // Move to TamoeHighland
	)

	// Travel to pit level 1
	a.logger.Info("Travel to pit level 1")
	actions = append(actions, a.builder.MoveToArea(area.PitLevel1))

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.NewStepChain(func(_ data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	return append(actions,
		a.builder.ClearArea(true, data.MonsterAnyFilter()), // Clear pit level 1
		a.builder.MoveToArea(area.PitLevel2),               // Travel to pit level 2
		a.builder.ClearArea(true, data.MonsterAnyFilter()), // Clear pit level 2
	)
}
