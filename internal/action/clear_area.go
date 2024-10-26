package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"log/slog"
	"sort"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func ClearAreaAroundPlayer(distance int, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ClearAreaAroundPlayer"

	originalPosition := ctx.Data.PlayerUnit.Position

	ctx.Logger.Debug("Clearing area around character...", slog.Int("distance", distance))

	return ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		monsters := d.Monsters.Enemies(filter)
		sort.Slice(monsters, func(i, j int) bool {
			distI := pather.DistanceFromPoint(originalPosition, monsters[i].Position)
			distJ := pather.DistanceFromPoint(originalPosition, monsters[j].Position)
			return distI < distJ
		})

		for _, m := range monsters {
			monsterDist := pather.DistanceFromPoint(originalPosition, m.Position)
			engageDistance := distance

			if ctx.Data.PlayerUnit.Area == area.ChaosSanctuary && IsMonsterSealElite(m) && ctx.CharacterCfg.Game.Diablo.AttackFromDistance != 0 {
				engageDistance = ctx.CharacterCfg.Game.Diablo.AttackFromDistance

				if monsterDist <= engageDistance {
					var targetPos data.Position
					currentPos := ctx.Data.PlayerUnit.Position

					if monsterDist > engageDistance {
						targetPos = step.GetSafePositionTowardsMonster(currentPos, m.Position, engageDistance)
					} else if monsterDist < engageDistance {
						targetPos = step.GetSafePositionTowardsMonster(currentPos, m.Position, engageDistance)
					} else {
						targetPos = currentPos
					}

					if targetPos != currentPos {
						if err := MoveToCoords(targetPos); err != nil {
							ctx.Logger.Warn("Failed to move to safe position",
								slog.String("error", err.Error()),
								slog.Any("monster", m.Name),
								slog.Any("position", targetPos))
							continue
						}
					}

					return m.UnitID, true
				}
			} else if monsterDist <= distance {
				// For other areas or non-seal elites, use the original logic
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
