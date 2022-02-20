package health

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"go.uber.org/zap"
	"time"
)

const (
	healingInterval     = time.Second * 6
	healingMercInterval = time.Second * 6
	manaInterval        = time.Second * 4
	rejuvInterval       = time.Second * 2
)

// Manager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type Manager struct {
	logger        *zap.Logger
	beltManager   BeltManager
	lastRejuv     time.Time
	lastRejuvMerc time.Time
	lastHeal      time.Time
	lastMana      time.Time
	lastMercHeal  time.Time
}

func NewHealthManager(logger *zap.Logger, beltManager BeltManager) Manager {
	return Manager{
		logger:      logger,
		beltManager: beltManager,
	}
}

func (hm *Manager) HandleHealthAndMana(d game.Data) {
	hpConfig := config.Config.Health
	// Safe area, skipping
	if d.Area.IsTown() {
		return
	}

	status := d.Health

	usedRejuv := false
	if time.Since(hm.lastRejuv) > rejuvInterval && (status.HPPercent() <= hpConfig.RejuvPotionAtLife || status.MPPercent() < hpConfig.RejuvPotionAtMana) {
		usedRejuv = hm.beltManager.DrinkPotion(game.RejuvenationPotion, false)
		if usedRejuv {
			hm.lastRejuv = time.Now()
		}
	}

	if !usedRejuv {
		if status.HPPercent() <= hpConfig.ChickenAt {
			hm.chicken(status)
			return
		}

		if status.HPPercent() <= hpConfig.HealingPotionAt && time.Since(hm.lastHeal) > healingInterval {
			hm.beltManager.DrinkPotion(game.HealingPotion, false)
			hm.lastHeal = time.Now()
		}

		if status.MPPercent() <= hpConfig.ManaPotionAt && time.Since(hm.lastMana) > manaInterval {
			hm.beltManager.DrinkPotion(game.ManaPotion, false)
			hm.lastMana = time.Now()
		}
	}

	// Mercenary
	if status.Merc.Alive {
		usedMercRejuv := false
		if time.Since(hm.lastRejuvMerc) > rejuvInterval && status.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			usedMercRejuv = hm.beltManager.DrinkPotion(game.RejuvenationPotion, true)
			if usedMercRejuv {
				hm.lastRejuvMerc = time.Now()
			}
		}

		if !usedMercRejuv {
			if status.MercHPPercent() <= hpConfig.MercChickenAt {
				hm.chicken(status)
				return
			}

			if status.MercHPPercent() <= hpConfig.MercHealingPotionAt && time.Since(hm.lastMercHeal) > healingMercInterval {
				hm.beltManager.DrinkPotion(game.HealingPotion, true)
				hm.lastMercHeal = time.Now()
			}
		}
	}
}

func (hm Manager) chicken(status game.Health) {
	hm.logger.Warn(fmt.Sprintf("Chicken! Current Health: %d (%d percent)", status.Life, status.HPPercent()))
	helper.ExitGame()
}
