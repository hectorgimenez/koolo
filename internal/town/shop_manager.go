package town

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	topCornerWindowWidthProportion  = 37.64
	topCornerWindowHeightProportion = 7.85
	ItemBoxSize                     = 40

	InventoryTopLeftX = 1.494
	InventoryTopLeftY = 2.071
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

func (sm ShopManager) BuyConsumables(d game.Data) {
	missingHealingPots := sm.bm.GetMissingCount(d, game.HealingPotion)
	missingManaPots := sm.bm.GetMissingCount(d, game.ManaPotion)

	sm.logger.Debug(fmt.Sprintf("Buying: %d Healing potions and %d Mana potions", missingHealingPots, missingManaPots))

	vendor := GetTownByArea(d.Area).RefillNPC()
	for _, i := range d.Items.Shop {
		if strings.Contains(i.Name, "HealingPotion") && i.Vendor == vendor && missingHealingPots > 1 {
			sm.buyItem(i, missingHealingPots)
			missingHealingPots = 0
			break
		}
	}
	for _, i := range d.Items.Shop {
		if strings.Contains(i.Name, "ManaPotion") && i.Vendor == vendor && missingManaPots > 1 {
			sm.buyItem(i, missingManaPots)
			missingManaPots = 0
			break
		}
	}

	if d.Items.Inventory.ShouldBuyTPs() {
		sm.logger.Debug("Filling TP Tome...")
		for _, i := range d.Items.Shop {
			if i.Name == game.ItemScrollOfTownPortal && i.Vendor == vendor {
				sm.buyFullStack(i)
				break
			}
		}
	}

	if d.Items.Inventory.ShouldBuyIDs() {
		sm.logger.Debug("Filling IDs Tome...")
		for _, i := range d.Items.Shop {
			if i.Name == game.ItemScrollOfIdentify && i.Vendor == vendor {
				sm.buyFullStack(i)
				break
			}
		}
	}
}

func (sm ShopManager) SellJunk(d game.Data) {
	for _, i := range d.Items.Inventory {
		if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 1 {
			x := int(float32(hid.GameAreaSizeX)/InventoryTopLeftX) + i.Position.X*ItemBoxSize + (ItemBoxSize / 2)
			y := int(float32(hid.GameAreaSizeY)/InventoryTopLeftY) + i.Position.Y*ItemBoxSize + (ItemBoxSize / 2)
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

func (sm ShopManager) buyItem(i game.Item, quantity int) {
	x, y := sm.getScreenCordinatesForItem(i)

	hid.MovePointer(x, y)
	time.Sleep(time.Millisecond * 250)
	for k := 0; k < quantity; k++ {
		hid.Click(hid.RightButton)
		time.Sleep(time.Millisecond * 500)
		sm.logger.Debug(fmt.Sprintf("Purchased %s [X:%d Y:%d]", i.Name, i.Position.X, i.Position.Y))
	}
}

func (sm ShopManager) buyFullStack(i game.Item) {
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

func (sm ShopManager) getScreenCordinatesForItem(i game.Item) (int, int) {
	topLeftShoppingWindowX := int(float32(hid.GameAreaSizeX) / topCornerWindowWidthProportion)
	topLeftShoppingWindowY := int(float32(hid.GameAreaSizeY) / topCornerWindowHeightProportion)

	x := topLeftShoppingWindowX + i.Position.X*ItemBoxSize + (ItemBoxSize / 2)
	y := topLeftShoppingWindowY + i.Position.Y*ItemBoxSize + (ItemBoxSize / 2)

	return x, y
}
