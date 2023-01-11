package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) VendorRefill() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		if b.shouldVisitVendor(data) {
			openShopStep := step.KeySequence("home", "down", "enter")
			// Jamella trade button is the first one
			if data.PlayerUnit.Area == area.ThePandemoniumFortress {
				openShopStep = step.KeySequence("home", "enter")
			}

			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(data.PlayerUnit.Area).RefillNPC()),
				openShopStep,
				step.SyncStep(func(data game.Data) error {
					// Small delay to allow the vendor window popup
					helper.Sleep(1000)

					return nil
				}),
				step.SyncStep(func(data game.Data) error {
					switchTab(4)
					b.sm.BuyConsumables(data)
					return nil
				}),
				step.SyncStep(func(data game.Data) error {
					b.sm.SellJunk(data)
					return nil
				}),
				step.KeySequence("esc"),
			)
		}

		return
	}, Resettable(), CanBeSkipped())
}

func (b Builder) shouldVisitVendor(data game.Data) bool {
	// Check if we should sell junk
	if len(data.Items.Inventory.NonLockedItems()) > 0 {
		return true
	}

	return b.bm.ShouldBuyPotions(data) || data.Items.Inventory.ShouldBuyTPs() || data.Items.Inventory.ShouldBuyIDs()
}
