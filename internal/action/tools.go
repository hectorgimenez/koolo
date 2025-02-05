package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func OpenTPIfLeader() error {
	ctx := context.Get()
	ctx.SetLastAction("OpenTPIfLeader")

	isLeader := ctx.CharacterCfg.Companion.Leader

	if isLeader {
		return step.OpenPortal()
	}

	return nil
}

func IsMonsterSealElite(monster data.Monster) bool {
	return monster.Type == data.MonsterTypeSuperUnique && (monster.Name == npc.OblivionKnight || monster.Name == npc.VenomLord || monster.Name == npc.StormCaster)
}

func PostRun(isLastRun bool) error {
	ctx := context.Get()
	ctx.SetLastAction("PostRun")

	// Allow some time for items drop to the ground, otherwise we might miss some
	utils.Sleep(200)
	ClearAreaAroundPlayer(5, data.MonsterAnyFilter())
	ItemPickup(-1)

	// Don't return town on last run
	if !isLastRun {
		return ReturnTown()
	}

	return nil
}
func AreaCorrection() error {
	ctx := context.Get()
	currentArea := ctx.Data.PlayerUnit.Area
	expectedArea := ctx.CurrentGame.AreaCorrection.ExpectedArea

	// Skip correction if in town, if we're in the expected area, or if expected area is not set
	if currentArea.IsTown() || currentArea == expectedArea || expectedArea == 0 {
		return nil
	}

	if ctx.CurrentGame.AreaCorrection.Enabled && ctx.CurrentGame.AreaCorrection.ExpectedArea != ctx.Data.AreaData.Area {
		ctx.Logger.Info("Accidentally went to adjacent area, returning to expected area",
			"current", ctx.Data.AreaData.Area.Area().Name,
			"expected", ctx.CurrentGame.AreaCorrection.ExpectedArea.Area().Name)
		return MoveToArea(ctx.CurrentGame.AreaCorrection.ExpectedArea)
	}

	return nil
}
func HidePortraits() error {
	ctx := context.Get()
	ctx.SetLastAction("HidePortraits")

	// Hide portraits if configured
	if ctx.CharacterCfg.HidePortraits && ctx.Data.OpenMenus.PortraitsShown {
		ctx.HID.PressKey(ctx.Data.KeyBindings.ShowPortraits.Key1[0])
	}
	return nil
}
func ClearMessages() error {
	ctx := context.Get()
	ctx.SetLastAction("ClearMessages")
	ctx.HID.PressKey(ctx.Data.KeyBindings.ClearMessages.Key1[0])
	return nil
}
