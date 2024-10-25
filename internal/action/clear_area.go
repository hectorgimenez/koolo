package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/npc"
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
			// Special case: Always allow (Vizier seal boss) even if off grid
			isVizier := m.Type == data.MonsterTypeSuperUnique && m.Name == npc.StormCaster

			// Skip monsters that are off grid
			if !isVizier && !ctx.Data.AreaData.IsInside(m.Position) {
				ctx.Logger.Debug("Skipping off-grid monster",
					slog.Any("monster", m.Name),
					slog.Any("position", m.Position))
				continue
			}

			monsterDist := pather.DistanceFromPoint(originalPosition, m.Position)
			if monsterDist <= distance {
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
