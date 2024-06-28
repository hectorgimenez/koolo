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

func (b *Builder) Repair() *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		for _, i := range d.Inventory.ByLocation(item.LocationEquipped) {

			// Get the durability stats
			durability, found := i.FindStat(stat.Durability, 0)
			maxDurability, maxDurabilityFound := i.FindStat(stat.MaxDurability, 0)

			// Calculate Durability percent
			durabilityPercent := (durability.Value / maxDurability.Value) * 100

			if maxDurabilityFound && !found || found && durabilityPercent <= 10 {

				b.Logger.Info(fmt.Sprintf("Repairing %s, item durability is %d percent", i.Name, durabilityPercent))

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
		}

		return nil
	})
}

func (b *Builder) RepairRequired() bool {

	gameData := b.Container.Reader.GetData(false)

	for _, i := range gameData.Inventory.ByLocation(item.LocationEquipped) {
		currentDurability, currentDurabilityFound := i.FindStat(stat.Durability, 0)
		maxDurability, maxDurabilityFound := i.FindStat(stat.MaxDurability, 0)

		// If we don't find the stats just return false
		if !currentDurabilityFound || !maxDurabilityFound {
			return false
		}

		// Let's calculate the durability percentage
		durabilityPercent := (currentDurability.Value / maxDurability.Value) * 100

		// Let's check if the item requires repair plus a few fail-safes
		if durabilityPercent <= 10 && maxDurability.Value > currentDurability.Value && !i.Ethereal {
			return true
		}
	}

	return false
}
