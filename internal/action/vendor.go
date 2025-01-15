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

func openTradeWindow(vendorNPC npc.ID) error {
	ctx := context.Get()

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

	return nil
}

// Act 1 Vendors:
//  Potions: Akara
//  Keys: Akara
//  Scrolls: Akara
//  Arrows/Bolts: Charsi

// Act 2 Vendors:
//  Potions: Lysander
//  Keys: Lysander
//  Scrolls: Drognan
//  Arrows/Bolts: Fara

// Act 3 Vendors:
//  Potions: Ormus
//  Keys: Hratli
//  Scrolls: Ormus
//  Arrows/Bolts: Hratli

// Act 4 Vendors:
//  Potions: Jamella
//  Keys: Jamella
//  Scrolls: Jamella
//  Arrows/Bolts: Halbu

// Act 5 Vendors:
//  Potions: Malah
//  Keys: Malah
//  Scrolls: Malah
//  Arrows/Bolts: Larzuk

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

	err := openTradeWindow(vendorNPC)
	if err != nil {
		return err
	}

	town.BuyConsumables(forceRefill)

	if sellJunk {
		town.SellJunk()
	}

	// At this point we are guaranteed to have purchased potions, as the selected vendorNPC will always have these.
	// Depending on the act, we may still need keys or scrolls.

	if town.ShouldBuyTPs() || town.ShouldBuyIDs() {
		restockTomes()
	}

	if ctx.Data.PlayerUnit.Class != data.Assassin {
		_, shouldBuyKeys := town.ShouldBuyKeys()
		if shouldBuyKeys {
			restockKeys()
		}
	}

	return step.CloseAllMenus()
}

func restockTomes() error {
	ctx := context.Get()

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

	err := openTradeWindow(vendorNPC)
	if err != nil {
		return err
	}

	if shouldBuyTPs {
		town.BuyTPs()
	}

	if shouldBuyIDs {
		town.BuyIDs()
	}

	return nil
}

func restockKeys() error {
	ctx := context.Get()

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

	err := openTradeWindow(vendorNPC)
	if err != nil {
		ctx.Logger.Info("Unable to interact with keys vendor", "error", err)
		return err
	}

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

	shouldVisit := ctx.BeltManager.ShouldBuyPotions() || town.ShouldBuyTPs() || town.ShouldBuyIDs()
	if shouldVisit {
		return true
	}

	_, shouldBuyKeys := town.ShouldBuyKeys()
	return shouldBuyKeys
}
