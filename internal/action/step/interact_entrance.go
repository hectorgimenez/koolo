package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type InteractEntranceStep struct {
	basicStep
	area                  area.Area
	waitingForInteraction bool
	mouseOverAttempts     int
}

func InteractEntrance(area area.Area) *InteractEntranceStep {
	return &InteractEntranceStep{
		basicStep: newBasicStep(),
		area:      area,
	}
}

func (m *InteractEntranceStep) Status(d data.Data) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time to render the UI
	if d.PlayerUnit.Area == m.area && time.Since(m.lastRun) > time.Second*1 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *InteractEntranceStep) Run(d data.Data) error {
	m.tryTransitionStatus(StatusInProgress)

	if m.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("area %d could not be interacted", m.area)
	}

	if (m.waitingForInteraction && time.Since(m.lastRun) < time.Second*1) || d.PlayerUnit.Area == m.area {
		return nil
	}

	m.lastRun = time.Now()
	for _, l := range d.AdjacentLevels {
		if l.Area == m.area {
			distance := pather.DistanceFromMe(d, l.Position)
			if distance > 10 {
				return errors.New("entrance too far away")
			}

			if l.IsEntrance {
				if d.HoverData.UnitType == 5 || d.HoverData.UnitType == 2 && d.HoverData.IsHovered {
					hid.Click(hid.LeftButton)
					m.waitingForInteraction = true
				}

				lx, ly := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, l.Position.X-2, l.Position.Y-2)
				x, y := helper.Spiral(m.mouseOverAttempts)
				hid.MovePointer(lx+x, ly+y)
				m.mouseOverAttempts++
				return nil
			}

			return fmt.Errorf("area %d is not an entrance", m.area)
		}
	}

	return fmt.Errorf("area %d not found", m.area)
}
