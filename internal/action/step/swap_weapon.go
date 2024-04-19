package step

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/skill"
)

type SwapWeaponStep struct {
	basicStep
	wantCTA bool
}

func SwapToMainWeapon() *SwapWeaponStep {
	return &SwapWeaponStep{
		basicStep: newBasicStep(),
	}
}

func SwapToCTA() *SwapWeaponStep {
	return &SwapWeaponStep{
		basicStep: newBasicStep(),
		wantCTA:   true,
	}
}

func (s *SwapWeaponStep) Status(d game.Data, _ container.Container) Status {
	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if (s.wantCTA && found) || (!s.wantCTA && !found) {
		return s.tryTransitionStatus(StatusCompleted)
	}

	return s.status
}

func (s *SwapWeaponStep) Run(d game.Data, container container.Container) error {
	s.tryTransitionStatus(StatusInProgress)

	if time.Since(s.lastRun) < time.Second {
		return nil
	}

	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if (s.wantCTA && found) || (!s.wantCTA && !found) {
		s.tryTransitionStatus(StatusCompleted)

		return nil
	}

	container.HID.PressKeyBinding(d.KeyBindings.SwapWeapons)

	s.lastRun = time.Now()

	return nil
}
