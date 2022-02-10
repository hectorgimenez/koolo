package town

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
	"time"
)

const (
	topCornerWindowWidthProportion  = 37.64
	topCornerWindowHeightProportion = 7.85
	itemBoxSize                     = 40
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
func (sm ShopManager) buyPotsAndTPs(buyTPs bool) {
	d := game.Status()
	missingHealingPots := sm.bm.GetMissingCount(game.HealingPotion)
	missingManaPots := sm.bm.GetMissingCount(game.ManaPotion)

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

	if buyTPs {
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

	mouseOps := []action.HIDOperation{action.NewMouseDisplacement(x, y, time.Millisecond*250)}
	for k := 0; k < quantity; k++ {
		mouseOps = append(mouseOps, action.NewMouseClick(hid.RightButton, time.Second*1))
	}
	mouseOps = append(mouseOps)

	action.Run(mouseOps...)
}

func (sm ShopManager) buyFullStack(i game.Item) {
	x, y := sm.getScreenCordinatesForItem(i)

	action.Run(
		action.NewMouseDisplacement(x, y, time.Millisecond*250),
		action.NewKeyDown("shift", time.Millisecond*100),
		action.NewMouseClick(hid.RightButton, time.Millisecond*300),
		action.NewKeyUp("shift", time.Second),
	)
}

func (sm ShopManager) getScreenCordinatesForItem(i game.Item) (int, int) {
	topLeftShoppingWindowX := int(float32(hid.GameAreaSizeX) / topCornerWindowWidthProportion)
	topLeftShoppingWindowY := int(float32(hid.GameAreaSizeY) / topCornerWindowHeightProportion)

	x := topLeftShoppingWindowX + i.Position.X*itemBoxSize + (itemBoxSize / 2)
	y := topLeftShoppingWindowY + i.Position.Y*itemBoxSize + (itemBoxSize / 2)

	return x, y
}
