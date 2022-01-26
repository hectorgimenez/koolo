package koolo

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"time"
)

const (
	monitorEvery        = time.Millisecond * 500
	healingInterval     = time.Second * 6
	healingMercInterval = time.Second * 6
	manaInterval        = time.Second * 4
)

type HealthRepository interface {
	CurrentStatus() (Status, error)
}

type Status struct {
	Life    int
	MaxLife int
	Mana    int
	MaxMana int
	Merc    MercStatus
}

type MercStatus struct {
	Alive   bool
	Life    int
	MaxLife int
}

func (s Status) HPPercent() int {
	return (s.Life / s.MaxLife) * 100
}

func (s Status) MPPercent() int {
	return (s.Mana / s.MaxMana) * 100
}

func (s Status) MercHPPercent() int {
	return (s.Merc.Life / s.Merc.MaxLife) * 100
}

// HealthManager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type HealthManager struct {
	hr           HealthRepository
	ah           chan<- action.Action
	cfg          config.Config
	lastHeal     time.Time
	lastMana     time.Time
	lastMercHeal time.Time
}

func NewHealthManager(hr HealthRepository, ah chan<- action.Action, cfg config.Config) HealthManager {
	return HealthManager{
		hr:  hr,
		ah:  ah,
		cfg: cfg,
	}
}

// Start will keep looking at life/mana levels from our character and mercenary and do best effort to keep them up
func (hm HealthManager) Start(ctx context.Context) error {
	ticker := time.NewTicker(monitorEvery)

	for {
		select {
		case <-ticker.C:
			hm.handleHealthAndMana()
		case <-ctx.Done():
			return nil
		}
	}
}

func (hm HealthManager) handleHealthAndMana() {
	hpConfig := hm.cfg.Health
	status, err := hm.hr.CurrentStatus()
	if err != nil {
		// TODO: Handle error
	}

	usedRejuv := false
	if status.HPPercent() <= hpConfig.RejuvPotionAtLife || status.MPPercent() < hpConfig.RejuvPotionAtMana {
		// TODO: Use Rejuv
		usedRejuv = true
	}

	if !usedRejuv {
		if status.HPPercent() <= hpConfig.ChickenAt {
			hm.chicken()
			return
		}

		if status.HPPercent() <= hpConfig.HealingPotionAt && time.Since(hm.lastHeal) > healingInterval {
			// TODO: Use Healing
			hm.lastHeal = time.Now()
		}

		if status.MPPercent() <= hpConfig.ManaPotionAt && time.Since(hm.lastMana) > manaInterval {
			// TODO: Use Mana
			hm.lastMana = time.Now()
		}
	}

	// Mercenary
	if status.Merc.Alive {
		usedMercRejuv := false
		if status.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			// TODO: Use Rejuv on Merc
			usedMercRejuv = true
		}

		if !usedMercRejuv {
			if status.MercHPPercent() <= hpConfig.MercChickenAt {
				hm.chicken()
				return
			}

			if status.MercHPPercent() <= hpConfig.MercHealingPotionAt && time.Since(hm.lastMercHeal) > healingMercInterval {
				// TODO Use Healing on Merc
				hm.lastMercHeal = time.Now()
			}
		}
	}
}

func (hm HealthManager) chicken() {
	a := action.NewAction(
		action.PriorityHigh,
		action.NewKeyPress("esc", time.Millisecond*500),
		action.NewKeyPress("down", time.Millisecond*50),
		action.NewKeyPress("enter", time.Millisecond*10),
	)
	hm.ah <- a
}
