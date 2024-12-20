package action

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func CubeAddItems(items ...data.Item) error {
	ctx := context.Get()
	ctx.SetLastAction("CubeAddItems")

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

	err := ensureCubeIsOpen()
	if err != nil {
		return err
	}

	err = ensureCubeIsEmpty()
	if err != nil {
		return err
	}

	for _, itm := range items {
		for _, updatedItem := range ctx.Data.Inventory.AllItems {
			if itm.UnitID == updatedItem.UnitID {
				ctx.Logger.Debug("Moving Item to the Horadric Cube", slog.String("Item", string(itm.Name)))

				screenPos := ui.GetScreenCoordsForItem(updatedItem)

				ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
				utils.Sleep(500)
			}
		}
	}

	return nil
}

func CubeTransmute() error {
	ctx := context.Get()

	err := ensureCubeIsOpen()
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

	// Take the items out of the cube
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationCube) {
		ctx.Logger.Debug("Moving Item to the inventory", slog.String("Item", string(itm.Name)))

		screenPos := ui.GetScreenCoordsForItem(itm)

		ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
		utils.Sleep(500)
	}

	return step.CloseAllMenus()
}

func ensureCubeIsEmpty() error {
	ctx := context.Get()
	if !ctx.Data.OpenMenus.Cube {
		return errors.New("horadric Cube window not detected")
	}

	cubeItems := ctx.Data.Inventory.ByLocation(item.LocationCube)
	if len(cubeItems) == 0 {
		return nil
	}

	ctx.Logger.Debug("Emptying the Horadric Cube")
	for _, itm := range cubeItems {
		ctx.Logger.Debug("Moving Item to the inventory", slog.String("Item", string(itm.Name)))

		screenPos := ui.GetScreenCoordsForItem(itm)

		ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
		utils.Sleep(700)

		itm, _ = ctx.Data.Inventory.FindByID(itm.UnitID)
		if itm.Location.LocationType == item.LocationCube {
			return fmt.Errorf("item %s could not be removed from the cube", itm.Name)
		}
	}

	ctx.HID.PressKey(win.VK_ESCAPE)
	utils.Sleep(300)

	stashInventory(true)

	return ensureCubeIsOpen()
}

func ensureCubeIsOpen() error {
	ctx := context.Get()
	ctx.Logger.Debug("Opening Horadric Cube...")

	if ctx.Data.OpenMenus.Cube {
		ctx.Logger.Debug("Horadric Cube window already open")
		return nil
	}

	cube, found := ctx.Data.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
	if !found {
		return errors.New("horadric cube not found in inventory")
	}

	// If cube is in stash, switch to the correct tab
	if cube.Location.LocationType == item.LocationStash || cube.Location.LocationType == item.LocationSharedStash {
		SwitchStashTab(cube.Location.Page + 1)
	}

	screenPos := ui.GetScreenCoordsForItem(cube)

	utils.Sleep(300)
	ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
	utils.Sleep(500)

	if ctx.Data.OpenMenus.Cube {
		ctx.Logger.Debug("Horadric Cube window detected")
		return nil
	}

	return errors.New("horadric Cube window not detected")
}
