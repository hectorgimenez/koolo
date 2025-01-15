package action

import (
	"fmt"

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

func Repair() error {
	ctx := context.Get()
	ctx.SetLastAction("Repair")

	if RepairRequired() {

		// Get the repair NPC for the town
		repairNPC := town.GetTownByArea(ctx.Data.PlayerUnit.Area).RepairNPC()

		// Act3 repair NPC handling
		if repairNPC == npc.Hratli {
			MoveToCoords(data.Position{X: 5224, Y: 5045})
		}

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
		if ctx.Data.LegacyGraphics {
			ctx.HID.Click(game.LeftButton, ui.RepairButtonXClassic, ui.RepairButtonYClassic)
		} else {
			ctx.HID.Click(game.LeftButton, ui.RepairButtonX, ui.RepairButtonY)
		}
		utils.Sleep(500)

		return step.CloseAllMenus()
	}

	return nil
}

func RepairRequired() bool {
	ctx := context.Get()
	ctx.SetLastAction("RepairRequired")

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
		// Skip indestructible items
		_, indestructible := i.FindStat(stat.Indestructible, 0)
		if i.Ethereal || indestructible {
			continue
		}

		currentDurability, currentDurabilityFound := i.FindStat(stat.Durability, 0)
		maxDurability, maxDurabilityFound := i.FindStat(stat.MaxDurability, 0)
		quantity, quantityFound := i.FindStat(stat.Quantity, 0)

		// If we have both stats, check percentage
		if currentDurabilityFound && maxDurabilityFound && !quantityFound {
			durabilityPercent := int((float64(currentDurability.Value) / float64(maxDurability.Value)) * 100)
			if durabilityPercent <= 20 {
				ctx.Logger.Info(fmt.Sprintf("Repairing %s, item durability is %d percent", i.Name, durabilityPercent))
				return true
			}
		}

		// If we only have current durability, check absolute value
		if currentDurabilityFound && !quantityFound {
			if currentDurability.Value <= 8 {
				ctx.Logger.Info(fmt.Sprintf("Repairing %s, item durability is %d", i.Name, currentDurability.Value))
				return true
			}
		}

		// Handle case where durability stat is missing but max durability exists
		// This likely indicates the item needs repair
		if maxDurabilityFound && !currentDurabilityFound && !quantityFound {
			ctx.Logger.Info(fmt.Sprintf("Repairing %s, item has max durability but no current durability", i.Name))
			return true
		}

		// If all other checks pass, look for quantity value (throwables)
		if quantityFound && quantity.Value <= 15 {
			ctx.Logger.Info(fmt.Sprintf("Repairing %s, item quantity is %d", i.Name, quantity.Value))
			return true
		}
	}

	return false
}
