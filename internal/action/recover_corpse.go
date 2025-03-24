package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func RecoverCorpse() error {
	ctx := context.Get()
	ctx.SetLastAction("RecoverCorpse")

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

		for _, i := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
			if i.IsPotion() {

				// Open inventory if it's not already open
				for !ctx.Data.OpenMenus.Inventory {
					ctx.Logger.Debug("Opening inventory to put potions back in belt")
					ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
					utils.Sleep(200) // Add small delay to allow the game to open the inventory
				}

				screenPos := ui.GetScreenCoordsForItem(i)
				ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.ShiftKey)
				utils.Sleep(250)
			}
		}

		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.ClearScreen)

	}
	return nil
}
