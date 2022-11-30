package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) VendorRefill() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		if b.shouldGoToVendor(data) {
			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(data.PlayerUnit.Area).RefillNPC()),
				step.KeySequence("home", "down", "enter"),
				step.SyncStep(func(data game.Data) error {
					// Small delay to allow the vendor window popup
					helper.Sleep(1000)

					return nil
				}),
				step.SyncStep(func(data game.Data) error {
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

func (b Builder) shouldGoToVendor(data game.Data) bool {
	// Check if we should sell junk
	if len(data.Items.Inventory.NonLockedItems()) > 0 {
		return true
	}

	return b.bm.ShouldBuyPotions(data) || data.Items.Inventory.ShouldBuyTPs() || data.Items.Inventory.ShouldBuyIDs()
}
