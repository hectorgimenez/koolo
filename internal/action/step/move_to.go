package step

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/container"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type MoveToStep struct {
	pathingStep
	destination     data.Position
	stopAtDistance  int
	nearestWalkable bool
	timeout         time.Duration
	startedAt       time.Time
	cfg             *config.CharacterCfg
}

type MoveToStepOption func(step *MoveToStep)

func MoveTo(cfg *config.CharacterCfg, destination data.Position, opts ...MoveToStepOption) *MoveToStep {
	step := &MoveToStep{
		pathingStep: newPathingStep(),
		destination: destination,
		timeout:     time.Second * 30,
		cfg:         cfg,
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

func (m *MoveToStep) Status(d data.Data, _ container.Container) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	distance := pather.DistanceFromMe(d, m.destination)
	if distance < 5 || distance < m.stopAtDistance {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToStep) Run(d data.Data, container container.Container) error {
	// Press the Teleport keybinding if it's available, otherwise use vigor (if available)
	if helper.CanTeleport(d) {
		if d.PlayerUnit.RightSkill != skill.Teleport {
			container.HID.PressKey(m.cfg.Bindings.Teleport)
		}
	} else if d.PlayerUnit.Skills[skill.Vigor].Level > 0 && m.cfg.Bindings.Paladin.Vigor != "" {
		if d.PlayerUnit.RightSkill != skill.Vigor {
			container.HID.PressKey(m.cfg.Bindings.Paladin.Vigor)
		}
	}

	if m.startedAt.IsZero() {
		m.startedAt = time.Now()
	}

	m.tryTransitionStatus(StatusInProgress)

	if m.timeout > 0 && time.Since(m.startedAt) > m.timeout {
		m.tryTransitionStatus(StatusCompleted)
		return nil
	}

	// Add some delay between clicks to let the character move to destination
	walkDuration := helper.RandomDurationMs(600, 1200)
	if !helper.CanTeleport(d) && time.Since(m.lastRun) < walkDuration {
		return nil
	}

	if helper.CanTeleport(d) && time.Since(m.lastRun) < m.cfg.Runtime.CastDuration {
		return nil
	}

	stuck := m.isPlayerStuck(d)

	if m.path == nil || !m.cachePath(d) || stuck {
		if stuck {
			if len(m.path.AstarPather) == 0 {
				container.PathFinder.RandomMovement()
				m.lastRun = time.Now()

				return nil
			} else {
				tile := m.path.AstarPather[len(m.path.AstarPather)-1].(*pather.Tile)
				m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
			}
		}

		path, _, found := container.PathFinder.GetPath(d, m.destination, m.blacklistedPositions...)
		if !found {
			// Try to find the nearest walkable place
			if m.nearestWalkable {
				path, _, found = container.PathFinder.GetClosestWalkablePath(d, m.destination, m.blacklistedPositions...)
				if !found {
					return errors.New("path could not be calculated, maybe there is an obstacle or a flying platform (arcane sanctuary)")
				}
				m.destination = path.Destination
			} else {
				return errors.New("path could not be calculated, maybe there is an obstacle or a flying platform (arcane sanctuary)")
			}
		}
		m.path = path
	}

	m.lastRun = time.Now()
	m.previousArea = d.PlayerUnit.Area
	if len(m.path.AstarPather) == 0 {
		return nil
	}
	container.PathFinder.MoveThroughPath(m.path, calculateMaxDistance(d, walkDuration), helper.CanTeleport(d))

	return nil
}

func (m *MoveToStep) Reset() {
	m.status = StatusNotStarted
	m.lastRun = time.Time{}
	m.startedAt = time.Time{}
}

func calculateMaxDistance(d data.Data, duration time.Duration) int {
	// We don't care too much if teleport is available, we can ignore corners, 90 degrees turns, etc
	if helper.CanTeleport(d) {
		return 25
	}

	// Calculate the distance we can walk in the given duration, based on the randomized time
	proposedDistance := int(float64(25) * duration.Seconds())
	realDistance := proposedDistance

	return realDistance
}
