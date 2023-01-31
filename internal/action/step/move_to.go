package step

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type MoveToStep struct {
	pathingStep
	destination     game.Position
	teleport        bool
	stopAtDistance  int
	nearestWalkable bool
}

type MoveToStepOption func(step *MoveToStep)

func MoveTo(toX, toY int, teleport bool, opts ...MoveToStepOption) *MoveToStep {
	step := &MoveToStep{
		pathingStep: newPathingStep(),
		destination: game.Position{
			X: toX,
			Y: toY,
		},
		teleport: teleport,
	}

	for _, o := range opts {
		o(step)
	}

	return step
}

func StopAtDistance(distance int) MoveToStepOption {
	return func(step *MoveToStep) {
		step.stopAtDistance = distance
	}
}

func ClosestWalkable() MoveToStepOption {
	return func(step *MoveToStep) {
		step.nearestWalkable = true
	}
}

func (m *MoveToStep) Status(data game.Data) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	distance := pather.DistanceFromMe(data, m.destination)
	if distance < 5 || distance <= m.stopAtDistance {
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
		if time.Since(m.lastRun) < helper.RandomDurationMs(200, 350) {
			return nil
		}
	}

	if m.teleport && time.Since(m.lastRun) < config.Config.Runtime.CastDuration {
		return nil
	}

	stuck := m.isPlayerStuck(data)

	if m.path == nil || !m.cachePath(data) || stuck {
		if stuck {
			tile := m.path.AstarPather[len(m.path.AstarPather)-1].(*pather.Tile)
			m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
		}

		path, _, found := pather.GetPath(data, m.destination, m.blacklistedPositions...)
		if !found {
			// Try to find the nearest walkable place
			if m.nearestWalkable {
				path, _, found = pather.GetClosestWalkablePath(data, m.destination, m.blacklistedPositions...)
				if !found {
					return errors.New("path could not be calculated, maybe there is an obstacle or a flying platform (arcane sanctuary)")
				}
				m.destination = path.Destination
			}

		}
		m.path = path
	}

	m.lastRun = time.Now()
	if len(m.path.AstarPather) == 0 {
		return nil
	}
	pather.MoveThroughPath(m.path, 25, m.teleport)

	return nil
}
