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
	ctx.ContextDebug.LastAction = "OpenTPIfLeader"

	isLeader := ctx.CharacterCfg.Companion.Leader

	if isLeader {
		return step.OpenPortal()
	}

	return nil
}

func IsMonsterSealElite(monster data.Monster) bool {
	if monster.Type == data.MonsterTypeSuperUnique && (monster.Name == npc.OblivionKnight || monster.Name == npc.VenomLord || monster.Name == npc.StormCaster) {
		return true
	}

	return false
}

func PostRun(isLastRun bool) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "PostRun"

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
	expectedArea := ctx.CurrentGame.ExpectedArea

	// Skip correction if in town, if we're in the expected area, or if expected area is not set
	if currentArea.IsTown() || currentArea == expectedArea || expectedArea == 0 {
		return nil
	}

	// Check if the expected area is adjacent to our current area
	for _, adjacentLevel := range ctx.Data.AdjacentLevels {
		if adjacentLevel.Area == expectedArea {
			ctx.Logger.Info("Accidentally went to adjacent area, returning to expected area",
				"current", currentArea.Area().Name,
				"expected", expectedArea.Area().Name)

			err := MoveToArea(expectedArea)
			if err != nil {
				ctx.Logger.Warn("Failed to move back to expected area",
					"error", err)
			}
			return nil
		}
	}

	return nil
}
