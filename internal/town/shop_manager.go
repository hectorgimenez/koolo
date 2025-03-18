package town

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func BuyConsumables(forceRefill bool) {
	ctx := context.Get()

	missingHealingPots := ctx.BeltManager.GetMissingCount(data.HealingPotion)
	missingManaPots := ctx.BeltManager.GetMissingCount(data.ManaPotion)

	ctx.Logger.Debug(fmt.Sprintf("Buying: %d Healing potions and %d Mana potions", missingHealingPots, missingManaPots))

	// We traverse the items in reverse order because vendor has the best potions at the end
	pot, found := findFirstMatch("superhealingpotion", "greaterhealingpotion", "healingpotion", "lighthealingpotion", "minorhealingpotion")
	if found && missingHealingPots > 0 {
		BuyItem(pot, missingHealingPots)
		missingHealingPots = 0
	}

	pot, found = findFirstMatch("supermanapotion", "greatermanapotion", "manapotion", "lightmanapotion", "minormanapotion")
	// In Normal greater potions are expensive as we are low level, let's keep with cheap ones
	if ctx.CharacterCfg.Game.Difficulty == "normal" {
		pot, found = findFirstMatch("manapotion", "lightmanapotion", "minormanapotion")
	}
	if found && missingManaPots > 0 {
		BuyItem(pot, missingManaPots)
		missingManaPots = 0
	}

	if ShouldBuyTPs() || forceRefill {
		if _, found := ctx.Data.Inventory.Find(item.TomeOfTownPortal, item.LocationInventory); !found {
			ctx.Logger.Info("TP Tome not found, buying one...")
			if itm, itmFound := ctx.Data.Inventory.Find(item.TomeOfTownPortal, item.LocationVendor); itmFound {
				BuyItem(itm, 1)
			}
		}
		ctx.Logger.Debug("Filling TP Tome...")
		if itm, found := ctx.Data.Inventory.Find(item.ScrollOfTownPortal, item.LocationVendor); found {
			buyFullStack(itm)
		}
	}

	if ShouldBuyIDs() || forceRefill {
		if _, found := ctx.Data.Inventory.Find(item.TomeOfIdentify, item.LocationInventory); !found {
			ctx.Logger.Info("ID Tome not found, buying one...")
			if itm, itmFound := ctx.Data.Inventory.Find(item.TomeOfIdentify, item.LocationVendor); itmFound {
				BuyItem(itm, 1)
			}
		}
		ctx.Logger.Debug("Filling IDs Tome...")
		if itm, found := ctx.Data.Inventory.Find(item.ScrollOfIdentify, item.LocationVendor); found {
			buyFullStack(itm)
		}
	}

	keyQuantity, shouldBuyKeys := ShouldBuyKeys()
	if ctx.Data.PlayerUnit.Class != data.Assassin && (shouldBuyKeys || forceRefill) {
		if itm, found := ctx.Data.Inventory.Find(item.Key, item.LocationVendor); found {
			ctx.Logger.Debug("Vendor with keys detected, provisioning...")

			qty, _ := itm.FindStat(stat.Quantity, 0)
			if (qty.Value + keyQuantity) <= 12 {
				buyFullStack(itm)
			}
		}
	}
}

func findFirstMatch(itemNames ...string) (data.Item, bool) {
	ctx := context.Get()
	for _, name := range itemNames {
		if itm, found := ctx.Data.Inventory.Find(item.Name(name), item.LocationVendor); found {
			return itm, true
		}
	}

	return data.Item{}, false
}

func ShouldBuyTPs() bool {
	portalTome, found := context.Get().Data.Inventory.Find(item.TomeOfTownPortal, item.LocationInventory)
	if !found {
		return true
	}

	qty, found := portalTome.FindStat(stat.Quantity, 0)

	return qty.Value < 15 || !found
}

func ShouldBuyIDs() bool {
	idTome, found := context.Get().Data.Inventory.Find(item.TomeOfIdentify, item.LocationInventory)
	if !found {
		return true
	}

	qty, found := idTome.FindStat(stat.Quantity, 0)

	return qty.Value < 15 || !found
}

func ShouldBuyKeys() (int, bool) {
	keys, found := context.Get().Data.Inventory.Find(item.Key, item.LocationInventory)
	if !found {
		return 0, true
	}

	qty, found := keys.FindStat(stat.Quantity, 0)
	if found && qty.Value >= 12 {
		return 12, false
	}

	return qty.Value, true
}

func SellJunk() {
	for _, i := range ItemsToBeSold() {
		if context.Get().Data.CharacterCfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 1 {
			SellItem(i)
		}
	}
}

func SellItem(i data.Item) {
	ctx := context.Get()
	screenPos := ui.GetScreenCoordsForItem(i)

	time.Sleep(500 * time.Millisecond)
	ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
	time.Sleep(500 * time.Millisecond)
	ctx.Logger.Debug(fmt.Sprintf("Item %s [%s] sold", i.Desc().Name, i.Quality.ToString()))
}

func BuyItem(i data.Item, quantity int) {
	ctx := context.Get()
	screenPos := ui.GetScreenCoordsForItem(i)

	time.Sleep(250 * time.Millisecond)
	for k := 0; k < quantity; k++ {
		ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
		time.Sleep(900 * time.Millisecond)
		ctx.Logger.Debug(fmt.Sprintf("Purchased %s [X:%d Y:%d]", i.Desc().Name, i.Position.X, i.Position.Y))
	}
}

func buyFullStack(i data.Item) {
	screenPos := ui.GetScreenCoordsForItem(i)

	context.Get().HID.ClickWithModifier(game.RightButton, screenPos.X, screenPos.Y, game.ShiftKey)
	time.Sleep(500 * time.Millisecond)
}

func ItemsToBeSold() (items []data.Item) {
	ctx := context.Get()
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if itm.IsFromQuest() {
			continue
		}

		if itm.Name == item.TomeOfTownPortal || itm.Name == item.TomeOfIdentify || itm.Name == item.Key || itm.Name == "WirtsLeg" {
			continue
		}

		if itm.IsRuneword {
			continue
		}

		if ctx.Data.CharacterCfg.Inventory.InventoryLock[itm.Position.Y][itm.Position.X] == 1 {
			// If item is a full match will be stashed, we don't want to sell it
			if _, result := ctx.Data.CharacterCfg.Runtime.Rules.EvaluateAll(itm); result == nip.RuleResultFullMatch && !itm.IsPotion() {
				continue
			}
			items = append(items, itm)
		}
	}

	return
}
