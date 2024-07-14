package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
)

//@TODO Remove usage of WeaponsUtils when the d2go project is updated to manage throwables
func (b *Builder) Repair() *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		shouldRepair := false
		for _, i := range d.Inventory.ByLocation(item.LocationEquipped) {
			if i.Ethereal {
				continue
			}
			if game.WeaponUtils.IsItemThrowable(i) {
				quantity, qtyFound := i.FindStat(stat.Quantity, 0)
				if qtyFound && quantity.Value <= game.WeaponUtils.GetThrowableMaxQuantity(i)/3 || quantity.Value < 20 {
					b.Logger.Info(fmt.Sprintf("Repairing %s, item quantity is at %d", i.Name, quantity.Value))
					shouldRepair = true
					continue
				}
			}

			// Get the durability stats
			durability, found := i.FindStat(stat.Durability, 0)
			maxDurability, maxDurabilityFound := i.FindStat(stat.MaxDurability, 0)

			// Calculate Durability percent
			durabilityPercent := -1

			if maxDurabilityFound && found {
				durabilityPercent = int((float64(durability.Value) / float64(maxDurability.Value)) * 100)
			}

			if maxDurabilityFound && !found || durabilityPercent != -1 && found && durabilityPercent <= 20 || found && durability.Value <= 5 {

				b.Logger.Info(fmt.Sprintf("Repairing %s, item durability is %d percent", i.Name, durabilityPercent))
				shouldRepair = true
			}
		}

		if shouldRepair {
			// Get the repair NPC for the town
			repairNPC := town.GetTownByArea(d.PlayerUnit.Area).RepairNPC()

			// Act3 repair NPC handling
			if repairNPC == npc.Hratli {
				actions = append(actions, b.MoveToCoords(data.Position{X: 5224, Y: 5045}))
			}

			keys := make([]byte, 0)
			keys = append(keys, win.VK_HOME)

			if repairNPC != npc.Halbu {
				keys = append(keys, win.VK_DOWN)
			}

			keys = append(keys, win.VK_RETURN)

			return append(actions, b.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).RepairNPC(),
				step.KeySequence(keys...),
				step.SyncStep(func(_ game.Data) error {
					helper.Sleep(100)
					if d.LegacyGraphics {
						b.HID.Click(game.LeftButton, ui.RepairButtonXClassic, ui.RepairButtonYClassic)
					} else {
						b.HID.Click(game.LeftButton, ui.RepairButtonX, ui.RepairButtonY)
					}
					helper.Sleep(500)
					return nil
				}),
				step.KeySequence(win.VK_ESCAPE),
			))
		}

		return nil
	})
}

func (b *Builder) RepairRequired() bool {

	gameData := b.Container.Reader.GetData(false)

	for _, i := range gameData.Inventory.ByLocation(item.LocationEquipped) {
		if i.Ethereal {
			continue
		}
		if game.WeaponUtils.IsItemThrowable(i) {
			quantity, qtyFound := i.FindStat(stat.Quantity, 0)
			if qtyFound && quantity.Value <= game.WeaponUtils.GetThrowableMaxQuantity(i)/3 || quantity.Value < 20 {
				b.Logger.Info(fmt.Sprintf("Repairing %s, item quantity is at %d", i.Name, quantity.Value))
				return true
			}
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
