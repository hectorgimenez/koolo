package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) VendorRefill() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if b.shouldGoToVendor(data) {
			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(data.Area).RefillNPC()),
				step.KeySequence("up", "down", "enter"),
				step.SyncStep(func(data game.Data) error {
					b.sm.BuyConsumables(data)
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
