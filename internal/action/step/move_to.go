package step

import (
	"errors"
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type MoveToStep struct {
	basicStep
	toX                  int
	toY                  int
	teleport             bool
	path                 []astar.Pather
	lastRunPositions     [][2]int
	blacklistedPositions [][2]int
}

func MoveTo(toX, toY int, teleport bool) *MoveToStep {
	return &MoveToStep{
		basicStep: newBasicStep(),
		toX:       toX,
		toY:       toY,
		teleport:  teleport,
	}
}

func (m *MoveToStep) Status(data game.Data) Status {
	distance := pather.DistanceFromMe(data, m.toX, m.toY)
	if distance < 5 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToStep) Run(data game.Data) error {
	if m.teleport && m.status == StatusNotStarted {
		hid.PressKey(config.Config.Bindings.Teleport)
	}
	m.tryTransitionStatus(StatusInProgress)

	// Throttle movement clicks in town
	if data.PlayerUnit.Area.IsTown() {
		if time.Since(m.lastRun) < helper.RandomDurationMs(300, 600) {
			return nil
		}
	}

	if m.teleport && time.Since(m.lastRun) < config.Config.Runtime.CastDuration {
		return nil
	}

	stuck := m.isPlayerStuck(data)

	if m.path == nil || !m.adjustPath(data) || stuck {
		if stuck {
			tile := m.path[len(m.path)-1].(*pather.Tile)
			m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
		}

		path, _, found := pather.GetPathToDestination(data, m.toX, m.toY, m.blacklistedPositions...)
		if !found {
			return errors.New("path chould not be calculated, maybe there is an obstacle or a flying platform (arcane sanctuary)")
		}
		m.path = path
	}

	m.lastRun = time.Now()
	pather.MoveThroughPath(m.path, 25, m.teleport)

	return nil
}

// Cache the path and try to reuse it
func (m *MoveToStep) adjustPath(data game.Data) bool {
	nearestKey := 0
	nearestDistance := 99999999
	for k, pos := range m.path {
		distance := pather.DistanceFromMe(data, pos.(*pather.Tile).X+data.AreaOrigin.X, pos.(*pather.Tile).Y+data.AreaOrigin.Y)
		if distance < nearestDistance {
			nearestDistance = distance
			nearestKey = k
		}
	}

	if nearestDistance < 5 && len(m.path) > nearestKey {
		//fmt.Println(fmt.Sprintf("Max deviation: %d, using Path Key: %d [%d]", nearestDistance, nearestKey, len(m.path)-1))
		m.path = m.path[:nearestKey]

		return true
	}

	return false
}

func (m *MoveToStep) isPlayerStuck(data game.Data) bool {
	m.lastRunPositions = append(m.lastRunPositions, [2]int{data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y})
	if len(m.lastRunPositions) > 3 {
		m.lastRunPositions = m.lastRunPositions[1:]
	} else {
		return false
	}

	for _, pos := range m.lastRunPositions {
		if pos[0] != data.PlayerUnit.Position.X || pos[1] != data.PlayerUnit.Position.Y {
			return false
		}
	}

	return true
}
