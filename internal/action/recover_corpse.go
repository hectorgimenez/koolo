package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) RecoverCorpse() *StepChainAction {
	return NewStepChain(func(d game.Data) (steps []step.Step) {
		b.Logger.Debug("Checking for character corpse...")
		if d.Corpse.Found {
			b.Logger.Info("Corpse found, let's recover our stuff...")
			steps = append(steps,
				step.SyncStepWithCheck(func(d game.Data) error {
					x, y := b.PathFinder.GameCoordsToScreenCords(
						d.PlayerUnit.Position.X,
						d.PlayerUnit.Position.Y,
						d.Corpse.Position.X,
						d.Corpse.Position.Y,
					)
					b.HID.Click(game.LeftButton, x, y)

					return nil
				}, func(d game.Data) step.Status {
					if d.Corpse.Found {
						return step.StatusInProgress
					}

					return step.StatusCompleted
				}),
			)
		} else {
			b.Logger.Debug("Character corpse not found :D")
		}

		return
	}, Resettable())
}
