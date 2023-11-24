package action

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
)

func (b *Builder) VendorRefill(forceRefill, sellJunk bool) *Chain {
	return NewChain(func(d data.Data) []Action {
		if !forceRefill && !b.shouldVisitVendor(d) {
			return nil
		}

		b.logger.Info("Visiting vendor...", zap.Bool("forceRefill", forceRefill))

		openShopStep := step.KeySequence("home", "down", "enter")
		vendorNPC := town.GetTownByArea(d.PlayerUnit.Area).RefillNPC()

		// Jamella trade button is the first one
		if vendorNPC == npc.Jamella {
			openShopStep = step.KeySequence("home", "enter")
		}

		if vendorNPC == npc.Drognan {
			if b.sm.ShouldBuyKeys(d) {
				vendorNPC = npc.Lysander
			}
		}

		return []Action{b.InteractNPC(vendorNPC,
			openShopStep,
			step.Wait(time.Second),
			step.SyncStep(func(d data.Data) error {
				switchTab(4)
				b.sm.BuyConsumables(d, forceRefill)
				return nil
			}),
			step.SyncStep(func(d data.Data) error {
				if sellJunk {
					b.sm.SellJunk(d)
				}
				return nil
			}),
			step.Wait(time.Second),
			step.KeySequence("esc"),
		)}
	})
}

func (b *Builder) BuyAtVendor(vendor npc.ID, items ...VendorItemRequest) *Chain {
	return NewChain(func(d data.Data) []Action {
		openShopStep := step.KeySequence("home", "down", "enter")

		// Jamella trade button is the first one
		if vendor == npc.Jamella {
			openShopStep = step.KeySequence("home", "enter")
		}

		return []Action{b.InteractNPC(vendor,
			openShopStep,
			step.Wait(time.Second),
			step.SyncStep(func(d data.Data) error {
				for _, i := range items {
					switchTab(i.Tab)
					itm, found := d.Items.Find(i.Item, item.LocationVendor)
					if found {
						b.sm.BuyItem(itm, i.Quantity)
					} else {
						b.logger.Warn("Item not found in vendor", zap.String("Item", string(i.Item)))
					}
				}

				return nil
			}),
			step.Wait(time.Second),
			step.KeySequence("esc"),
		)}
	})
}

type VendorItemRequest struct {
	Item     item.Name
	Quantity int
	Tab      int // At this point I have no idea how to detect the Tab the Item is in the vendor (1-4)
}

func (b *Builder) shouldVisitVendor(d data.Data) bool {
	// Check if we should sell junk
	if len(town.ItemsToBeSold(d)) > 0 {
		return true
	}

	// Skip the vendor if we don't have enough gold to do anything... this is not the optimal scenario,
	// but I have no idea how to check vendor Item prices.
	if d.PlayerUnit.TotalGold() < 1000 {
		return false
	}

	return b.bm.ShouldBuyPotions(d) || b.sm.ShouldBuyTPs(d) || b.sm.ShouldBuyIDs(d)
}
