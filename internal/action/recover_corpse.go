package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

func (b Builder) RecoverCorpse() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if data.Corpse.Found {
			b.logger.Info("Corpse found, let's recover our stuff...")
			steps = append(steps,
				step.SyncStep(func(data game.Data) error {
					x, y := pather.GameCoordsToScreenCords(
						data.PlayerUnit.Position.X,
						data.PlayerUnit.Position.Y,
						data.Corpse.Position.X,
						data.Corpse.Position.Y,
					)
					hid.MovePointer(x, y)
					time.Sleep(time.Millisecond * 156)
					hid.Click(hid.LeftButton)

					return nil
				}),
			)
		}

		return
	})
}
