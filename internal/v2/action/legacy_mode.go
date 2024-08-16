package action

import (
	"time"

	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/v2/context"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

func SwitchToLegacyMode() {
	ctx := context.Get()

	if ctx.CharacterCfg.ClassicMode && !ctx.Data.LegacyGraphics {
		ctx.Logger.Debug("Switching to legacy mode...")
		ctx.HID.PressKey(ctx.Data.KeyBindings.LegacyToggle.Key1[0])
		step.Wait(time.Millisecond * 500) // Add small delay to allow the game to switch

		// Close the mini panel if option is enabled
		if ctx.CharacterCfg.CloseMiniPanel {
			helper.Sleep(100)
			ctx.HID.Click(game.LeftButton, ui.CloseMiniPanelClassicX, ui.CloseMiniPanelClassicY)
			helper.Sleep(100)
		}
	}
}
