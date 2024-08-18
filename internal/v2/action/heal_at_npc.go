package action

import (
	"github.com/hectorgimenez/koolo/internal/v2/action/step"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/town"
)

func HealAtNPC() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "HealAtNPC"

	shouldHeal := false
	if ctx.Data.PlayerUnit.HPPercent() < 80 {
		ctx.Logger.Info("Current life is %d%%, healing on NPC", ctx.Data.PlayerUnit.HPPercent())
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
