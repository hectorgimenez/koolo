package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) ClearAreaAroundPlayer(distance int) *DynamicAction {
	return BuildDynamic(func(data game.Data) ([]step.Step, bool) {
		for _, m := range data.Monsters.Enemies() {
			d := pather.DistanceFromMe(data, m.Position.X, m.Position.Y)
			if d <= distance {
				b.logger.Debug("Clearing area...", zap.Int("monsterID", int(m.Name)))
				return b.ch.KillMonsterSequence(data, m.UnitID), true
			}
		}

		return nil, false
	}, CanBeSkipped())
}
