package action

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func ClearAreaAroundPlayer(radius int, filter data.MonsterFilter) error {
	return ClearAreaAroundPosition(context.Get().Data.PlayerUnit.Position, radius, filter)
}

// let character's specific combat logic handle attack distance (no overwrite)
func ClearAreaAroundPosition(pos data.Position, radius int, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ClearAreaAroundPosition"
	ctx.Logger.Debug("Clearing area around position...", slog.Int("radius", radius))

	return ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var targetMonster *data.Monster
		var minDistance = float64(radius)

		for _, m := range d.Monsters.Enemies(filter) {
			// Skip monsters outside the area data bounds or unwalkable positions
			if !ctx.Data.AreaData.IsInside(m.Position) || !ctx.Data.AreaData.IsWalkable(m.Position) {
				continue
			}

			// Check if monster is within the overall clearing radius of the target position
			distanceToTarget := pather.DistanceFromPoint(pos, m.Position)
			if distanceToTarget > radius {
				continue
			}

			// Only update target if this monster is closer to the clearing center
			if targetMonster == nil || float64(distanceToTarget) < minDistance {
				targetMonster = &m
				minDistance = float64(distanceToTarget)
			}
		}

		//can be removed no longer used
		if targetMonster != nil {
			// Special case for Chaos Sanctuary
			if ctx.Data.PlayerUnit.Area == area.ChaosSanctuary && IsMonsterSealElite(*targetMonster) && ctx.CharacterCfg.Game.Diablo.AttackFromDistance != 0 {
				targetPos := ctx.PathFinder.GetSafePositionTowardsMonster(ctx.Data.PlayerUnit.Position, targetMonster.Position, ctx.CharacterCfg.Game.Diablo.AttackFromDistance)
				if targetPos != ctx.Data.PlayerUnit.Position {
					if err := MoveToCoords(targetPos); err != nil {
						ctx.Logger.Warn("Failed to move to safe position",
							slog.String("error", err.Error()),
							slog.Any("monster", targetMonster.Name),
							slog.Any("position", targetPos))
					}
				}
			}

			return targetMonster.UnitID, true
		}

		return 0, false
	}, nil)
}
