package health

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"log/slog"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
)

type BeltManager struct {
	logger         *slog.Logger
	hid            *game.HID
	cfg            *config.CharacterCfg
	supervisorName string
}

func NewBeltManager(logger *slog.Logger, hid *game.HID, cfg *config.CharacterCfg, supervisorName string) BeltManager {
	return BeltManager{
		logger:         logger,
		hid:            hid,
		cfg:            cfg,
		supervisorName: supervisorName,
	}
}

func (bm BeltManager) DrinkPotion(d game.Data, potionType data.PotionType, merc bool) bool {
	p, found := d.Items.Belt.GetFirstPotion(potionType)
	if found {
		binding := d.KeyBindings.UseBelt[p.X]
		if merc {
			bm.hid.PressKeyWithModifier(binding.Key1[0], game.ShiftKey)
			bm.logger.Debug(fmt.Sprintf("Using %s potion on Mercenary [Column: %d]. HP: %d", potionType, p.X+1, d.MercHPPercent()))
			event.Send(event.UsedPotion(event.Text(bm.supervisorName, ""), potionType, true))
			return true
		}
		bm.hid.PressKeyBinding(binding)
		bm.logger.Debug(fmt.Sprintf("Using %s potion [Column: %d]. HP: %d MP: %d", potionType, p.X+1, d.PlayerUnit.HPPercent(), d.PlayerUnit.MPPercent()))
		event.Send(event.UsedPotion(event.Text(bm.supervisorName, ""), potionType, false))
		return true
	}

	return false
}

// ShouldBuyPotions will return true if more than 25% of belt is empty (ignoring rejuv)
func (bm BeltManager) ShouldBuyPotions(d game.Data) bool {
	targetHealingAmount := bm.cfg.Inventory.BeltColumns.Total(data.HealingPotion) * d.Items.Belt.Rows()
	targetManaAmount := bm.cfg.Inventory.BeltColumns.Total(data.ManaPotion) * d.Items.Belt.Rows()
	targetRejuvAmount := bm.cfg.Inventory.BeltColumns.Total(data.RejuvenationPotion) * d.Items.Belt.Rows()

	currentHealing, currentMana, currentRejuv := bm.getCurrentPotions(d)

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

func (bm BeltManager) getCurrentPotions(d game.Data) (int, int, int) {
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

func (bm BeltManager) GetMissingCount(d game.Data, potionType data.PotionType) int {
	currentHealing, currentMana, currentRejuv := bm.getCurrentPotions(d)

	switch potionType {
	case data.HealingPotion:
		targetAmount := bm.cfg.Inventory.BeltColumns.Total(data.HealingPotion) * d.Items.Belt.Rows()
		missingPots := targetAmount - currentHealing
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case data.ManaPotion:
		targetAmount := bm.cfg.Inventory.BeltColumns.Total(data.ManaPotion) * d.Items.Belt.Rows()
		missingPots := targetAmount - currentMana
		if missingPots < 0 {
			return 0
		}
		return missingPots
	case data.RejuvenationPotion:
		targetAmount := bm.cfg.Inventory.BeltColumns.Total(data.RejuvenationPotion) * d.Items.Belt.Rows()
		missingPots := targetAmount - currentRejuv
		if missingPots < 0 {
			return 0
		}
		return missingPots
	}

	return 0
}
