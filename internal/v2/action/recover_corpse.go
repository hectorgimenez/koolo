package action

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/ui"
)

func RecoverCorpse() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "RecoverCorpse"

	if ctx.Data.Corpse.Found {
		ctx.Logger.Info("Corpse found, let's recover our stuff...")
		x, y := ui.GameCoordsToScreenCords(
			ctx.Data.Corpse.Position.X,
			ctx.Data.Corpse.Position.Y,
		)
		ctx.HID.Click(game.LeftButton, x, y)
	}

	return nil
}
