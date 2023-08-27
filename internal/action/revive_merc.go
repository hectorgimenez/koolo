package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) ReviveMerc() *Chain {
	return NewChain(func(d data.Data) []Action {
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if config.Config.Character.UseMerc && d.MercHPPercent() <= 0 {
			if isLevelingChar && d.PlayerUnit.Area == area.RogueEncampment {
				// Ignoring because merc is not hired yet
				return nil
			}

			b.logger.Info("Merc is dead, let's revive it!")

			return []Action{
				b.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).MercContractorNPC(),
					step.KeySequence("home", "down", "enter", "esc"),
				),
			}
		}

		return nil
	})
}
