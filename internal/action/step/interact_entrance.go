package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"

	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type InteractEntranceStep struct {
	basicStep
	area                  area.ID
	waitingForInteraction bool
	mouseOverAttempts     int
	currentMouseCoords    data.Position
}

func InteractEntrance(area area.ID) *InteractEntranceStep {
	return &InteractEntranceStep{
		basicStep: newBasicStep(),
		area:      area,
	}
}

func (m *InteractEntranceStep) Status(d game.Data, _ container.Container) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time to render the UI
	if d.PlayerUnit.Area == m.area && time.Since(m.lastRun) > time.Second*1 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *InteractEntranceStep) Run(d game.Data, container container.Container) error {
	m.tryTransitionStatus(StatusInProgress)

	if m.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("area %s [%d] could not be interacted", m.area.Area().Name, m.area)
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
				lx, ly := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, l.Position.X-2, l.Position.Y-2)
				if d.HoverData.UnitType == 5 || d.HoverData.UnitType == 2 && d.HoverData.IsHovered {
					container.HID.Click(game.LeftButton, m.currentMouseCoords.X, m.currentMouseCoords.Y)
					m.waitingForInteraction = true
				}

				x, y := helper.Spiral(m.mouseOverAttempts)
				m.currentMouseCoords = data.Position{X: lx + x, Y: ly + y}
				container.HID.MovePointer(lx+x, ly+y)
				m.mouseOverAttempts++
				return nil
			}

			return fmt.Errorf("area %s [%d]  is not an entrance", m.area.Area().Name, m.area)
		}
	}

	return fmt.Errorf("area %s [%d]  not found", m.area.Area().Name, m.area)
}
