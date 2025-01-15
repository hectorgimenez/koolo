package action

import (
	"log/slog"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/lxn/win"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
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

	SwitchStashTab(4)
	ctx.RefreshGameData()
	town.BuyConsumables(forceRefill)

	if sellJunk {
		town.SellJunk()
	}

	return step.CloseAllMenus()
}

func RestockTomes() error {
	ctx := context.Get()
	ctx.SetLastAction("RestockTomes")

	shouldBuyTPs := town.ShouldBuyTPs()
	shouldBuyIDs := town.ShouldBuyIDs()

	if !shouldBuyTPs && !shouldBuyIDs {
		return nil
	}

	ctx.Logger.Info("Visiting vendor to buy scrolls...")

	var vendorNPC npc.ID
	currentArea := ctx.Data.PlayerUnit.Area
	if currentArea == area.RogueEncampment {
		vendorNPC = npc.Akara
	} else if currentArea == area.LutGholein {
		vendorNPC = npc.Drognan
	} else if currentArea == area.KurastDocks {
		vendorNPC = npc.Ormus
	} else if currentArea == area.ThePandemoniumFortress {
		vendorNPC = npc.Jamella
	} else if currentArea == area.Harrogath {
		vendorNPC = npc.Malah
	}

	if vendorNPC == 0 {
		ctx.Logger.Info("Unable to find scroll vendor...")

		return nil
	}

	err := InteractNPC(vendorNPC)
	if err != nil {
		return err
	}

	ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)

	SwitchStashTab(4)
	ctx.RefreshGameData()

	if shouldBuyTPs {
		town.BuyTPs()
	}

	if shouldBuyIDs {
		town.BuyIDs()
	}

	return nil
}

func RestockKeys() error {
	ctx := context.Get()
	ctx.SetLastAction("RestockKeys")

	if ctx.Data.PlayerUnit.Class == data.Assassin {
		return nil
	}

	keyQuantity, needsBuy := town.ShouldBuyKeys()

	if !needsBuy {
		return nil
	}

	ctx.Logger.Info("Visiting vendor to buy keys...")

	var vendorNPC npc.ID
	currentArea := ctx.Data.PlayerUnit.Area
	if currentArea == area.RogueEncampment {
		vendorNPC = npc.Akara
	} else if currentArea == area.LutGholein {
		vendorNPC = npc.Lysander
	} else if currentArea == area.KurastDocks {
		vendorNPC = npc.Hratli
	} else if currentArea == area.ThePandemoniumFortress {
		vendorNPC = npc.Jamella
	} else if currentArea == area.Harrogath {
		vendorNPC = npc.Malah
	}

	if vendorNPC == 0 {
		ctx.Logger.Info("Unable to find keys vendor...")

		return nil
	}

	err := InteractNPC(vendorNPC)
	if err != nil {
		return err
	}

	ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)

	SwitchStashTab(4)
	ctx.RefreshGameData()

	if itm, found := ctx.Data.Inventory.Find(item.Key, item.LocationVendor); found {
		ctx.Logger.Debug("Vendor with keys detected, provisioning...")

		qty, _ := itm.FindStat(stat.Quantity, 0)
		if (qty.Value + keyQuantity) <= 12 {
			town.BuyFullStack(itm)
		}
	}

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
