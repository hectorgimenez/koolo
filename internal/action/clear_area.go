package action

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) ClearAreaAroundPlayer(distance int) *DynamicAction {
	return b.ch.KillMonsterSequence(func(data game.Data) (game.UnitID, bool) {
		for _, m := range data.Monsters.Enemies() {
			d := pather.DistanceFromMe(data, m.Position.X, m.Position.Y)
			if d <= distance {
				b.logger.Debug("Clearing area...", zap.Int("monsterID", int(m.Name)))
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
