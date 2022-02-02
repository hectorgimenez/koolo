package health

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"go.uber.org/zap"
	"time"
)

type BeltManager struct {
	logger              *zap.Logger
	cfg                 config.Config
	inventoryRepository data.DataRepository
	actionChan          chan<- action.Action
}

func NewBeltManager(logger *zap.Logger, cfg config.Config, repository data.DataRepository, actionChan chan<- action.Action) BeltManager {
	return BeltManager{
		logger:              logger,
		cfg:                 cfg,
		inventoryRepository: repository,
		actionChan:          actionChan,
	}
}

func (pm BeltManager) DrinkPotion(potionType data.PotionType, merc bool) {
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

	pm.logger.Warn(fmt.Sprintf("Tried to use %s but failed, no more pots left!", potionType))
}

// ShouldBuyPotions will return true if more than 25% of belt is empty (ignoring rejuv)
func (pm BeltManager) ShouldBuyPotions() bool {
	targetHealingAmount := pm.cfg.Inventory.BeltColumns.Healing * pm.cfg.Inventory.BeltRows
	targetManaAmount := pm.cfg.Inventory.BeltColumns.Mana * pm.cfg.Inventory.BeltRows

	currentHealing, currentMana, currentRejuv := pm.getCurrentPotions()

	pm.logger.Debug(fmt.Sprintf("Belt Status: %d healing, %d mana, %d rejuv.", currentHealing, currentMana, currentRejuv))

	if currentHealing < int(float32(targetHealingAmount)*0.75) || currentMana < int(float32(targetManaAmount)*0.75) {
		pm.logger.Debug("Need more pots, let's buy them.")
		return true
	}

	return false
}

func (pm BeltManager) getCurrentPotions() (int, int, int) {
	currentHealing := 0
	currentMana := 0
	currentRejuv := 0
	for _, p := range pm.belt().Potions {
		if p.Type == data.HealingPotion {
			currentHealing++
		}
		if p.Type == data.ManaPotion {
			currentMana++
		}
		if p.Type == data.RejuvenationPotion {
			currentRejuv++
		}
	}

	return currentHealing, currentMana, currentRejuv
}

func (pm BeltManager) GetMissingCount(potionType data.PotionType) int {
	currentHealing, currentMana, _ := pm.getCurrentPotions()

	if potionType == data.HealingPotion {
		targetAmount := pm.cfg.Inventory.BeltColumns.Healing * pm.cfg.Inventory.BeltRows
		return targetAmount - currentHealing
	}

	targetAmount := pm.cfg.Inventory.BeltColumns.Mana * pm.cfg.Inventory.BeltRows
	return targetAmount - currentMana
}

func (pm BeltManager) belt() data.Belt {
	return pm.inventoryRepository.GameData().Items.Belt
}

func (pm BeltManager) getBindingBasedOnColumn(potion data.Potion) string {
	switch potion.Position.X {
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
