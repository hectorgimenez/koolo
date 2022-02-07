package town

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game/data"
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
	logger     *zap.Logger
	dr         data.DataRepository
	bm         health.BeltManager
	actionChan chan<- action.Action
}

func NewShopManager(logger *zap.Logger, dr data.DataRepository, bm health.BeltManager, actionChan chan<- action.Action) ShopManager {
	return ShopManager{
		logger:     logger,
		dr:         dr,
		bm:         bm,
		actionChan: actionChan,
	}
}
func (sm ShopManager) buyPotsAndTPs() {
	d := sm.dr.GameData()
	missingHealingPots := sm.bm.GetMissingCount(data.HealingPotion)
	missingManaPots := sm.bm.GetMissingCount(data.ManaPotion)

	sm.logger.Debug(fmt.Sprintf("Buying: %d Healing potions and %d Mana potions", missingHealingPots, missingManaPots))

	for _, i := range d.Items.Shop {
		if i.Name == data.ItemSuperHealingPotion && missingHealingPots > 1 {
			sm.buyItem(i, missingHealingPots)
			missingHealingPots = 0
		}
	}
	for _, i := range d.Items.Shop {
		if i.Name == data.ItemSuperManaPotion && missingManaPots > 1 {
			sm.buyItem(i, missingManaPots)
			missingManaPots = 0
		}
	}
}

func (sm ShopManager) buyItem(i data.Item, quantity int) {
	topLeftShoppingWindowX := int(float32(hid.GameAreaSizeX) / topCornerWindowWidthProportion)
	topLeftShoppingWindowY := int(float32(hid.GameAreaSizeY) / topCornerWindowHeightProportion)

	x := topLeftShoppingWindowX + i.Position.X*itemBoxSize + (itemBoxSize / 2)
	y := topLeftShoppingWindowY + i.Position.Y*itemBoxSize + (itemBoxSize / 2)

	mouseOps := []action.HIDOperation{action.NewMouseDisplacement(x, y, time.Millisecond*250)}
	for k := 0; k < quantity; k++ {
		mouseOps = append(mouseOps, action.NewMouseClick(hid.RightButton, time.Second*1))
	}
	mouseOps = append(mouseOps)
	sm.actionChan <- action.NewAction(action.PriorityNormal, mouseOps...)
}
