package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type InteractObjectStep struct {
	basicStep
	objectName            string
	waitingForInteraction bool
	isCompleted           func(game.Data) bool
}

func InteractObject(objectName string, isCompleted func(game.Data) bool) *InteractObjectStep {
	return &InteractObjectStep{
		basicStep: basicStep{
			status: StatusNotStarted,
		},
		objectName:  objectName,
		isCompleted: isCompleted,
	}
}

func (i *InteractObjectStep) Status(data game.Data) Status {
	// Give some extra time to render the UI
	if i.isCompleted(data) && time.Since(i.lastRun) > time.Second*1 {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractObjectStep) Run(data game.Data) error {
	i.tryTransitionStatus(StatusInProgress)
	if time.Since(i.lastRun) < time.Millisecond*500 {
		return nil
	}

	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second*3 {
		return nil
	}

	i.lastRun = time.Now()
	for _, o := range data.Objects {
		if o.Name == i.objectName {
			if o.IsHovered {
				hid.Click(hid.LeftButton)
				i.waitingForInteraction = true
				return nil
			} else {
				distance := helper.DistanceFromPoint(data, o.Position.X, o.Position.Y)

				if distance > 15 {
					path, _, _ := helper.GetPathToDestination(data, o.Position.X, o.Position.Y)
					helper.MoveThroughPath(path, 15, false)
					return nil
				}
				x, y := helper.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, o.Position.X-2, o.Position.Y-2)
				hid.MovePointer(x, y)

				return nil
			}
		}
	}

	return fmt.Errorf("object %s not found", i.objectName)
}
