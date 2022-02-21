package town

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
	"time"
)

const (
	topCornerWindowWidthProportion  = 37.64
	topCornerWindowHeightProportion = 7.85
	ItemBoxSize                     = 40
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

func (sm ShopManager) BuyPotsAndTPs(d game.Data) {
	missingHealingPots := sm.bm.GetMissingCount(d, game.HealingPotion)
	missingManaPots := sm.bm.GetMissingCount(d, game.ManaPotion)

	sm.logger.Debug(fmt.Sprintf("Buying: %d Healing potions and %d Mana potions", missingHealingPots, missingManaPots))

	for _, i := range d.Items.Shop {
		if i.Name == game.ItemSuperHealingPotion && missingHealingPots > 1 {
			sm.buyItem(i, missingHealingPots)
			missingHealingPots = 0
			break
		}
	}
	for _, i := range d.Items.Shop {
		if i.Name == game.ItemSuperManaPotion && missingManaPots > 1 {
			sm.buyItem(i, missingManaPots)
			missingManaPots = 0
			break
		}
	}

	if d.Items.Inventory.ShouldBuyTPs() {
		sm.logger.Debug("Filling TP Tome...")
		for _, i := range d.Items.Shop {
			if i.Name == game.ItemScrollTownPortal {
				sm.buyFullStack(i)
				break
			}
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
	}
}

func (sm ShopManager) buyFullStack(i game.Item) {
	x, y := sm.getScreenCordinatesForItem(i)

	hid.MovePointer(x, y)
	time.Sleep(time.Millisecond * 250)
	hid.KeyDown("shift")
	time.Sleep(time.Millisecond * 100)
	hid.Click(hid.RightButton)
	time.Sleep(time.Millisecond * 300)
	hid.KeyUp("shift")
	time.Sleep(time.Second)
}

func (sm ShopManager) getScreenCordinatesForItem(i game.Item) (int, int) {
	topLeftShoppingWindowX := int(float32(hid.GameAreaSizeX) / topCornerWindowWidthProportion)
	topLeftShoppingWindowY := int(float32(hid.GameAreaSizeY) / topCornerWindowHeightProportion)

	x := topLeftShoppingWindowX + i.Position.X*ItemBoxSize + (ItemBoxSize / 2)
	y := topLeftShoppingWindowY + i.Position.Y*ItemBoxSize + (ItemBoxSize / 2)

	return x, y
}
