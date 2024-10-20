package action

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func ClearAreaAroundPlayer(distance int, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ClearAreaAroundPlayer"

	originalPosition := data.Position{}

	ctx.Logger.Debug("Clearing area around character...", slog.Int("distance", distance))

	return ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if originalPosition.X == 0 && originalPosition.Y == 0 {
			originalPosition = d.PlayerUnit.Position
		}

		for _, m := range d.Monsters.Enemies(filter) {
			monsterDist := pather.DistanceFromPoint(originalPosition, m.Position)
			shouldEngage := IsMonsterSealElite(m) || d.AreaData.IsWalkable(m.Position)

			if monsterDist <= distance && shouldEngage {
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
