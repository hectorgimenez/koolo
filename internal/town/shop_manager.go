package town

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
)

const (
	topCornerVendorWindowX = 109
	topCornerVendorWindowY = 147
	ItemBoxSize            = 33
	InventoryTopLeftX      = 846
	InventoryTopLeftY      = 369
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

func (sm ShopManager) BuyConsumables(d data.Data) {
	missingHealingPots := sm.bm.GetMissingCount(d, data.HealingPotion)
	missingManaPots := sm.bm.GetMissingCount(d, data.ManaPotion)

	sm.logger.Debug(fmt.Sprintf("Buying: %d Healing potions and %d Mana potions", missingHealingPots, missingManaPots))

	for _, i := range d.Items.Shop {
		if strings.Contains(string(i.Name), "HealingPotion") && i.IsVendor && missingHealingPots > 1 {
			sm.buyItem(i, missingHealingPots)
			missingHealingPots = 0
			break
		}
	}
	for _, i := range d.Items.Shop {
		if strings.Contains(string(i.Name), "ManaPotion") && i.IsVendor && missingManaPots > 1 {
			sm.buyItem(i, missingManaPots)
			missingManaPots = 0
			break
		}
	}

	if sm.ShouldBuyTPs(d) {
		sm.logger.Debug("Filling TP Tome...")
		for _, i := range d.Items.Shop {
			if i.Name == item.ScrollOfTownPortal && i.IsVendor {
				sm.buyFullStack(i)
				break
			}
		}
	}

	if sm.ShouldBuyIDs(d) {
		sm.logger.Debug("Filling IDs Tome...")
		for _, i := range d.Items.Shop {
			if i.Name == item.ScrollOfIdentify && i.IsVendor {
				sm.buyFullStack(i)
				break
			}
		}
	}

	for _, i := range d.Items.Shop {
		if string(i.Name) == "Key" && i.IsVendor {
			if sm.ShouldBuyKeys(d) {
				sm.logger.Debug("Vendor with keys detected, provisioning...")
				sm.buyFullStack(i)
				break
			}
		}
	}
}

func (sm ShopManager) ShouldBuyTPs(d data.Data) bool {
	for _, it := range d.Items.Inventory {
		if it.Name != item.TomeOfTownPortal {
			continue
		}

		qty, found := it.Stats[stat.Quantity]
		if qty.Value <= rand.Intn(5-1)+1 || !found {
			return true
		}
	}
	return false
}

func (sm ShopManager) ShouldBuyIDs(d data.Data) bool {
	for _, it := range d.Items.Inventory {
		if it.Name != item.TomeOfIdentify {
			continue
		}

		qty, found := it.Stats[stat.Quantity]
		if qty.Value <= rand.Intn(7-3)+1 || !found {
			return true
		}
	}
	return false
}

func (sm ShopManager) ShouldBuyKeys(d data.Data) bool {
	for _, it := range d.Items.Inventory {
		if it.Name != item.Key {
			continue
		}

		qty, found := it.Stats[stat.Quantity]
		if found && qty.Value == 12 {
			return false
		}
	}
	return true
}

func (sm ShopManager) SellJunk(d data.Data) {
	for _, i := range d.Items.Inventory {
		if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 1 {
			sm.logger.Debug(fmt.Sprintf("Item %s [%s] sold", i.Name, i.Quality))
			x := InventoryTopLeftX + i.Position.X*ItemBoxSize + (ItemBoxSize / 2)
			y := InventoryTopLeftY + i.Position.Y*ItemBoxSize + (ItemBoxSize / 2)
			hid.MovePointer(x, y)
			helper.Sleep(100)
			hid.KeyDown("control")
			helper.Sleep(50)
			hid.Click(hid.LeftButton)
			helper.Sleep(150)
			hid.KeyUp("control")
			helper.Sleep(500)
		}
	}
}

func (sm ShopManager) buyItem(i data.Item, quantity int) {
	x, y := sm.getScreenCordinatesForItem(i)

	hid.MovePointer(x, y)
	helper.Sleep(250)
	for k := 0; k < quantity; k++ {
		hid.Click(hid.RightButton)
		helper.Sleep(800)
		sm.logger.Debug(fmt.Sprintf("Purchased %s [X:%d Y:%d]", i.Name, i.Position.X, i.Position.Y))
	}
}

func (sm ShopManager) buyFullStack(i data.Item) {
	x, y := sm.getScreenCordinatesForItem(i)

	hid.MovePointer(x, y)
	helper.Sleep(250)
	hid.KeyDown("shift")
	helper.Sleep(100)
	hid.Click(hid.RightButton)
	helper.Sleep(300)
	hid.KeyUp("shift")
	helper.Sleep(500)
}

func (sm ShopManager) getScreenCordinatesForItem(i data.Item) (int, int) {
	x := topCornerVendorWindowX + i.Position.X*ItemBoxSize + (ItemBoxSize / 2)
	y := topCornerVendorWindowY + i.Position.Y*ItemBoxSize + (ItemBoxSize / 2)

	return x, y
}
