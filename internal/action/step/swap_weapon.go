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

func (s *SwapWeaponStep) Status(d data.Data) Status {
	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if (s.wantCTA && found) || (!s.wantCTA && !found) {
		s.tryTransitionStatus(StatusCompleted)
	}

	return s.status
}

func (s *SwapWeaponStep) Run(d data.Data) error {
	s.tryTransitionStatus(StatusInProgress)

	s.lastRun = time.Now()

	_, found := d.PlayerUnit.Skills[skill.BattleOrders]
	if (s.wantCTA && found) || (!s.wantCTA && !found) {
		s.tryTransitionStatus(StatusCompleted)

		return nil
	}

	s.lastRun = time.Now()
	hid.PressKey(s.binding)

	return nil
}
