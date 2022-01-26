package health

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	koolo "github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/inventory"
	"go.uber.org/atomic"
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
	hr           Repository
	actionChan   chan<- action.Action
	beltManager  BeltManager
	gm           koolo.GameManager
	cfg          config.Config
	lastHeal     time.Time
	lastMana     time.Time
	lastMercHeal time.Time
	active       *atomic.Bool
}

func NewHealthManager(hr Repository, actionChan chan<- action.Action, beltManager BeltManager, gm koolo.GameManager, cfg config.Config) Manager {
	return Manager{
		hr:          hr,
		actionChan:  actionChan,
		beltManager: beltManager,
		gm:          gm,
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
	status, err := hm.hr.CurrentStatus()
	if err != nil {
		// TODO: Handle error
	}

	usedRejuv := false
	if status.HPPercent() <= hpConfig.RejuvPotionAtLife || status.MPPercent() < hpConfig.RejuvPotionAtMana {
		hm.beltManager.DrinkPotion(inventory.RejuvenationPotion, false)
		usedRejuv = true
	}

	if !usedRejuv {
		if status.HPPercent() <= hpConfig.ChickenAt {
			hm.chicken(status)
			return
		}

		if status.HPPercent() <= hpConfig.HealingPotionAt && time.Since(hm.lastHeal) > healingInterval {
			hm.beltManager.DrinkPotion(inventory.HealingPotion, false)
			hm.lastHeal = time.Now()
		}

		if status.MPPercent() <= hpConfig.ManaPotionAt && time.Since(hm.lastMana) > manaInterval {
			hm.beltManager.DrinkPotion(inventory.ManaPotion, false)
			hm.lastMana = time.Now()
		}
	}

	// Mercenary
	if status.Merc.Alive {
		usedMercRejuv := false
		if status.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			hm.beltManager.DrinkPotion(inventory.RejuvenationPotion, true)
			usedMercRejuv = true
		}

		if !usedMercRejuv {
			if status.MercHPPercent() <= hpConfig.MercChickenAt {
				hm.chicken(status)
				return
			}

			if status.MercHPPercent() <= hpConfig.MercHealingPotionAt && time.Since(hm.lastMercHeal) > healingMercInterval {
				hm.beltManager.DrinkPotion(inventory.HealingPotion, true)
				hm.lastMercHeal = time.Now()
			}
		}
	}
}

func (hm Manager) chicken(status Status) {
	// TODO: Print status
	hm.gm.ExitGame()
}
