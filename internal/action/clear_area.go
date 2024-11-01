package action

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
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
	ctx.SetLastAction("ClearAreaAroundPosition")
	ctx.Logger.Debug("Clearing area around position...", slog.Int("radius", radius))

	return ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies(filter) {
			distanceToTarget := pather.DistanceFromPoint(pos, m.Position)
			if ctx.Data.AreaData.IsWalkable(m.Position) && distanceToTarget <= radius {
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
