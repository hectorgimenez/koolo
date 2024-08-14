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
            if !b.RepairRequired() {
                return nil
            }

            b.Logger.Info("Repair required, interacting with repair NPC")

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

            return append(actions, b.InteractNPC(repairNPC,
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
		if maxDurabilityFound && !currentDurabilityFound || durabilityPercent != -1 && currentDurabilityFound && durabilityPercent <= 20 || currentDurabilityFound && currentDurability.Value <= 2 {
			return true
		}
	}

	return false
}
