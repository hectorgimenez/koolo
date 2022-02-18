package health

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"go.uber.org/zap"
	"time"
)

const (
	healingInterval     = time.Second * 6
	healingMercInterval = time.Second * 6
	manaInterval        = time.Second * 4
)

// Manager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type Manager struct {
	logger       *zap.Logger
	beltManager  BeltManager
	cfg          config.Config
	lastHeal     time.Time
	lastMana     time.Time
	lastMercHeal time.Time
}

func NewHealthManager(logger *zap.Logger, beltManager BeltManager, cfg config.Config) Manager {
	return Manager{
		logger:      logger,
		beltManager: beltManager,
		cfg:         cfg,
	}
}

func (hm *Manager) HandleHealthAndMana(d game.Data) {
	hpConfig := hm.cfg.Health
	// Safe area, skipping
	if d.Area.IsTown() {
		return
	}

	status := d.Health

	usedRejuv := false
	if status.HPPercent() <= hpConfig.RejuvPotionAtLife || status.MPPercent() < hpConfig.RejuvPotionAtMana {
		hm.beltManager.DrinkPotion(game.RejuvenationPotion, false)
		usedRejuv = true
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
		if status.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			hm.beltManager.DrinkPotion(game.RejuvenationPotion, true)
			usedMercRejuv = true
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
	//helper.ExitGame()
}
