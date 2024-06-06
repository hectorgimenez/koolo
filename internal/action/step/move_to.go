package step

import (
	"errors"
	"time"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
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
}

type MoveToStepOption func(step *MoveToStep)

func MoveTo(destination data.Position, opts ...MoveToStepOption) *MoveToStep {
	step := &MoveToStep{
		pathingStep:    newPathingStep(),
		destination:    destination,
		timeout:        time.Second * 30,
		stopAtDistance: 7,
	}

	for _, o := range opts {
		o(step)
	}

	return step
}

func StopAtDistance(distance int) MoveToStepOption {
	return func(step *MoveToStep) {
		step.stopAtDistance = distance + 5 // Add some padding, origin and destination point are not walkable and should be ignored
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

func (m *MoveToStep) Status(d game.Data, container container.Container) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	distance := pather.DistanceFromMe(d, m.destination)
	if distance < m.stopAtDistance {
		// In case distance is lower, we double-check with the pathfinder and the full path instead of euclidean distance
		_, distance, found := container.PathFinder.GetPath(d, m.destination)
		if !found || distance < m.stopAtDistance {
			return m.tryTransitionStatus(StatusCompleted)
		}
	}

	return m.status
}

func (m *MoveToStep) Run(d game.Data, container container.Container) error {
	// Press the Teleport keybinding if it's available, otherwise use vigor (if available)
	if d.CanTeleport() {
		if d.PlayerUnit.RightSkill != skill.Teleport {
			container.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(skill.Teleport))
		}
	} else if kb, found := d.KeyBindings.KeyBindingForSkill(skill.Vigor); found {
		if d.PlayerUnit.RightSkill != skill.Vigor {
			container.HID.PressKeyBinding(kb)
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
	if !d.CanTeleport() && time.Since(m.lastRun) < walkDuration {
		return nil
	}

	if d.CanTeleport() && time.Since(m.lastRun) < d.PlayerCastDuration() {
		return nil
	}

	stuck := m.isPlayerStuck(d)

	if m.path == nil || !m.cachePath(d) || stuck {
		if stuck {
			if len(m.path) == 0 {
				randomPosX, randomPosY := pather.FindFirstWalkable(d.PlayerUnit.Position, d.AreaOrigin, d.CollisionGrid, 15)
				screenX, screenY := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, randomPosX, randomPosY)
				container.PathFinder.MoveCharacter(d, screenX, screenY)
				m.lastRun = time.Now()

				return nil
			} else {
				tile := m.path[len(m.path)-1].(*pather.Tile)
				m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
			}
		}

		path, _, found := container.PathFinder.GetClosestWalkablePath(d, m.destination, m.blacklistedPositions...)
		if !found {
			if pather.DistanceFromMe(d, m.destination) < m.stopAtDistance+5 {
				m.tryTransitionStatus(StatusCompleted)
				return nil
			}

			return errors.New("path could not be calculated, maybe there is an obstacle or a flying platform (arcane sanctuary)")
		}
		m.path = path
	}

	m.lastRun = time.Now()
	m.previousArea = d.PlayerUnit.Area
	if len(m.path) == 0 {
		return nil
	}
	container.PathFinder.MoveThroughPath(d, m.path, calculateMaxDistance(d, walkDuration))

	return nil
}

func (m *MoveToStep) Reset() {
	m.status = StatusNotStarted
	m.lastRun = time.Time{}
	m.startedAt = time.Time{}
}

func calculateMaxDistance(d game.Data, duration time.Duration) int {
	// We don't care too much if teleport is available, we can ignore corners, 90 degrees turns, etc
	if d.CanTeleport() {
		return 25
	}

	// Calculate the distance we can walk in the given duration, based on the randomized time
	proposedDistance := int(float64(25) * duration.Seconds())
	realDistance := proposedDistance

	return realDistance
}
