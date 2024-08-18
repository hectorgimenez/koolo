package action

import (
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/ui"
	"github.com/hectorgimenez/koolo/internal/v2/utils"

	"github.com/hectorgimenez/koolo/internal/game"
)

func SwitchToLegacyMode() {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "SwitchToLegacyMode"

	if ctx.CharacterCfg.ClassicMode && !ctx.Data.LegacyGraphics {
		ctx.Logger.Debug("Switching to legacy mode...")
		ctx.HID.PressKey(ctx.Data.KeyBindings.LegacyToggle.Key1[0])
		utils.Sleep(500) // Add small delay to allow the game to switch

		// Close the mini panel if option is enabled
		if ctx.CharacterCfg.CloseMiniPanel {
			utils.Sleep(100)
			ctx.HID.Click(game.LeftButton, ui.CloseMiniPanelClassicX, ui.CloseMiniPanelClassicY)
			utils.Sleep(100)
		}
	}
}
