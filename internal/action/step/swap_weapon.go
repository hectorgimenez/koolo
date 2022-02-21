package step

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type SwapWeaponStep struct {
	basicStep
	binding             string
	initialWeaponWasCTA bool
}

func SwapWeapon() *SwapWeaponStep {
	return &SwapWeaponStep{
		basicStep: newBasicStep(),
		binding:   config.Config.Bindings.SwapWeapon,
	}
}

func (s *SwapWeaponStep) Status(data game.Data) Status {
	if s.status == StatusNotStarted {
		return StatusNotStarted
	}

	_, found := data.PlayerUnit.Skills[game.SkillBattleOrders]
	if found && s.initialWeaponWasCTA || !found && !s.initialWeaponWasCTA {
		return s.tryTransitionStatus(StatusInProgress)
	}

	return s.tryTransitionStatus(StatusCompleted)
}

func (s *SwapWeaponStep) Run(data game.Data) error {
	// Add some delay to let the weapon switch
	if time.Since(s.lastRun) < time.Second {
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
