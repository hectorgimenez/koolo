package action

import (
	"fmt"

	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/lxn/win"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b *Builder) Heal() *Chain {
	return NewChain(func(d game.Data) []Action {
		shouldHeal := false
		if d.PlayerUnit.HPPercent() < 80 {
			b.Logger.Info(fmt.Sprintf("Current life is %d%%, healing on NPC", d.PlayerUnit.HPPercent()))
			shouldHeal = true
		}

		if d.PlayerUnit.HasDebuff() {
			b.Logger.Info("Debuff detected, healing on NPC")
			shouldHeal = true
		}

		if shouldHeal {
			return []Action{b.InteractNPC(
				town.GetTownByArea(d.PlayerUnit.Area).HealNPC(),
				step.SyncStep(func(d game.Data) error {
					helper.Sleep(300)
					b.HID.PressKey(win.VK_ESCAPE)
					helper.Sleep(100)
					return nil
				}),
			)}
		}

		return nil
	})
}
