package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) ReviveMerc() *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if config.Config.Character.UseMerc && d.MercHPPercent() <= 0 {
			if isLevelingChar && d.PlayerUnit.Area == area.RogueEncampment {
				// Ignoring because merc is not hired yet
				return []step.Step{}
			}

			b.logger.Info("Merc is dead, let's revive it!")

			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).MercContractorNPC()),
				step.KeySequence("home", "down", "enter", "esc"),
			)
		}

		return
	}, Resettable(), CanBeSkipped())
}
