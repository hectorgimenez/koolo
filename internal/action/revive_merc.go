package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/lxn/win"
)

func (b *Builder) ReviveMerc() *Chain {
	return NewChain(func(d game.Data) []Action {
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if d.CharacterCfg.Character.UseMerc && d.MercHPPercent() <= 0 {
			if isLevelingChar && d.PlayerUnit.Area == area.RogueEncampment && d.CharacterCfg.Game.Difficulty == difficulty.Normal {
				// Ignoring because merc is not hired yet
				return nil
			}

			b.Logger.Info("Merc is dead, let's revive it!")

			mercNPC := town.GetTownByArea(d.PlayerUnit.Area).MercContractorNPC()

			keySequence := []byte{win.VK_HOME, win.VK_DOWN, win.VK_RETURN, win.VK_ESCAPE}
			if mercNPC == npc.Tyrael2 {
				keySequence = []byte{win.VK_END, win.VK_UP, win.VK_RETURN, win.VK_ESCAPE}
			}

			return []Action{
				b.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).MercContractorNPC(),
					step.KeySequence(keySequence...),
				),
			}
		}

		return nil
	})
}
