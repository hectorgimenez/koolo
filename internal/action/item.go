package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func doesExceedQuantity(rule nip.Rule) bool {
	ctx := context.Get()
	ctx.SetLastAction("doesExceedQuantity")

	stashItems := ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash)

	maxQuantity := rule.MaxQuantity()
	if maxQuantity == 0 {
		return false
	}

	if maxQuantity == 0 {
		return false
	}

	matchedItemsInStash := 0

	for _, stashItem := range stashItems {
		res, _ := rule.Evaluate(stashItem)
		if res == nip.RuleResultFullMatch {
			matchedItemsInStash += 1
		}
	}

	return matchedItemsInStash >= maxQuantity
}

func DropMouseItem() {
	ctx := context.Get()
	ctx.SetLastAction("DropMouseItem")

	if len(ctx.Data.Inventory.ByLocation(item.LocationCursor)) > 0 {
		utils.Sleep(1000)
		ctx.HID.Click(game.LeftButton, 500, 500)
		utils.Sleep(1000)
	}
}

func DropInventoryItem(i data.Item) error {
	ctx := context.Get()
	ctx.SetLastAction("DropInventoryItem")

	closeAttempts := 0

	// Check if any other menu is open, except the inventory
	for ctx.Data.OpenMenus.IsMenuOpen() {

		// Press escape to close it
		ctx.HID.PressKey(0x1B) // ESC
		utils.Sleep(500)
		closeAttempts++

		if closeAttempts >= 5 {
			return fmt.Errorf("failed to close open menu after 5 attempts")
		}
	}

	if i.Location.LocationType == item.LocationInventory {

		// Check if the inventory is open, if not open it
		if !ctx.Data.OpenMenus.Inventory {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
		}

		// Wait a second
		utils.Sleep(1000)

		screenPos := ui.GetScreenCoordsForItem(i)
		ctx.HID.MovePointer(screenPos.X, screenPos.Y)
		utils.Sleep(250)
		ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
		utils.Sleep(500)

		// Close the inventory if its still open, which should be at this point
		if ctx.Data.OpenMenus.Inventory {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
		}
	}

	return nil
}
func IsInLockedInventorySlot(itm data.Item) bool {
	// Check if item is in inventory
	if itm.Location.LocationType != item.LocationInventory {
		return false
	}

	// Get the lock configuration from character config
	ctx := context.Get()
	lockConfig := ctx.CharacterCfg.Inventory.InventoryLock
	if len(lockConfig) == 0 {
		return false
	}

	// Calculate row and column in inventory
	row := itm.Position.Y
	col := itm.Position.X

	// Check if position is within bounds
	if row >= len(lockConfig) || col >= len(lockConfig[0]) {
		return false
	}

	// 0 means locked, 1 means unlocked
	return lockConfig[row][col] == 0
}
