package health

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/event/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"go.uber.org/zap"
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
	logger        *zap.Logger
	beltManager   BeltManager
	gameManager   *helper.GameManager
	lastRejuv     time.Time
	lastRejuvMerc time.Time
	lastHeal      time.Time
	lastMana      time.Time
	lastMercHeal  time.Time
}

func NewHealthManager(logger *zap.Logger, beltManager BeltManager, gm *helper.GameManager) Manager {
	return Manager{
		logger:      logger,
		beltManager: beltManager,
		gameManager: gm,
	}
}

func (hm *Manager) HandleHealthAndMana(d data.Data) error {
	hpConfig := config.Config.Health
	// Safe area, skipping
	if d.PlayerUnit.Area.IsTown() {
		return nil
	}

	if d.PlayerUnit.HPPercent() <= 0 {
		stat.FinishCurrentRun(event.Death)

		return ErrDied
	}

	usedRejuv := false
	if time.Since(hm.lastRejuv) > rejuvInterval && (d.PlayerUnit.HPPercent() <= hpConfig.RejuvPotionAtLife || d.PlayerUnit.MPPercent() < hpConfig.RejuvPotionAtMana) {
		usedRejuv = hm.beltManager.DrinkPotion(d, data.RejuvenationPotion, false)
		if usedRejuv {
			hm.lastRejuv = time.Now()
		}
	}

	if !usedRejuv {
		if d.PlayerUnit.HPPercent() <= hpConfig.ChickenAt {
			stat.FinishCurrentRun(event.Chicken)
			return fmt.Errorf("%w: Current Health: %d percent", ErrChicken, d.PlayerUnit.HPPercent())
		}

		if d.PlayerUnit.HPPercent() <= hpConfig.HealingPotionAt && time.Since(hm.lastHeal) > healingInterval {
			hm.beltManager.DrinkPotion(d, data.HealingPotion, false)
			hm.lastHeal = time.Now()
		}

		if d.PlayerUnit.MPPercent() <= hpConfig.ManaPotionAt && time.Since(hm.lastMana) > manaInterval {
			hm.beltManager.DrinkPotion(d, data.ManaPotion, false)
			hm.lastMana = time.Now()
		}
	}

	// Mercenary
	if d.MercHPPercent() > 0 {
		usedMercRejuv := false
		if time.Since(hm.lastRejuvMerc) > rejuvInterval && d.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			usedMercRejuv = hm.beltManager.DrinkPotion(d, data.RejuvenationPotion, true)
			if usedMercRejuv {
				hm.lastRejuvMerc = time.Now()
			}
		}

		if !usedMercRejuv {
			if d.MercHPPercent() <= hpConfig.MercChickenAt {
				stat.FinishCurrentRun(event.MercChicken)
				return fmt.Errorf("%w: Current Merc Health: %d percent", ErrMercChicken, d.MercHPPercent())
			}

			if d.MercHPPercent() <= hpConfig.MercHealingPotionAt && time.Since(hm.lastMercHeal) > healingMercInterval {
				hm.beltManager.DrinkPotion(d, data.HealingPotion, true)
				hm.lastMercHeal = time.Now()
			}
		}
	}

	return nil
}
