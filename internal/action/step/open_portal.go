package step

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type OpenPortalStep struct {
	basicStep
}

func OpenPortal() *OpenPortalStep {
	return &OpenPortalStep{basicStep: newBasicStep()}
}

func (s *OpenPortalStep) Status(data game.Data) Status {
	if s.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time, sometimes if we move the mouse over the portal before is shown
	// and there is an intractable entity behind it, will keep it focused
	if time.Since(s.LastRun()) > time.Second*1 {
		for _, o := range data.Objects {
			if o.IsPortal() {
				return s.tryTransitionStatus(StatusCompleted)
			}
		}
	}

	return StatusInProgress
}

func (s *OpenPortalStep) Run(_ game.Data) error {
	// Give some time to portal to popup before retrying...
	if time.Since(s.LastRun()) < time.Second*3 {
		return nil
	}

	hid.PressKey(config.Config.Bindings.TP)
	helper.Sleep(250)
	hid.Click(hid.RightButton)
	s.lastRun = time.Now()

	return nil
}
