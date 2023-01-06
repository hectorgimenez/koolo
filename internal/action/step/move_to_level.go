package step

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type MoveToAreaStep struct {
	pathingStep
	area                  area.Area
	waitingForInteraction bool
}

func MoveToLevel(area area.Area) *MoveToAreaStep {
	return &MoveToAreaStep{
		pathingStep: newPathingStep(),
		area:        area,
	}
}

func (m *MoveToAreaStep) Status(data game.Data) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time to render the UI
	if data.PlayerUnit.Area == m.area && time.Since(m.lastRun) > time.Second*1 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToAreaStep) Run(data game.Data) error {
	if m.status == StatusNotStarted {
		hid.PressKey(config.Config.Bindings.Teleport)
	}
	m.tryTransitionStatus(StatusInProgress)
	if time.Since(m.lastRun) < config.Config.Runtime.CastDuration {
		return nil
	}

	if m.waitingForInteraction && time.Since(m.lastRun) < time.Second*1 {
		return nil
	}

	m.lastRun = time.Now()
	for _, l := range data.AdjacentLevels {
		if l.Area == m.area {
			distance := pather.DistanceFromMe(data, l.Position.X, l.Position.Y)
			if distance > 5 {
				stuck := m.isPlayerStuck(data)
				if m.path == nil || !m.cachePath(data) || stuck {
					if stuck {
						tile := m.path[len(m.path)-1].(*pather.Tile)
						m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
					}
					path, _, found := pather.GetPathToDestination(data, l.Position.X, l.Position.Y)
					if !found {
						return errors.New("path could not be calculated, maybe there is an obstacle")
					}
					m.path = path
				}
				pather.MoveThroughPath(m.path, 25, true)
				return nil
			}

			x, y := pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, l.Position.X-2, l.Position.Y-2)
			hid.MovePointer(x, y)
			helper.Sleep(100)
			hid.Click(hid.LeftButton)
			m.waitingForInteraction = true
			return nil
		}
	}

	return fmt.Errorf("area %s not found", m.area)
}
