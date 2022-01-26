package health

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/inventory"
	"time"
)

type BeltManager struct {
	cfg                 config.Config
	inventoryRepository inventory.InventoryRepository
	actionChan          chan<- action.Action
}

func NewBeltManager(cfg config.Config, repository inventory.InventoryRepository, actionChan chan<- action.Action) BeltManager {
	return BeltManager{
		cfg:                 cfg,
		inventoryRepository: repository,
		actionChan:          actionChan,
	}
}

func (pm BeltManager) DrinkPotion(potionType inventory.PotionType, merc bool) {
	belt := pm.belt()
	p, found := belt.GetFirstPotion(potionType)
	if found {
		binding := pm.getBindingBasedOnColumn(p)
		if merc {
			pm.actionChan <- action.NewAction(action.PriorityHigh, action.NewKeyPress("shift", time.Millisecond*50, binding))
			return
		}
		pm.actionChan <- action.NewAction(action.PriorityHigh, action.NewKeyPress(binding, time.Millisecond*50))
		return
	}
	// TODO: Log warning, no more potions available
}

func (pm BeltManager) belt() inventory.Belt {
	return pm.inventoryRepository.Inventory().Belt
}

func (pm BeltManager) getBindingBasedOnColumn(potion inventory.Potion) string {
	switch potion.Column {
	case 0:
		return pm.cfg.Bindings.Potion1
	case 1:
		return pm.cfg.Bindings.Potion2
	case 2:
		return pm.cfg.Bindings.Potion3
	case 3:
		return pm.cfg.Bindings.Potion4
	}

	return ""
}
