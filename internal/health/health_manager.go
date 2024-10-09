package health

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

var ErrDied = errors.New("you died :(")
var ErrChicken = errors.New("chicken")
var ErrMercChicken = errors.New("mercenary chicken")

const (
	healingInterval     = time.Second * 4
	healingMercInterval = time.Second * 6
	manaInterval        = time.Second * 4
	rejuvInterval       = time.Second * 2
)

// Manager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type Manager struct {
	lastRejuv     time.Time
	lastRejuvMerc time.Time
	lastHeal      time.Time
	lastMana      time.Time
	lastMercHeal  time.Time
	beltManager   *BeltManager
	data          *game.Data
}

func NewHealthManager(bm *BeltManager, data *game.Data) *Manager {
	return &Manager{
		beltManager: bm,
		data:        data,
	}
}

func (hm *Manager) HandleHealthAndMana() error {
	cfg := hm.data.CharacterCfg.Health
	now := time.Now()
	if hm.data.PlayerUnit.Area.IsTown() {
		return nil
	}

	if hm.data.PlayerUnit.HPPercent() <= 0 {
		return ErrDied
	}

	if hm.data.PlayerUnit.HPPercent() <= cfg.ChickenAt {
		return fmt.Errorf("%w: Current Health: %d percent", ErrChicken, hm.data.PlayerUnit.HPPercent())
	}

	// Prioritize Rejuvenation potions if available
	if hm.beltManager.HasPotion(data.RejuvenationPotion) {
		if (hm.data.PlayerUnit.HPPercent() <= cfg.RejuvPotionAtLife || hm.data.PlayerUnit.MPPercent() <= cfg.RejuvPotionAtMana) &&
			now.Sub(hm.lastRejuv) > rejuvInterval {
			if hm.beltManager.DrinkPotion(data.RejuvenationPotion, false) {
				hm.lastRejuv = now
				return nil
			}
		}
	}

	// If no rejuv available or not needed, check for healing potions
	if hm.data.PlayerUnit.HPPercent() <= cfg.HealingPotionAt && now.Sub(hm.lastHeal) > healingInterval {
		if hm.beltManager.DrinkPotion(data.HealingPotion, false) {
			hm.lastHeal = now
		} else {
			return fmt.Errorf("%w: No healing potions. Health: %d percent", ErrChicken, hm.data.PlayerUnit.HPPercent())
		}
	}

	// Check for mana potions
	if hm.data.PlayerUnit.MPPercent() <= cfg.ManaPotionAt && now.Sub(hm.lastMana) > manaInterval {
		if hm.beltManager.DrinkPotion(data.ManaPotion, false) {
			hm.lastMana = now
		}
	}

	// Mercenary health management
	if hm.data.MercHPPercent() > 0 {
		if hm.data.MercHPPercent() <= cfg.MercChickenAt {
			return fmt.Errorf("%w: Merc Health: %d percent", ErrMercChicken, hm.data.MercHPPercent())
		}

		if hm.data.MercHPPercent() <= cfg.MercRejuvPotionAt && now.Sub(hm.lastRejuvMerc) > rejuvInterval {
			if hm.beltManager.DrinkPotion(data.RejuvenationPotion, true) {
				hm.lastRejuvMerc = now
				return nil
			}
		}

		if hm.data.MercHPPercent() <= cfg.MercHealingPotionAt && now.Sub(hm.lastMercHeal) > healingMercInterval {
			if hm.beltManager.DrinkPotion(data.HealingPotion, true) {
				hm.lastMercHeal = now
			}
		}
	}

	return nil
}
