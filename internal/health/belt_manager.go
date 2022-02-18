package health

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
)

type BeltManager struct {
	logger *zap.Logger
	cfg    config.Config
}

func NewBeltManager(logger *zap.Logger, cfg config.Config) BeltManager {
	return BeltManager{
		logger: logger,
		cfg:    cfg,
	}
}

func (pm BeltManager) DrinkPotion(potionType game.PotionType, merc bool) {
	belt := pm.belt()
	p, found := belt.GetFirstPotion(potionType)
	if found {
		binding := pm.getBindingBasedOnColumn(p)
		if merc {
			hid.PressKeyCombination("shift", binding)
			pm.logger.Debug(fmt.Sprintf("Using %s potion on Mercenary", potionType))
			return
		}
		hid.PressKey(binding)
		pm.logger.Debug(fmt.Sprintf("Drinking %s potion", potionType))
		return
	}

	pm.logger.Warn(fmt.Sprintf("Tried to use %s but failed, no more pots left!", potionType))
}

// ShouldBuyPotions will return true if more than 25% of belt is empty (ignoring rejuv)
func (pm BeltManager) ShouldBuyPotions() bool {
	targetHealingAmount := pm.cfg.Inventory.BeltColumns.Healing * pm.cfg.Inventory.BeltRows
	targetManaAmount := pm.cfg.Inventory.BeltColumns.Mana * pm.cfg.Inventory.BeltRows

	currentHealing, currentMana, currentRejuv := pm.getCurrentPotions()

	pm.logger.Debug(fmt.Sprintf("Belt Health: %d healing, %d mana, %d rejuv.", currentHealing, currentMana, currentRejuv))

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
		if p.Type == game.HealingPotion {
			currentHealing++
		}
		if p.Type == game.ManaPotion {
			currentMana++
		}
		if p.Type == game.RejuvenationPotion {
			currentRejuv++
		}
	}

	return currentHealing, currentMana, currentRejuv
}

func (pm BeltManager) GetMissingCount(potionType game.PotionType) int {
	currentHealing, currentMana, _ := pm.getCurrentPotions()

	if potionType == game.HealingPotion {
		targetAmount := pm.cfg.Inventory.BeltColumns.Healing * pm.cfg.Inventory.BeltRows
		missingPots := targetAmount - currentHealing
		if missingPots < 0 {
			return 0
		}
		return missingPots
	}

	targetAmount := pm.cfg.Inventory.BeltColumns.Mana * pm.cfg.Inventory.BeltRows
	missingPots := targetAmount - currentMana
	if missingPots < 0 {
		return 0
	}
	return missingPots
}

func (pm BeltManager) belt() game.Belt {
	return game.Status().Items.Belt
}

func (pm BeltManager) getBindingBasedOnColumn(potion game.Potion) string {
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
