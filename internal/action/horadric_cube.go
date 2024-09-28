package action

import (
	"errors"
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func CubeAddItems(items ...data.Item) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "CubeAddItems"

	cube, found := ctx.Data.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
	if !found {
		ctx.Logger.Info("No Horadric Cube found in inventory")
		return nil
	}

	// Ensure stash is open
	if !ctx.Data.OpenMenus.Stash {
		bank, _ := ctx.Data.Objects.FindOne(object.Bank)
		err := InteractObject(bank, func() bool {
			return ctx.Data.OpenMenus.Stash
		})
		if err != nil {
			return err
		}
	}

	ctx.Logger.Info("Adding items to the Horadric Cube", slog.Any("items", items))

	// If items are on the Stash, pickup them to the inventory
	for _, itm := range items {
		nwIt := itm
		if nwIt.Location.LocationType != item.LocationStash && nwIt.Location.LocationType != item.LocationSharedStash {
			continue
		}

		// Check in which tab the item is and switch to it
		switch nwIt.Location.LocationType {
		case item.LocationStash:
			SwitchStashTab(1)
		case item.LocationSharedStash:
			SwitchStashTab(nwIt.Location.Page + 1)
		}

		ctx.Logger.Debug("Item found on the stash, picking it up", slog.String("Item", string(nwIt.Name)))
		screenPos := ui.GetScreenCoordsForItem(nwIt)

		ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
		utils.Sleep(300)
	}

	err := ensureCubeIsOpen(cube)
	if err != nil {
		return err
	}

	for _, itm := range items {
		for _, updatedItem := range ctx.Data.Inventory.AllItems {
			if itm.UnitID == updatedItem.UnitID {
				ctx.Logger.Debug("Moving Item to the Horadric Cube", slog.String("Item", string(itm.Name)))

				screenPos := ui.GetScreenCoordsForItem(updatedItem)

				ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
				utils.Sleep(300)
			}
		}
	}

	return nil
}

func CubeTransmute() error {
	ctx := context.Get()

	cube, found := ctx.Data.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
	if !found {
		ctx.Logger.Info("No Horadric Cube found in inventory")
		return nil
	}

	err := ensureCubeIsOpen(cube)
	if err != nil {
		return err
	}

	ctx.Logger.Debug("Transmuting items in the Horadric Cube")
	utils.Sleep(150)

	if ctx.Data.LegacyGraphics {
		ctx.HID.Click(game.LeftButton, ui.CubeTransmuteBtnXClassic, ui.CubeTransmuteBtnYClassic)
	} else {
		ctx.HID.Click(game.LeftButton, ui.CubeTransmuteBtnX, ui.CubeTransmuteBtnY)
	}

	utils.Sleep(2000)

	if ctx.Data.LegacyGraphics {
		ctx.HID.ClickWithModifier(game.LeftButton, ui.CubeTakeItemXClassic, ui.CubeTakeItemYClassic, game.CtrlKey)
	} else {
		ctx.HID.ClickWithModifier(game.LeftButton, ui.CubeTakeItemX, ui.CubeTakeItemY, game.CtrlKey)
	}

	utils.Sleep(300)

	return step.CloseAllMenus()
}

func ensureCubeIsOpen(cube data.Item) error {
	ctx := context.Get()
	ctx.Logger.Debug("Opening Horadric Cube...")

	// Switch to the tab
	SwitchStashTab(cube.Location.Page + 1)

	screenPos := ui.GetScreenCoordsForItem(cube)

	utils.Sleep(300)
	ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
	utils.Sleep(200)

	if ctx.Data.OpenMenus.Cube {
		ctx.Logger.Debug("Horadric Cube window detected")
		return nil
	}

	return errors.New("horadric Cube window not detected")
}
