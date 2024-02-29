package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) RecoverCorpse() *StepChainAction {
	return NewStepChain(func(d data.Data) (steps []step.Step) {
		b.logger.Debug("Checking for character corpse...")
		if d.Corpse.Found {
			b.logger.Info("Corpse found, let's recover our stuff...")
			steps = append(steps,
				step.SyncStepWithCheck(func(d data.Data) error {
					x, y := b.pf.GameCoordsToScreenCords(
						d.PlayerUnit.Position.X,
						d.PlayerUnit.Position.Y,
						d.Corpse.Position.X,
						d.Corpse.Position.Y,
					)
					b.hid.Click(game.LeftButton, x, y)

					return nil
				}, func(d data.Data) step.Status {
					if d.Corpse.Found {
						return step.StatusInProgress
					}

					return step.StatusCompleted
				}),
			)
		} else {
			b.logger.Debug("Character corpse not found :D")
		}

		return
	}, Resettable())
}
