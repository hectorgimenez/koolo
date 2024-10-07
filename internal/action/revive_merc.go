package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/lxn/win"
)

func ReviveMerc() {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ReviveMerc"

	_, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if ctx.CharacterCfg.Character.UseMerc && ctx.Data.MercHPPercent() <= 0 {
		if isLevelingChar && ctx.Data.PlayerUnit.Area == area.RogueEncampment && ctx.CharacterCfg.Game.Difficulty == difficulty.Normal {
			// Ignoring because merc is not hired yet
			return
		}

		ctx.Logger.Info("Merc is dead, let's revive it!")

		mercNPC := town.GetTownByArea(ctx.Data.PlayerUnit.Area).MercContractorNPC()
		InteractNPC(mercNPC)

		if mercNPC == npc.Tyrael2 {
			ctx.HID.KeySequence(win.VK_END, win.VK_UP, win.VK_RETURN, win.VK_ESCAPE)
		} else {
			ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN, win.VK_ESCAPE)
		}
	}
}
