package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func RestockArrowsOrBolts() error {
	ctx := context.Get()
	ctx.SetLastAction("RestockArrowsOrBolts")

	if !RestockArrowsOrBoltsRequired() {
		return nil
	}

	ctx.Logger.Info("Attempting to move to repair NPC to buy arrows")

	// Get the repair NPC for the town
	repairNPC := town.GetTownByArea(ctx.Data.PlayerUnit.Area).RepairNPC()

	if repairNPC == npc.Hratli { // Act 3
		MoveToCoords(data.Position{X: 5224, Y: 5045})
	} else if repairNPC == npc.Larzuk { // Act 5
		// TODO: Is this necessary?
		MoveToCoords(data.Position{X: 5143, Y: 5041})
	}

	ctx.Logger.Info("Interacting with repair NPC to buy arrows")

	err := InteractNPC(repairNPC)
	if err != nil {
		return err
	}

	if repairNPC != npc.Halbu {
		ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
	} else {
		ctx.HID.KeySequence(win.VK_HOME, win.VK_RETURN)
	}

	utils.Sleep(100)

	if repairNPC == npc.Larzuk || repairNPC == npc.Charsi {
		utils.Sleep(2000)
		// Switch to Misc tab
		ctx.HID.Click(game.LeftButton, ui.VendorTab4X, ui.VendorTab4Y)
		utils.Sleep(500)
	}

	town.BuyArrows()

	return step.CloseAllMenus()
}

func RestockArrowsOrBoltsRequired() bool {
	ctx := context.Get()

	bowFound := false

	// Are they using a bow?
	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
		if i.Type().IsType(item.TypeBow) || i.Type().IsType(item.TypeAmazonBow) {
			bowFound = true
			break
		}
	}

	if !bowFound {
		return false
	}

	arrowsEquipped, arrowsEquippedFound := ctx.Data.Inventory.Find("Arrows", item.LocationEquipped)

	// We have arrows equipped, do we have enough?
	if arrowsEquippedFound {
		qtyEquipped, found := arrowsEquipped.FindStat(stat.Quantity, 0)

		// We have no arrows equipped, let's buy some
		if !found {
			return true
		}

		if qtyEquipped.Value < 20 {
			arrowsInventory, arrowsInventoryFound := ctx.Data.Inventory.Find("Arrows", item.LocationInventory)

			// We have less than 20 equipped and none in inventory, go buy more!
			if !arrowsInventoryFound {
				ctx.Logger.Info("No arrows in inventory, buying more...")

				return true
			}

			qtyInventory, _ := arrowsInventory.FindStat(stat.Quantity, 0)

			// We have less than 20 arrows in inventory, let's buy more
			if qtyInventory.Value < 20 {
				ctx.Logger.Info("Less than 20 arrows in inventory, buying more...")

				return true
			}
		}

		return false
	} else {
		ctx.Logger.Info("No arrows equipped, buying some...")

		return true
	}
}
