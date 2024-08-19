package action

import (
	"errors"

	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/ui"
	"github.com/hectorgimenez/koolo/internal/v2/utils"
)

func RecoverCorpse() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "RecoverCorpse"

	if ctx.Data.Corpse.Found {
		ctx.Logger.Info("Corpse found, let's recover our stuff...")

		attempts := 0
		for ctx.Data.Corpse.Found && attempts < 15 {
			utils.Sleep(500)
			x, y := ui.GameCoordsToScreenCords(
				ctx.Data.Corpse.Position.X,
				ctx.Data.Corpse.Position.Y,
			)
			ctx.HID.Click(game.LeftButton, x, y)
			attempts++
		}
		if ctx.Data.Corpse.Found {
			return errors.New("could not recover corpse")
		}
	}

	return nil
}
