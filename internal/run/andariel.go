package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
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
	ctx *context.Status
}

func NewAndariel() *Andariel {
	return &Andariel{
		ctx: context.Get(),
	}
}

func (a Andariel) Name() string {
	return string(config.AndarielRun)
}

func (a Andariel) Run() error {
	// Moving to Catacombs Level 4
	a.ctx.Logger.Info("Moving to Catacombs 4")
	err := action.WayPoint(area.CatacombsLevel2)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	err = action.MoveToArea(area.CatacombsLevel3)
	action.OpenTPIfLeader()
	action.MoveToArea(area.CatacombsLevel4)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	// Clearing inside room
	a.ctx.Logger.Info("Clearing inside room")

	if a.ctx.CharacterCfg.Game.Andariel.ClearRoom {
		action.MoveToCoords(andarielClearPos1)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos2)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos3)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos4)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos5)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielAttackPos1)
		action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())

	} else {
		action.MoveToCoords(andarielStartingPosition)
	}

	// Disable item pickup while fighting Andariel (prevent picking up items if nearby monsters die)
	a.ctx.DisableItemPickup()

	// Attacking Andariel
	a.ctx.Logger.Info("Killing Andariel")
	err = a.ctx.Char.KillAndariel()

	// Enable item pickup after the fight
	a.ctx.EnableItemPickup()

	return err
}
