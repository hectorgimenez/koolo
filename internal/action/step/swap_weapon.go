package step

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/hid"
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

func (s *SwapWeaponStep) Status(d data.Data) Status {
	if s.status == StatusNotStarted {
		return StatusNotStarted
	}
	if s.status == StatusCompleted {
		return StatusCompleted
	}
	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if found && s.initialWeaponWasCTA || !found && !s.initialWeaponWasCTA {
		return s.tryTransitionStatus(StatusInProgress)
	}

	return s.tryTransitionStatus(StatusCompleted)
}

func (s *SwapWeaponStep) Run(d data.Data) error {
	// Add some delay to let the weapon switch
	if time.Since(s.lastRun) < time.Second {
		return nil
	}

	s.tryTransitionStatus(StatusInProgress)
	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if found {
		s.initialWeaponWasCTA = true
	}

	s.lastRun = time.Now()
	hid.PressKey(s.binding)

	return nil
}
