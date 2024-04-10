package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"log/slog"
)

func (b *Builder) ClearAreaAroundPlayer(distance int) Action {
	originalPosition := data.Position{}
	return b.ch.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if originalPosition.X == 0 && originalPosition.Y == 0 {
			originalPosition = d.PlayerUnit.Position
		}

		for _, m := range d.Monsters.Enemies() {
			d := pather.DistanceFromPoint(originalPosition, m.Position)
			if d <= distance {
				b.Logger.Debug("Clearing area...", slog.Int("monsterID", int(m.Name)))
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}
