package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"

	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

var andarielStartingPosition = data.Position{
	X: 22561,
	Y: 9553,
}

var andarielClearPos1 = data.Position{
	X: 22570,
	Y: 9591,
}

var andarielClearPos2 = data.Position{
	X: 22547,
	Y: 9593,
}

var andarielClearPos3 = data.Position{
	X: 22533,
	Y: 9591,
}

var andarielClearPos4 = data.Position{
	X: 22535,
	Y: 9579,
}

var andarielClearPos5 = data.Position{
	X: 22548,
	Y: 9580,
}

var andarielAttackPos1 = data.Position{
	X: 22548,
	Y: 9570,
}

// Placeholder for second attack position
//var andarielAttackPos2 = data.Position{
//	X: 22548,
//	Y: 9590,
//}

type Andariel struct {
	baseRun
}

func (a Andariel) Name() string {
	return string(config.AndarielRun)
}

func (a Andariel) BuildActions() (actions []action.Action) {
	// Moving to Catacombs Level 4
	a.logger.Info("Moving to Catacombs 4")
	actions = append(actions,
		a.builder.WayPoint(area.CatacombsLevel2),
		a.builder.MoveToArea(area.CatacombsLevel3),
		a.builder.MoveToArea(area.CatacombsLevel4),
	)

	// Clearing inside room
	a.logger.Info("Clearing inside room")

	if a.CharacterCfg.Game.Andariel.ClearRoom {
		actions = append(actions,
			a.builder.MoveToCoords(andarielClearPos1),
			a.builder.ClearAreaAroundPlayer(10, data.MonsterAnyFilter()),
			a.builder.MoveToCoords(andarielClearPos2),
			a.builder.ClearAreaAroundPlayer(10, data.MonsterAnyFilter()),
			a.builder.MoveToCoords(andarielClearPos3),
			a.builder.ClearAreaAroundPlayer(10, data.MonsterAnyFilter()),
			a.builder.MoveToCoords(andarielClearPos4),
			a.builder.ClearAreaAroundPlayer(10, data.MonsterAnyFilter()),
			a.builder.MoveToCoords(andarielClearPos5),
			a.builder.ClearAreaAroundPlayer(10, data.MonsterAnyFilter()),
			a.builder.MoveToCoords(andarielAttackPos1),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
		)
	} else {
		actions = append(actions,
			a.builder.MoveToCoords(andarielStartingPosition),
		)
	}

	// Attacking Andariel
	a.logger.Info("Killing Andariel")
	actions = append(actions,
		a.char.KillAndariel(),
	)
	return actions
}
