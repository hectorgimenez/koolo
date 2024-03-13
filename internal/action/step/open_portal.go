package step

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type OpenPortalStep struct {
	basicStep
	tpKB string
}

func OpenPortal(tpKB string) *OpenPortalStep {
	return &OpenPortalStep{
		basicStep: newBasicStep(),
		tpKB:      tpKB,
	}
}

func (s *OpenPortalStep) Status(d data.Data, _ container.Container) Status {
	if s.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time, sometimes if we move the mouse over the portal before is shown
	// and there is an intractable entity behind it, will keep it focused
	if time.Since(s.LastRun()) > time.Second*1 {
		for _, o := range d.Objects {
			if o.IsPortal() {
				return s.tryTransitionStatus(StatusCompleted)
			}
		}
	}

	return StatusInProgress
}

func (s *OpenPortalStep) Run(_ data.Data, container container.Container) error {
	// Give some time to portal to popup before retrying...
	if time.Since(s.LastRun()) < time.Second*2 {
		return nil
	}

	container.HID.PressKey(s.tpKB)
	helper.Sleep(250)
	container.HID.Click(game.RightButton, 300, 300)
	s.lastRun = time.Now()

	return nil
}
