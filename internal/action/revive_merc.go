package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) ReviveMerc() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if b.cfg.Character.UseMerc && !data.Health.Merc.Alive {
			b.logger.Info("Merc is dead, let's revive it!")

			steps = append(steps,
				step.NewInteractNPC(town.GetTownByArea(data.Area).MercContractorNPC(), b.pf),
				step.NewKeySequence("up", "down", "enter", "esc"),
			)
		}

		return
	})
}
