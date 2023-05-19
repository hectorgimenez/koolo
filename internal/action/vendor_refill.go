package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) VendorRefill() *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		if b.shouldVisitVendor(d) {
			b.logger.Info("Visiting vendor...")

			openShopStep := step.KeySequence("home", "down", "enter")
			// Jamella trade button is the first one
			if d.PlayerUnit.Area == area.ThePandemoniumFortress {
				openShopStep = step.KeySequence("home", "enter")
			}

			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).RefillNPC()),
				openShopStep,
				step.SyncStep(func(d data.Data) error {
					// Small delay to allow the vendor window popup
					helper.Sleep(1000)

					return nil
				}),
				step.SyncStep(func(d data.Data) error {
					switchTab(4)
					b.sm.BuyConsumables(d)
					return nil
				}),
				step.SyncStep(func(d data.Data) error {
					b.sm.SellJunk(d)
					return nil
				}),
				step.KeySequence("esc"),
			)
		}

		return
	}, Resettable(), CanBeSkipped())
}

func (b Builder) shouldVisitVendor(d data.Data) bool {
	// Check if we should sell junk
	if len(nonLockedItems(d)) > 0 {
		return true
	}

	// Skip the vendor if we don't have enough gold to do anything... this is not the optimal scenario,
	// but I have no idea how to check vendor item prices.
	if d.PlayerUnit.TotalGold() < 1000 {
		return false
	}

	return b.bm.ShouldBuyPotions(d) || b.sm.ShouldBuyTPs(d) || b.sm.ShouldBuyIDs(d)
}

func nonLockedItems(d data.Data) (items []data.Item) {
	for _, item := range d.Items.Inventory {
		if config.Config.Inventory.InventoryLock[item.Position.Y][item.Position.X] == 1 {
			items = append(items, item)
		}
	}

	return
}
