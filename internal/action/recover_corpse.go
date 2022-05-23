package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b Builder) RecoverCorpse() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		if data.Corpse.Found {
			b.logger.Info("Corpse found, let's recover our stuff...")
			steps = append(steps,
				step.SyncStepWithCheck(func(data game.Data) error {
					x, y := pather.GameCoordsToScreenCords(
						data.PlayerUnit.Position.X,
						data.PlayerUnit.Position.Y,
						data.Corpse.Position.X,
						data.Corpse.Position.Y,
					)
					hid.MovePointer(x, y)
					helper.Sleep(300)
					hid.Click(hid.LeftButton)

					return nil
				}, func(data game.Data) step.Status {
					if data.Corpse.Found {
						return step.StatusInProgress
					}

					return step.StatusCompleted
				}),
			)
		}

		return
	}, Resettable())
}
