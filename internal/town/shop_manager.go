package town

import (
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type ShopManager struct {
	logger    *slog.Logger
	bm        health.BeltManager
	container container.Container
}

func NewShopManager(logger *slog.Logger, bm health.BeltManager, container container.Container) ShopManager {
	return ShopManager{
		logger:    logger,
		bm:        bm,
		container: container,
	}
}

func (sm ShopManager) BuyConsumables(d game.Data, forceRefill bool) {
	missingHealingPots := sm.bm.GetMissingCount(d, data.HealingPotion)
	missingManaPots := sm.bm.GetMissingCount(d, data.ManaPotion)

	sm.logger.Debug(fmt.Sprintf("Buying: %d Healing potions and %d Mana potions", missingHealingPots, missingManaPots))

	// We traverse the items in reverse order because vendor has the best potions at the end
	pot, found := sm.findFirstMatch(d, "superhealingpotion", "greaterhealingpotion", "healingpotion", "lighthealingpotion", "minorhealingpotion")
	if found && missingHealingPots > 0 {
		sm.BuyItem(pot, missingHealingPots)
		missingHealingPots = 0
	}

	pot, found = sm.findFirstMatch(d, "supermanapotion", "greatermanapotion", "manapotion", "lightmanapotion", "minormanapotion")
	// In Normal greater potions are expensive as we are low level, let's keep with cheap ones
	if d.CharacterCfg.Game.Difficulty == "normal" {
		pot, found = sm.findFirstMatch(d, "manapotion", "lightmanapotion", "minormanapotion")
	}
	if found && missingManaPots > 0 {
		sm.BuyItem(pot, missingManaPots)
		missingManaPots = 0
	}

	if sm.ShouldBuyTPs(d) || forceRefill {
		if _, found := d.Inventory.Find(item.TomeOfTownPortal, item.LocationInventory); !found {
			sm.logger.Info("TP Tome not found, buying one...")
			if itm, itmFound := d.Inventory.Find(item.TomeOfTownPortal, item.LocationVendor); itmFound {
				sm.BuyItem(itm, 1)
			}
		}
		sm.logger.Debug("Filling TP Tome...")
		if itm, found := d.Inventory.Find(item.ScrollOfTownPortal, item.LocationVendor); found {
			sm.buyFullStack(itm)
		}
	}

	if sm.ShouldBuyIDs(d) || forceRefill {
		if _, found := d.Inventory.Find(item.TomeOfIdentify, item.LocationInventory); !found {
			sm.logger.Info("ID Tome not found, buying one...")
			if itm, itmFound := d.Inventory.Find(item.TomeOfIdentify, item.LocationVendor); itmFound {
				sm.BuyItem(itm, 1)
			}
		}
		sm.logger.Debug("Filling IDs Tome...")
		if itm, found := d.Inventory.Find(item.ScrollOfIdentify, item.LocationVendor); found {
			sm.buyFullStack(itm)
		}
	}

	keyQuantity, shouldBuyKeys := sm.ShouldBuyKeys(d)
	if d.CharacterCfg.Character.Class != "mosaic" && (shouldBuyKeys || forceRefill) {
		if itm, found := d.Inventory.Find(item.Key, item.LocationVendor); found {
			sm.logger.Debug("Vendor with keys detected, provisioning...")
			qty, _ := itm.FindStat(stat.Quantity, 0)
			if (qty.Value + keyQuantity) <= 12 {
				sm.buyFullStack(itm)
			}
		}
	}
}

func (sm ShopManager) findFirstMatch(d game.Data, itemNames ...string) (data.Item, bool) {
	for _, name := range itemNames {
		if itm, found := d.Inventory.Find(item.Name(name), item.LocationVendor); found {
			return itm, true
		}
	}

	return data.Item{}, false
}

func (sm ShopManager) ShouldBuyTPs(d game.Data) bool {
	portalTome, found := d.Inventory.Find(item.TomeOfTownPortal, item.LocationInventory)
	if !found {
		return true
	}

	qty, found := portalTome.FindStat(stat.Quantity, 0)

	return qty.Value <= rand.Intn(5-1)+1 || !found
}

func (sm ShopManager) ShouldBuyIDs(d game.Data) bool {
	idTome, found := d.Inventory.Find(item.TomeOfIdentify, item.LocationInventory)
	if !found {
		return true
	}

	qty, found := idTome.FindStat(stat.Quantity, 0)

	return qty.Value <= rand.Intn(7-3)+1 || !found
}

func (sm ShopManager) ShouldBuyKeys(d game.Data) (int, bool) {
	keys, found := d.Inventory.Find(item.Key, item.LocationInventory)
	if !found {
		return 12, false
	}

	qty, found := keys.FindStat(stat.Quantity, 0)
	if found && qty.Value >= 12 {
		return 12, false
	}

	return qty.Value, true
}

func (sm ShopManager) SellJunk(d game.Data) {
	for _, i := range ItemsToBeSold(d) {
		if d.CharacterCfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 1 {
			sm.SellItem(i)
		}
	}
}

func (sm ShopManager) SellItem(i data.Item) {
	screenPos := sm.container.UIManager.GetScreenCoordsForItem(i)

	helper.Sleep(500)
	sm.container.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
	helper.Sleep(500)
	sm.logger.Debug(fmt.Sprintf("Item %s [%s] sold", i.Desc().Name, i.Quality.ToString()))
}

func (sm ShopManager) BuyItem(i data.Item, quantity int) {
	screenPos := sm.container.UIManager.GetScreenCoordsForItem(i)

	helper.Sleep(250)
	for k := 0; k < quantity; k++ {
		sm.container.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
		helper.Sleep(900)
		sm.logger.Debug(fmt.Sprintf("Purchased %s [X:%d Y:%d]", i.Desc().Name, i.Position.X, i.Position.Y))
	}
}

func (sm ShopManager) buyFullStack(i data.Item) {
	screenPos := sm.container.UIManager.GetScreenCoordsForItem(i)

	sm.container.HID.ClickWithModifier(game.RightButton, screenPos.X, screenPos.Y, game.ShiftKey)
	helper.Sleep(500)
}

func ItemsToBeSold(d game.Data) (items []data.Item) {
	for _, itm := range d.Inventory.ByLocation(item.LocationInventory) {
		if itm.IsFromQuest() {
			continue
		}

		if itm.Name == item.TomeOfTownPortal || itm.Name == item.TomeOfIdentify || itm.Name == item.Key || itm.Name == "WirtsLeg" {
			continue
		}

		if itm.IsRuneword {
			continue
		}

		if d.CharacterCfg.Inventory.InventoryLock[itm.Position.Y][itm.Position.X] == 1 {
			// If item is a full match will be stashed, we don't want to sell it
			if _, result := d.CharacterCfg.Runtime.Rules.EvaluateAll(itm); result == nip.RuleResultFullMatch && !itm.IsPotion() {
				continue
			}
			items = append(items, itm)
		}
	}

	return
}
