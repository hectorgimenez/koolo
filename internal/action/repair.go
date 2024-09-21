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
	ctx.ContextDebug.LastAction = "Repair"

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {

		_, indestructible := i.FindStat(stat.Indestructible, 0)

		if i.Ethereal || indestructible {
			continue
		}

		// Get the durability stats

		durability, found := i.FindStat(stat.Durability, 0)
		maxDurability, maxDurabilityFound := i.FindStat(stat.MaxDurability, 0)

		// Calculate Durability percent
		durabilityPercent := -1

		if maxDurabilityFound && found {
			durabilityPercent = int((float64(durability.Value) / float64(maxDurability.Value)) * 100)
		}

		// Restructured conditionals for when to attempt repair
		if (maxDurabilityFound && !found) ||
			(durabilityPercent != -1 && found && durabilityPercent <= 20) ||
			(found && durabilityPercent == -1 && durability.Value <= 2) {

			ctx.Logger.Info(fmt.Sprintf("Repairing %s, item durability is %d percent", i.Name, durabilityPercent))

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
	}

	return nil
}

func RepairRequired() bool {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "RepairRequired"

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {

		_, indestructible := i.FindStat(stat.Indestructible, 0)

		if i.Ethereal || indestructible {
			continue
		}

		currentDurability, currentDurabilityFound := i.FindStat(stat.Durability, 0)
		maxDurability, maxDurabilityFound := i.FindStat(stat.MaxDurability, 0)

		durabilityPercent := -1

		if maxDurabilityFound && currentDurabilityFound {
			durabilityPercent = int((float64(currentDurability.Value) / float64(maxDurability.Value)) * 100)
		}

		// If we don't find the stats just continue
		if !currentDurabilityFound && !maxDurabilityFound {
			continue
		}

		// Let's check if the item requires repair plus a few fail-safes
		if maxDurabilityFound && !currentDurabilityFound || durabilityPercent != -1 && currentDurabilityFound && durabilityPercent <= 20 || currentDurabilityFound && currentDurability.Value <= 5 {
			return true
		}
	}

	return false
}
