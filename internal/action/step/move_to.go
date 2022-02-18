package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"time"
)

type MoveTo struct {
	basicStep
	pf       helper.PathFinderV2
	toX      int
	toY      int
	teleport bool
}

func NewMoveTo(toX, toY int, teleport bool, pf helper.PathFinderV2) *MoveTo {
	return &MoveTo{
		basicStep: newBasicStep(),
		pf:        pf,
		toX:       toX,
		toY:       toY,
		teleport:  teleport,
	}
}

func (m *MoveTo) Status(data game.Data) Status {
	_, distance, _ := m.pf.GetPathToDestination(data, m.toX, m.toY)
	if distance < 6 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveTo) Run(data game.Data) error {
	m.tryTransitionStatus(StatusInProgress)
	// TODO: In case of teleport, calculate fcr frames for waiting time
	if time.Since(m.lastRun) < time.Millisecond*500 {
		return nil
	}

	m.lastRun = time.Now()
	// TODO: Handle not found
	path, _, _ := m.pf.GetPathToDestination(data, m.toX, m.toY)
	m.pf.MoveThroughPath(path, 20, m.teleport)

	return nil
}
