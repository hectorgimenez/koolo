package step

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/config"
)

type SwapWeaponStep struct {
	basicStep
	binding string
	wantCTA bool
}

func SwapToMainWeapon() *SwapWeaponStep {
	return &SwapWeaponStep{
		basicStep: newBasicStep(),
		binding:   config.Config.Bindings.SwapWeapon,
	}
}

func SwapToCTA() *SwapWeaponStep {
	return &SwapWeaponStep{
		basicStep: newBasicStep(),
		binding:   config.Config.Bindings.SwapWeapon,
		wantCTA:   true,
	}
}

func (s *SwapWeaponStep) Status(d data.Data, _ container.Container) Status {
	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if (s.wantCTA && found) || (!s.wantCTA && !found) {
		return s.tryTransitionStatus(StatusCompleted)
	}

	return s.status
}

func (s *SwapWeaponStep) Run(d data.Data, container container.Container) error {
	s.tryTransitionStatus(StatusInProgress)

	if time.Since(s.lastRun) < time.Second {
		return nil
	}

	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if (s.wantCTA && found) || (!s.wantCTA && !found) {
		s.tryTransitionStatus(StatusCompleted)

		return nil
	}

	container.HID.PressKey(s.binding)

	s.lastRun = time.Now()

	return nil
}
