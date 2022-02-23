package health

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/stats"
	"go.uber.org/zap"
)

type BeltManager struct {
	logger *zap.Logger
}

func NewBeltManager(logger *zap.Logger) BeltManager {
	return BeltManager{
		logger: logger,
	}
}

func (pm BeltManager) DrinkPotion(data game.Data, potionType game.PotionType, merc bool) bool {
	p, found := data.Items.Belt.GetFirstPotion(potionType)
	if found {
		binding := pm.getBindingBasedOnColumn(p)
		if merc {
			hid.PressKeyCombination("shift", binding)
			pm.logger.Debug(fmt.Sprintf("Using %s potion on Mercenary [Row: %d]. HP: %d", potionType, p.Position.X+1, data.Health.MercHPPercent()))
			stats.UsedPotion(potionType, true)
			return true
		}
		hid.PressKey(binding)
		pm.logger.Debug(fmt.Sprintf("Using %s potio [Row: %d]. HP: %d MP: %d", potionType, p.Position.X+1, data.Health.HPPercent(), data.Health.MPPercent()))
		stats.UsedPotion(potionType, false)
		return true
	}

	return false
}

// ShouldBuyPotions will return true if more than 25% of belt is empty (ignoring rejuv)
func (pm BeltManager) ShouldBuyPotions(data game.Data) bool {
	targetHealingAmount := config.Config.Inventory.BeltColumns.Healing * config.Config.Inventory.BeltRows
	targetManaAmount := config.Config.Inventory.BeltColumns.Mana * config.Config.Inventory.BeltRows
	targetRejuvAmount := config.Config.Inventory.BeltColumns.Rejuvenation * config.Config.Inventory.BeltRows

	currentHealing, currentMana, currentRejuv := pm.getCurrentPotions(data)

	pm.logger.Debug(fmt.Sprintf(
		"Belt Status Health: %d/%d healing, %d/%d mana, %d/%d rejuv.",
		currentHealing,
		targetHealingAmount,
		currentMana,
		targetManaAmount,
		currentRejuv,
		targetRejuvAmount,
	))

	if currentHealing < int(float32(targetHealingAmount)*0.75) || currentMana < int(float32(targetManaAmount)*0.75) {
		pm.logger.Debug("Need more pots, let's buy them.")
		return true
	}

	return false
}

func (pm BeltManager) getCurrentPotions(data game.Data) (int, int, int) {
	currentHealing := 0
	currentMana := 0
	currentRejuv := 0
	for _, p := range data.Items.Belt.Potions {
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

func (pm BeltManager) GetMissingCount(data game.Data, potionType game.PotionType) int {
	currentHealing, currentMana, currentRejuv := pm.getCurrentPotions(data)

	switch potionType {
	case game.HealingPotion:
		targetAmount := config.Config.Inventory.BeltColumns.Healing * config.Config.Inventory.BeltRows
		missingPots := targetAmount - currentHealing
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case game.ManaPotion:
		targetAmount := config.Config.Inventory.BeltColumns.Mana * config.Config.Inventory.BeltRows
		missingPots := targetAmount - currentMana
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case game.RejuvenationPotion:
		targetAmount := config.Config.Inventory.BeltColumns.Rejuvenation * config.Config.Inventory.BeltRows
		missingPots := targetAmount - currentRejuv
		if missingPots < 0 {
			return 0
		}
		return missingPots
	}

	return 0
}

func (pm BeltManager) getBindingBasedOnColumn(potion game.Potion) string {
	switch potion.Position.X {
	case 0:
		return config.Config.Bindings.Potion1
	case 1:
		return config.Config.Bindings.Potion2
	case 2:
		return config.Config.Bindings.Potion3
	case 3:
		return config.Config.Bindings.Potion4
	}

	return ""
}
