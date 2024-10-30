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

func ClearAreaAroundPosition(pos data.Position, radius int, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ClearAreaAroundPosition"
	ctx.Logger.Debug("Clearing area around position...", slog.Int("radius", radius))

	return ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var closestMonster *data.Monster
		closestDistance := float64(radius)

		for i, m := range d.Monsters.Enemies(filter) {
			// Skip monsters outside the area data bounds or unwalkable positions
			if !ctx.Data.AreaData.IsInside(m.Position) || !ctx.Data.AreaData.IsWalkable(m.Position) {
				continue
			}

			monsterDistance := pather.DistanceFromPoint(pos, m.Position)
			playerToMonsterDistance := pather.DistanceFromPoint(ctx.Data.PlayerUnit.Position, m.Position)
			attackDistance := radius

			// Hack the attack distance only for Chaos Sanctuary run
			if ctx.Data.PlayerUnit.Area == area.ChaosSanctuary && IsMonsterSealElite(m) && ctx.CharacterCfg.Game.Diablo.AttackFromDistance != 0 {
				attackDistance = ctx.CharacterCfg.Game.Diablo.AttackFromDistance
			}

			// If monster is within attack range of player, target it
			if playerToMonsterDistance <= attackDistance {
				return m.UnitID, true
			}

			// If monster is within radius of the target position and closer than current closest
			if monsterDistance <= radius && (closestMonster == nil || float64(monsterDistance) < closestDistance) {
				closestMonster = &d.Monsters.Enemies(filter)[i]
				closestDistance = float64(monsterDistance)
			}
		}

		// If we found a monster within the radius but not in attack range, move towards it
		if closestMonster != nil {
			targetPos := ctx.PathFinder.GetSafePositionTowardsMonster(ctx.Data.PlayerUnit.Position, closestMonster.Position, radius)

			if targetPos != ctx.Data.PlayerUnit.Position {
				if err := MoveToCoords(targetPos); err != nil {
					ctx.Logger.Warn("Failed to move to safe position",
						slog.String("error", err.Error()),
						slog.Any("monster", closestMonster.Name),
						slog.Any("position", targetPos))
				}
			}

			return closestMonster.UnitID, true
		}

		return 0, false
	}, nil)
}
