package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
	"math"
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
	return monster.Type == data.MonsterTypeSuperUnique && (monster.Name == npc.OblivionKnight || monster.Name == npc.VenomLord || monster.Name == npc.StormCaster)
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

// FindNearestWalkablePosition finds the nearest walkable position to the given position
func FindNearestWalkablePosition(pos data.Position) data.Position {
	ctx := context.Get()
	if ctx.Data.AreaData.Grid.IsWalkable(pos) {
		return pos
	}

	for radius := 1; radius <= 10; radius++ {
		for x := pos.X - radius; x <= pos.X+radius; x++ {
			for y := pos.Y - radius; y <= pos.Y+radius; y++ {
				checkPos := data.Position{X: x, Y: y}
				if ctx.Data.AreaData.Grid.IsWalkable(checkPos) {
					return checkPos
				}
			}
		}
	}

	return pos
}

func GetSafePositionTowardsMonster(playerPos, monsterPos data.Position, safeDistance int) data.Position {
	dx := float64(monsterPos.X - playerPos.X)
	dy := float64(monsterPos.Y - playerPos.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > float64(safeDistance) {
		ratio := float64(safeDistance) / distance
		safePos := data.Position{
			X: playerPos.X + int(dx*ratio),
			Y: playerPos.Y + int(dy*ratio),
		}
		return FindNearestWalkablePosition(safePos)
	}

	return playerPos
}

func GetSafePositionAwayFromMonster(playerPos, monsterPos data.Position, safeDistance int) data.Position {
	dx := float64(playerPos.X - monsterPos.X)
	dy := float64(playerPos.Y - monsterPos.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < float64(safeDistance) {
		ratio := float64(safeDistance) / distance
		safePos := data.Position{
			X: monsterPos.X + int(dx*ratio),
			Y: monsterPos.Y + int(dy*ratio),
		}
		return FindNearestWalkablePosition(safePos)
	}

	return playerPos
}
