package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type InteractObject struct {
	basicStep
	objectName            string
	pf                    helper.PathFinderV2
	waitingForInteraction bool
	isCompleted           func(game.Data) bool
}

func NewInteractObject(objectName string, isCompleted func(game.Data) bool, pf helper.PathFinderV2) *InteractObject {
	return &InteractObject{
		basicStep: basicStep{
			status: StatusNotStarted,
		},
		objectName:  objectName,
		pf:          pf,
		isCompleted: isCompleted,
	}
}

func (i *InteractObject) Status(data game.Data) Status {
	// Give some extra time to render the UI
	if i.isCompleted(data) && time.Since(i.lastRun) > time.Second*1 {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractObject) Run(data game.Data) error {
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
				path, distance, _ := i.pf.GetPathToDestination(data, o.Position.X, o.Position.Y)
				if distance > 15 {
					i.pf.MoveThroughPath(path, 15, false)
					return nil
				}
				i.pf.MoveThroughPath(path, 0, false)

				return nil
			}
		}
	}

	return fmt.Errorf("object %s not found", i.objectName)
}
