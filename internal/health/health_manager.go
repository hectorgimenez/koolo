package health

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/helper"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"time"
)

const (
	monitorEvery        = time.Millisecond * 500
	healingInterval     = time.Second * 6
	healingMercInterval = time.Second * 6
	manaInterval        = time.Second * 4
)

// Manager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type Manager struct {
	logger       *zap.Logger
	hr           Repository
	eventChan    chan<- event.Event
	beltManager  BeltManager
	cfg          config.Config
	lastHeal     time.Time
	lastMana     time.Time
	lastMercHeal time.Time
	active       *atomic.Bool
}

func NewHealthManager(logger *zap.Logger, hr Repository, eventChan chan<- event.Event, beltManager BeltManager, cfg config.Config) Manager {
	return Manager{
		logger:      logger,
		hr:          hr,
		eventChan:   eventChan,
		beltManager: beltManager,
		cfg:         cfg,
		active:      atomic.NewBool(false),
	}
}

// Start will keep looking at life/mana levels from our character and mercenary and do best effort to keep them up
func (hm Manager) Start(ctx context.Context) error {
	ticker := time.NewTicker(monitorEvery)

	for {
		select {
		case <-ticker.C:
			if hm.active.Load() {
				hm.handleHealthAndMana()
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (hm Manager) Pause() {
	hm.active.Swap(false)
}

func (hm Manager) Resume() {
	hm.active.Swap(true)
}

func (hm Manager) handleHealthAndMana() {
	hpConfig := hm.cfg.Health
	status := hm.hr.CurrentStatus()

	usedRejuv := false
	if status.HPPercent() <= hpConfig.RejuvPotionAtLife || status.MPPercent() < hpConfig.RejuvPotionAtMana {
		hm.beltManager.DrinkPotion(data.RejuvenationPotion, false)
		usedRejuv = true
	}

	if !usedRejuv {
		if status.HPPercent() <= hpConfig.ChickenAt {
			hm.chicken(status)
			return
		}

		if status.HPPercent() <= hpConfig.HealingPotionAt && time.Since(hm.lastHeal) > healingInterval {
			hm.beltManager.DrinkPotion(data.HealingPotion, false)
			hm.lastHeal = time.Now()
		}

		if status.MPPercent() <= hpConfig.ManaPotionAt && time.Since(hm.lastMana) > manaInterval {
			hm.beltManager.DrinkPotion(data.ManaPotion, false)
			hm.lastMana = time.Now()
		}
	}

	// Mercenary
	if status.Merc.Alive {
		usedMercRejuv := false
		if status.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			hm.beltManager.DrinkPotion(data.RejuvenationPotion, true)
			usedMercRejuv = true
		}

		if !usedMercRejuv {
			if status.MercHPPercent() <= hpConfig.MercChickenAt {
				hm.chicken(status)
				return
			}

			if status.MercHPPercent() <= hpConfig.MercHealingPotionAt && time.Since(hm.lastMercHeal) > healingMercInterval {
				hm.beltManager.DrinkPotion(data.HealingPotion, true)
				hm.lastMercHeal = time.Now()
			}
		}
	}
}

func (hm Manager) chicken(status Status) {
	hm.logger.Warn(fmt.Sprintf("Chicken! Current Health: %d (%d percent)", status.Life, status.HPPercent()))
	helper.ExitGame(hm.eventChan)
}
