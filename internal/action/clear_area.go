package action

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b *Builder) ClearAreaAroundPlayer(distance int, filter data.MonsterFilter) Action {
	originalPosition := data.Position{}
	return b.ch.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if originalPosition.X == 0 && originalPosition.Y == 0 {
			originalPosition = d.PlayerUnit.Position
		}

		for _, m := range d.Monsters.Enemies(filter) {
			monsterDist := pather.DistanceFromPoint(originalPosition, m.Position)
			shouldEngage := b.IsMonsterSealElite(m) || pather.IsWalkable(m.Position, d.AreaOrigin, d.CollisionGrid)

			if monsterDist <= distance && shouldEngage {
				b.Logger.Debug("Clearing area...", slog.Int("monsterID", int(m.Name)))
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
