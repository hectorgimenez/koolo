package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) VendorRefill() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		shouldBuyTPs := data.Items.Inventory.ShouldBuyTPs()

		if b.bm.ShouldBuyPotions(data) || shouldBuyTPs {
			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(data.Area).RefillNPC()),
				step.KeySequence("up", "down", "enter"),
				step.SyncStep(func(data game.Data) error {
					b.sm.BuyPotsAndTPs(data)
					return nil
				}),
				step.KeySequence("esc"),
			)
		}

		return
	})
}
