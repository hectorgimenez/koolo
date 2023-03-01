package action

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) ClearAreaAroundPlayer(distance int) *DynamicAction {
	originalPosition := game.Position{}
	return b.ch.KillMonsterSequence(func(data game.Data) (game.UnitID, bool) {
		if originalPosition.X == 0 && originalPosition.Y == 0 {
			originalPosition = data.PlayerUnit.Position
		}

		for _, m := range data.Monsters.Enemies() {
			d := pather.DistanceFromPoint(originalPosition, m.Position)
			if d <= distance {
				b.logger.Debug("Clearing area...", zap.Int("monsterID", int(m.Name)))
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
