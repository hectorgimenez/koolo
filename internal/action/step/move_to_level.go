package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type MoveToAreaStep struct {
	basicStep
	area                  game.Area
	waitingForInteraction bool
}

func MoveToLevel(area game.Area) *MoveToAreaStep {
	return &MoveToAreaStep{
		basicStep: newBasicStep(),
		area:      area,
	}
}

func (m *MoveToAreaStep) Status(data game.Data) Status {
	// Give some extra time to render the UI
	if data.Area == m.area && time.Since(m.lastRun) > time.Second*1 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToAreaStep) Run(data game.Data) error {
	if m.status == StatusNotStarted {
		hid.PressKey(config.Config.Bindings.Teleport)
	}
	m.tryTransitionStatus(StatusInProgress)
	if time.Since(m.lastRun) < time.Millisecond*500 {
		return nil
	}

	if m.waitingForInteraction && time.Since(m.lastRun) < time.Second*3 {
		return nil
	}

	m.lastRun = time.Now()
	for _, l := range data.AdjacentLevels {
		if l.Area == m.area {
			distance := helper.DistanceFromPoint(data, l.Position.X, l.Position.Y)
			if distance > 10 {
				path, _, _ := helper.GetPathToDestination(data, l.Position.X, l.Position.Y)
				helper.MoveThroughPath(path, 25, true)
				return nil
			}

			x, y := helper.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, l.Position.X, l.Position.Y)
			hid.MovePointer(x, y)
			helper.Sleep(100)
			hid.Click(hid.LeftButton)
			m.waitingForInteraction = true
			return nil
		}
	}

	return fmt.Errorf("area %s not found", m.area)
}
