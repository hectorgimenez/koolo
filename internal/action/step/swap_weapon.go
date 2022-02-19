package step

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type SwapWeapon struct {
	basicStep
	binding             string
	initialWeaponWasCTA bool
}

func NewSwapWeapon(cfg config.Config) *SwapWeapon {
	return &SwapWeapon{
		basicStep: newBasicStep(),
		binding:   cfg.Bindings.SwapWeapon,
	}
}

func (s *SwapWeapon) Status(data game.Data) Status {
	if s.status == StatusNotStarted {
		return StatusNotStarted
	}
	// Add some delay to let the weapon switch
	if time.Since(s.lastRun) < time.Second {
		return s.status
	}

	_, found := data.PlayerUnit.Skills[game.SkillBattleOrders]
	if found && s.initialWeaponWasCTA || !found && !s.initialWeaponWasCTA {
		s.tryTransitionStatus(StatusInProgress)
	}

	return s.tryTransitionStatus(StatusCompleted)
}

func (s *SwapWeapon) Run(data game.Data) error {
	// Add some delay to let the weapon switch
	if time.Since(s.lastRun) < time.Second*2 {
		return nil
	}

	s.tryTransitionStatus(StatusInProgress)
	_, found := data.PlayerUnit.Skills[game.SkillBattleOrders]
	if found {
		s.initialWeaponWasCTA = true
	}

	s.lastRun = time.Now()
	hid.PressKey(s.binding)

	return nil
}
