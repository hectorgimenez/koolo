package action

import (
	"fmt"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
)

func HealAtNPC() error {
	ctx := context.Get()
	ctx.SetLastAction("HealAtNPC")

	shouldHeal := false
	if ctx.Data.PlayerUnit.HPPercent() < 80 {
		ctx.Logger.Info(fmt.Sprintf("Current life is %d, healing on NPC", ctx.Data.PlayerUnit.HPPercent()))
		shouldHeal = true
	}

	if ctx.Data.PlayerUnit.HasDebuff() {
		ctx.Logger.Info("Debuff detected, healing on NPC")
		shouldHeal = true
	}

	if shouldHeal {
		err := InteractNPC(town.GetTownByArea(ctx.Data.PlayerUnit.Area).HealNPC())
		if err != nil {
			ctx.Logger.Warn("Failed to heal on NPC: %v", err)
		}
	}

	return step.CloseAllMenus()
}
