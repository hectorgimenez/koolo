package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
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

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	return append(actions,
		a.builder.ClearArea(true, data.MonsterAnyFilter()), // Clear pit level 1
		a.builder.MoveToArea(area.PitLevel2),               // Travel to pit level 2
		a.builder.ClearArea(true, data.MonsterAnyFilter()), // Clear pit level 2
	)
}
