package step

import (
	"fmt"
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
			path, distance, _ := helper.GetPathToDestination(data, l.Position.X, l.Position.Y-2)
			if distance > 15 {
				helper.MoveThroughPath(path, 15, true)
				return nil
			}
			helper.MoveThroughPath(path, 0, false)
			hid.Click(hid.LeftButton)
			m.waitingForInteraction = true
			return nil
		}
	}

	return fmt.Errorf("area %s not found", m.area)
}
