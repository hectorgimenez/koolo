package action

import (
	"log/slog"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/lxn/win"

	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
)

func VendorRefill(forceRefill, sellJunk bool) error {
	ctx := context.Get()
	ctx.SetLastAction("VendorRefill")

	if !forceRefill && !shouldVisitVendor() {
		return nil
	}

	ctx.Logger.Info("Visiting vendor...", slog.Bool("forceRefill", forceRefill))

	vendorNPC := town.GetTownByArea(ctx.Data.PlayerUnit.Area).RefillNPC()
	if vendorNPC == npc.Drognan {
		_, needsBuy := town.ShouldBuyKeys()
		if needsBuy {
			vendorNPC = npc.Lysander
		}
	}
	err := InteractNPC(vendorNPC)
	if err != nil {
		return err
	}

	// Jamella trade button is the first one
	if vendorNPC == npc.Jamella {
		ctx.HID.KeySequence(win.VK_HOME, win.VK_RETURN)
	} else {
		ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
	}

	if sellJunk {
		town.SellJunk()
	}
	SwitchStashTab(4)
	ctx.RefreshGameData()
	town.BuyConsumables(forceRefill)

	return step.CloseAllMenus()
}

func BuyAtVendor(vendor npc.ID, items ...VendorItemRequest) error {
	ctx := context.Get()
	ctx.SetLastAction("BuyAtVendor")

	err := InteractNPC(vendor)
	if err != nil {
		return err
	}

	// Jamella trade button is the first one
	if vendor == npc.Jamella {
		ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
	} else {
		ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
	}

	for _, i := range items {
		SwitchStashTab(i.Tab)
		itm, found := ctx.Data.Inventory.Find(i.Item, item.LocationVendor)
		if found {
			town.BuyItem(itm, i.Quantity)
		} else {
			ctx.Logger.Warn("Item not found in vendor", slog.String("Item", string(i.Item)))
		}
	}

	return step.CloseAllMenus()
}

type VendorItemRequest struct {
	Item     item.Name
	Quantity int
	Tab      int // At this point I have no idea how to detect the Tab the Item is in the vendor (1-4)
}

func shouldVisitVendor() bool {
	ctx := context.Get()
	ctx.SetLastStep("shouldVisitVendor")

	// Check if we should sell junk
	if len(town.ItemsToBeSold()) > 0 {
		return true
	}

	// Skip the vendor if we don't have enough gold to do anything... this is not the optimal scenario,
	// but I have no idea how to check vendor Item prices.
	if ctx.Data.PlayerUnit.TotalPlayerGold() < 1000 {
		return false
	}

	return ctx.BeltManager.ShouldBuyPotions() || town.ShouldBuyTPs() || town.ShouldBuyIDs()
}
