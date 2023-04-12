package step

import (
	"errors"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type MoveToStep struct {
	pathingStep
	destination     data.Position
	teleport        bool
	stopAtDistance  int
	nearestWalkable bool
	timeout         time.Duration
	startedAt       time.Time
}

type MoveToStepOption func(step *MoveToStep)

func MoveTo(toX, toY int, teleport bool, opts ...MoveToStepOption) *MoveToStep {
	step := &MoveToStep{
		pathingStep: newPathingStep(),
		destination: data.Position{
			X: toX,
			Y: toY,
		},
		teleport: teleport,
		timeout:  time.Second * 30,
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

func WithTimeout(timeout time.Duration) MoveToStepOption {
	return func(step *MoveToStep) {
		step.timeout = timeout
	}
}

func (m *MoveToStep) Status(d data.Data) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	distance := pather.DistanceFromMe(d, m.destination)
	if distance < 5 || distance <= m.stopAtDistance {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToStep) Run(d data.Data) error {
	if m.teleport && m.status == StatusNotStarted {
		hid.PressKey(config.Config.Bindings.Teleport)
	}
	if m.startedAt.IsZero() {
		m.startedAt = time.Now()
	}

	m.tryTransitionStatus(StatusInProgress)

	if m.timeout > 0 && time.Since(m.startedAt) > m.timeout {
		m.tryTransitionStatus(StatusCompleted)
		return nil
	}

	// Throttle movement clicks in town
	if d.PlayerUnit.Area.IsTown() {
		if time.Since(m.lastRun) < helper.RandomDurationMs(200, 350) {
			return nil
		}
	}

	if m.teleport && time.Since(m.lastRun) < config.Config.Runtime.CastDuration {
		return nil
	}

	stuck := m.isPlayerStuck(d)

	if m.path == nil || !m.cachePath(d) || stuck {
		if stuck {
			if len(m.path.AstarPather) == 0 {
				pather.RandomMovement()
				m.lastRun = time.Now()

				return nil
			} else {
				tile := m.path.AstarPather[len(m.path.AstarPather)-1].(*pather.Tile)
				m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
			}
		}

		path, _, found := pather.GetPath(d, m.destination, m.blacklistedPositions...)
		if !found {
			// Try to find the nearest walkable place
			if m.nearestWalkable {
				path, _, found = pather.GetClosestWalkablePath(d, m.destination, m.blacklistedPositions...)
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
