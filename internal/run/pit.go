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
	openChests := a.CharacterCfg.Game.Pit.OpenChests

	actions = append(actions,
		a.builder.WayPoint(area.OuterCloister),
		a.builder.MoveToArea(area.MonasteryGate),
		a.builder.MoveToArea(area.TamoeHighland),
	)

	if a.CharacterCfg.Game.Pit.MoveThroughBlackMarsh {
		actions = []action.Action{
			a.builder.WayPoint(area.BlackMarsh),
			a.builder.MoveToArea(area.TamoeHighland),
		}
	}

	a.logger.Info("Travel to pit level 1")
	actions = append(actions, a.builder.MoveToArea(area.PitLevel1))

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	return append(actions,
		a.builder.ClearArea(openChests, data.MonsterAnyFilter()), // Clear pit level 1
		a.builder.MoveToArea(area.PitLevel2),                     // Travel to pit level 2
		a.builder.ClearArea(openChests, data.MonsterAnyFilter()), // Clear pit level 2
	)
}
