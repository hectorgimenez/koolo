package town

import (
	"fmt"
	"math/rand"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/ui"
	"go.uber.org/zap"
)

type ShopManager struct {
	logger *zap.Logger
	bm     health.BeltManager
}

func NewShopManager(logger *zap.Logger, bm health.BeltManager) ShopManager {
	return ShopManager{
		logger: logger,
		bm:     bm,
	}
}

func (sm ShopManager) BuyConsumables(d data.Data, forceRefill bool) {
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
	if config.Config.Game.Difficulty == "normal" {
		pot, found = sm.findFirstMatch(d, "manapotion", "lightmanapotion", "minormanapotion")
	}
	if found && missingManaPots > 0 {
		sm.BuyItem(pot, missingManaPots)
		missingManaPots = 0
	}

	if sm.ShouldBuyTPs(d) || forceRefill {
		if _, found := d.Items.Find(item.TomeOfTownPortal, item.LocationInventory); !found {
			sm.logger.Info("TP Tome not found, buying one...")
			if itm, itmFound := d.Items.Find(item.TomeOfTownPortal, item.LocationVendor); itmFound {
				sm.BuyItem(itm, 1)
			}
		}
		sm.logger.Debug("Filling TP Tome...")
		if itm, found := d.Items.Find(item.ScrollOfTownPortal, item.LocationVendor); found {
			sm.buyFullStack(itm)
		}
	}

	if sm.ShouldBuyIDs(d) || forceRefill {
		if _, found := d.Items.Find(item.TomeOfIdentify, item.LocationInventory); !found {
			sm.logger.Info("ID Tome not found, buying one...")
			if itm, itmFound := d.Items.Find(item.TomeOfIdentify, item.LocationVendor); itmFound {
				sm.BuyItem(itm, 1)
			}
		}
		sm.logger.Debug("Filling IDs Tome...")
		if itm, found := d.Items.Find(item.ScrollOfIdentify, item.LocationVendor); found {
			sm.buyFullStack(itm)
		}
	}

	if sm.ShouldBuyKeys(d) || forceRefill {
		if itm, found := d.Items.Find(item.Key, item.LocationVendor); found {
			sm.logger.Debug("Vendor with keys detected, provisioning...")
			sm.buyFullStack(itm)
		}
	}
}

func (sm ShopManager) findFirstMatch(d data.Data, itemNames ...string) (data.Item, bool) {
	for _, name := range itemNames {
		if itm, found := d.Items.Find(item.Name(name), item.LocationVendor); found {
			return itm, true
		}
	}

	return data.Item{}, false
}

func (sm ShopManager) ShouldBuyTPs(d data.Data) bool {
	portalTome, found := d.Items.Find(item.TomeOfTownPortal, item.LocationInventory)
	if !found {
		return true
	}

	qty, found := portalTome.Stats[stat.Quantity]

	return qty.Value <= rand.Intn(5-1)+1 || !found
}

func (sm ShopManager) ShouldBuyIDs(d data.Data) bool {
	idTome, found := d.Items.Find(item.TomeOfIdentify, item.LocationInventory)
	if !found {
		return true
	}

	qty, found := idTome.Stats[stat.Quantity]

	return qty.Value <= rand.Intn(7-3)+1 || !found
}

func (sm ShopManager) ShouldBuyKeys(d data.Data) bool {
	keys, found := d.Items.Find(item.Key, item.LocationInventory)
	if !found {
		return false
	}

	qty, found := keys.Stats[stat.Quantity]
	if found && qty.Value == 12 {
		return false
	}

	return true
}

func (sm ShopManager) SellJunk(d data.Data) {
	for _, i := range ItemsToBeSold(d) {
		if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 1 {
			sm.SellItem(i)
		}
	}
}

func (sm ShopManager) SellItem(i data.Item) {
	screenPos := ui.GetScreenCoordsForItem(i)
	hid.MovePointer(screenPos.X, screenPos.Y)
	helper.Sleep(100)
	hid.KeyDown("control")
	helper.Sleep(50)
	hid.Click(hid.LeftButton)
	helper.Sleep(150)
	hid.KeyUp("control")
	helper.Sleep(500)
	sm.logger.Debug(fmt.Sprintf("Item %s [%d] sold", i.Name, i.Quality))
}

func (sm ShopManager) BuyItem(i data.Item, quantity int) {
	screenPos := ui.GetScreenCoordsForItem(i)
	hid.MovePointer(screenPos.X, screenPos.Y)
	helper.Sleep(250)
	for k := 0; k < quantity; k++ {
		hid.Click(hid.RightButton)
		helper.Sleep(900)
		sm.logger.Debug(fmt.Sprintf("Purchased %s [X:%d Y:%d]", i.Name, i.Position.X, i.Position.Y))
	}
}

func (sm ShopManager) buyFullStack(i data.Item) {
	screenPos := ui.GetScreenCoordsForItem(i)
	hid.MovePointer(screenPos.X, screenPos.Y)
	helper.Sleep(250)
	hid.KeyDown("shift")
	helper.Sleep(100)
	hid.Click(hid.RightButton)
	helper.Sleep(300)
	hid.KeyUp("shift")
	helper.Sleep(500)
}

func ItemsToBeSold(d data.Data) (items []data.Item) {
	for _, itm := range d.Items.ByLocation(item.LocationInventory) {
		if itm.IsFromQuest() {
			continue
		}

		if config.Config.Inventory.InventoryLock[itm.Position.Y][itm.Position.X] == 1 {
			items = append(items, itm)
		}
	}

	return
}
