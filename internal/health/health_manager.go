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
	hpConfig := hm.data.CharacterCfg.Health
	// Safe area, skipping
	if hm.data.PlayerUnit.Area.IsTown() {
		return nil
	}

	if hm.data.PlayerUnit.HPPercent() <= 0 {
		return ErrDied
	}

	usedRejuv := false
	if time.Since(hm.lastRejuv) > rejuvInterval && (hm.data.PlayerUnit.HPPercent() <= hpConfig.RejuvPotionAtLife || hm.data.PlayerUnit.MPPercent() < hpConfig.RejuvPotionAtMana) {
		usedRejuv = hm.beltManager.DrinkPotion(data.RejuvenationPotion, false)
		if usedRejuv {
			hm.lastRejuv = time.Now()
		}
	}

	if !usedRejuv {
		if hm.data.PlayerUnit.HPPercent() <= hpConfig.ChickenAt {
			return fmt.Errorf("%w: Current Health: %d percent", ErrChicken, hm.data.PlayerUnit.HPPercent())
		}

		if hm.data.PlayerUnit.HPPercent() <= hpConfig.HealingPotionAt && time.Since(hm.lastHeal) > healingInterval {
			hm.beltManager.DrinkPotion(data.HealingPotion, false)
			hm.lastHeal = time.Now()
		}

		if hm.data.PlayerUnit.MPPercent() <= hpConfig.ManaPotionAt && time.Since(hm.lastMana) > manaInterval {
			hm.beltManager.DrinkPotion(data.ManaPotion, false)
			hm.lastMana = time.Now()
		}
	}

	// Mercenary
	if hm.data.MercHPPercent() > 0 {
		usedMercRejuv := false
		if time.Since(hm.lastRejuvMerc) > rejuvInterval && hm.data.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			usedMercRejuv = hm.beltManager.DrinkPotion(data.RejuvenationPotion, true)
			if usedMercRejuv {
				hm.lastRejuvMerc = time.Now()
			}
		}

		if !usedMercRejuv {
			if hm.data.MercHPPercent() <= hpConfig.MercChickenAt {
				return fmt.Errorf("%w: Current Merc Health: %d percent", ErrMercChicken, hm.data.MercHPPercent())
			}

			if hm.data.MercHPPercent() <= hpConfig.MercHealingPotionAt && time.Since(hm.lastMercHeal) > healingMercInterval {
				hm.beltManager.DrinkPotion(data.HealingPotion, true)
				hm.lastMercHeal = time.Now()
			}
		}
	}

	return nil
}
