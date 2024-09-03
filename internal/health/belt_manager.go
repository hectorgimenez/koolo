package health

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
)

type BeltManager struct {
	data       *game.Data
	hid        *game.HID
	logger     *slog.Logger
	supervisor string
}

func NewBeltManager(data *game.Data, hid *game.HID, logger *slog.Logger, supervisor string) *BeltManager {
	return &BeltManager{
		data:       data,
		hid:        hid,
		logger:     logger,
		supervisor: supervisor,
	}
}

func (bm BeltManager) DrinkPotion(potionType data.PotionType, merc bool) bool {
	p, found := bm.data.Inventory.Belt.GetFirstPotion(potionType)
	if found {
		binding := bm.data.KeyBindings.UseBelt[p.X]
		if merc {
			bm.hid.PressKeyWithModifier(binding.Key1[0], game.ShiftKey)
			bm.logger.Debug(fmt.Sprintf("Using %s potion on Mercenary [Column: %d]. HP: %d", potionType, p.X+1, bm.data.MercHPPercent()))
			event.Send(event.UsedPotion(event.Text(bm.supervisor, ""), potionType, true))
			return true
		}
		bm.hid.PressKeyBinding(binding)
		bm.logger.Debug(fmt.Sprintf("Using %s potion [Column: %d]. HP: %d MP: %d", potionType, p.X+1, bm.data.PlayerUnit.HPPercent(), bm.data.PlayerUnit.MPPercent()))
		event.Send(event.UsedPotion(event.Text(bm.supervisor, ""), potionType, false))
		return true
	}

	return false
}

// ShouldBuyPotions will return true if more than 25% of belt is empty (ignoring rejuv)
func (bm BeltManager) ShouldBuyPotions() bool {
	targetHealingAmount := bm.data.CharacterCfg.Inventory.BeltColumns.Total(data.HealingPotion) * bm.data.Inventory.Belt.Rows()
	targetManaAmount := bm.data.CharacterCfg.Inventory.BeltColumns.Total(data.ManaPotion) * bm.data.Inventory.Belt.Rows()
	targetRejuvAmount := bm.data.CharacterCfg.Inventory.BeltColumns.Total(data.RejuvenationPotion) * bm.data.Inventory.Belt.Rows()

	currentHealing, currentMana, currentRejuv := bm.getCurrentPotions()

	bm.logger.Debug(fmt.Sprintf(
		"Belt Stats Health: %d/%d healing, %d/%d mana, %d/%d rejuv.",
		currentHealing,
		targetHealingAmount,
		currentMana,
		targetManaAmount,
		currentRejuv,
		targetRejuvAmount,
	))

	if currentHealing < int(float32(targetHealingAmount)*0.75) || currentMana < int(float32(targetManaAmount)*0.75) {
		bm.logger.Debug("Need more pots, let's buy them.")
		return true
	}

	return false
}

func (bm BeltManager) getCurrentPotions() (int, int, int) {
	currentHealing := 0
	currentMana := 0
	currentRejuv := 0
	for _, i := range bm.data.Inventory.Belt.Items {
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

func (bm BeltManager) GetMissingCount(potionType data.PotionType) int {
	currentHealing, currentMana, currentRejuv := bm.getCurrentPotions()

	switch potionType {
	case data.HealingPotion:
		targetAmount := bm.data.CharacterCfg.Inventory.BeltColumns.Total(data.HealingPotion) * bm.data.Inventory.Belt.Rows()
		missingPots := targetAmount - currentHealing
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case data.ManaPotion:
		targetAmount := bm.data.CharacterCfg.Inventory.BeltColumns.Total(data.ManaPotion) * bm.data.Inventory.Belt.Rows()
		missingPots := targetAmount - currentMana
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case data.RejuvenationPotion:
		targetAmount := bm.data.CharacterCfg.Inventory.BeltColumns.Total(data.RejuvenationPotion) * bm.data.Inventory.Belt.Rows()
		missingPots := targetAmount - currentRejuv
		if missingPots < 0 {
			return 0
		}
		return missingPots
	}

	return 0
}
