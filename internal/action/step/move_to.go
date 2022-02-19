package step

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type MoveTo struct {
	basicStep
	toX      int
	toY      int
	teleport bool
}

func NewMoveTo(toX, toY int, teleport bool) *MoveTo {
	return &MoveTo{
		basicStep: newBasicStep(),
		toX:       toX,
		toY:       toY,
		teleport:  teleport,
	}
}

func (m *MoveTo) Status(data game.Data) Status {
	_, distance, _ := helper.GetPathToDestination(data, m.toX, m.toY)
	if distance < 6 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveTo) Run(data game.Data) error {
	if m.teleport && m.status == StatusNotStarted {
		hid.PressKey(config.Config.Bindings.Teleport)
	}

	m.tryTransitionStatus(StatusInProgress)
	// TODO: In case of teleport, calculate fcr frames for waiting time
	if time.Since(m.lastRun) < time.Millisecond*500 {
		return nil
	}

	m.lastRun = time.Now()
	// TODO: Handle not found
	path, _, _ := helper.GetPathToDestination(data, m.toX, m.toY)
	helper.MoveThroughPath(path, 20, m.teleport)

	return nil
}
