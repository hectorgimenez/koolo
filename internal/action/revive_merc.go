package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) ReviveMerc() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		if config.Config.Character.UseMerc && data.MercHPPercent() <= 0 {
			b.logger.Info("Merc is dead, let's revive it!")

			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(data.PlayerUnit.Area).MercContractorNPC()),
				step.KeySequence("home", "down", "enter", "esc"),
			)
		}

		return
	}, Resettable(), CanBeSkipped())
}
