package health

import (
	"fmt"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event/stat"
	"github.com/hectorgimenez/koolo/internal/hid"
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

func (pm BeltManager) DrinkPotion(d data.Data, potionType data.PotionType, merc bool) bool {
	p, found := d.Items.Belt.GetFirstPotion(potionType)
	if found {
		binding := pm.getBindingBasedOnColumn(p)
		if merc {
			hid.KeyDown("shift")
			hid.PressKey(binding)
			hid.KeyUp("shift")
			pm.logger.Debug(fmt.Sprintf("Using %s potion on Mercenary [Column: %d]. HP: %d", potionType, p.X+1, d.MercHPPercent()))
			stat.UsedPotion(potionType, true)
			return true
		}
		hid.PressKey(binding)
		pm.logger.Debug(fmt.Sprintf("Using %s potion [Column: %d]. HP: %d MP: %d", potionType, p.X+1, d.PlayerUnit.HPPercent(), d.PlayerUnit.MPPercent()))
		stat.UsedPotion(potionType, false)
		return true
	}

	return false
}

// ShouldBuyPotions will return true if more than 25% of belt is empty (ignoring rejuv)
func (pm BeltManager) ShouldBuyPotions(d data.Data) bool {
	targetHealingAmount := config.Config.Inventory.BeltColumns.Healing * d.Items.Belt.Rows()
	targetManaAmount := config.Config.Inventory.BeltColumns.Mana * d.Items.Belt.Rows()
	targetRejuvAmount := config.Config.Inventory.BeltColumns.Rejuvenation * d.Items.Belt.Rows()

	currentHealing, currentMana, currentRejuv := pm.getCurrentPotions(d)

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

func (pm BeltManager) getCurrentPotions(d data.Data) (int, int, int) {
	currentHealing := 0
	currentMana := 0
	currentRejuv := 0
	for _, i := range d.Items.Belt.Items {
		if strings.Contains(string(i.Name), string(data.HealingPotion)) {
			currentHealing++
			continue
		}
		if strings.Contains(string(i.Name), string(data.ManaPotion)) {
			currentMana++
			continue
		}
		if strings.Contains(string(i.Name), string(data.RejuvenationPotion)) {
			currentRejuv++
		}
	}

	return currentHealing, currentMana, currentRejuv
}

func (pm BeltManager) GetMissingCount(d data.Data, potionType data.PotionType) int {
	currentHealing, currentMana, currentRejuv := pm.getCurrentPotions(d)

	switch potionType {
	case data.HealingPotion:
		targetAmount := config.Config.Inventory.BeltColumns.Healing * d.Items.Belt.Rows()
		missingPots := targetAmount - currentHealing
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case data.ManaPotion:
		targetAmount := config.Config.Inventory.BeltColumns.Mana * d.Items.Belt.Rows()
		missingPots := targetAmount - currentMana
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case data.RejuvenationPotion:
		targetAmount := config.Config.Inventory.BeltColumns.Rejuvenation * d.Items.Belt.Rows()
		missingPots := targetAmount - currentRejuv
		if missingPots < 0 {
			return 0
		}
		return missingPots
	}

	return 0
}

func (pm BeltManager) getBindingBasedOnColumn(position data.Position) string {
	switch position.X {
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
