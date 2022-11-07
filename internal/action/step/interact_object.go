package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type InteractObjectStep struct {
	pathingStep
	objectName            object.Name
	waitingForInteraction bool
	isCompleted           func(game.Data) bool
	mouseOverAttempts     int
}

func InteractObject(name object.Name, isCompleted func(game.Data) bool) *InteractObjectStep {
	return &InteractObjectStep{
		pathingStep: newPathingStep(),
		objectName:  name,
		isCompleted: isCompleted,
	}
}

func (i *InteractObjectStep) Status(data game.Data) Status {
	// Give some extra time to render the UI
	if time.Since(i.lastRun) > time.Second*1 && i.isCompleted(data) {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractObjectStep) Run(data game.Data) error {
	if i.isCompleted(data) {
		return nil
	}

	// Throttle movement clicks
	if time.Since(i.lastRun) < helper.RandomDurationMs(300, 600) {
		return nil
	}

	if i.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("object %s could not be interacted", i.objectName)
	}

	if i.consecutivePathNotFound >= maxPathNotFoundRetries {
		return fmt.Errorf("error moving to %s: %w", i.objectName, errPathNotFound)
	}

	i.tryTransitionStatus(StatusInProgress)

	// Give some time before retrying the interaction
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second*3 {
		return nil
	}

	for _, o := range data.Objects {
		if o.Name == i.objectName {
			if o.IsHovered {
				hid.Click(hid.LeftButton)
				i.waitingForInteraction = true
				i.lastRun = time.Now()
				return nil
			} else {
				distance := pather.DistanceFromPoint(data, o.Position.X, o.Position.Y)

				if distance > 15 {
					path, _, found := pather.GetPathToDestination(data, o.Position.X, o.Position.Y)
					if !found {
						pather.RandomMovement()
						i.consecutivePathNotFound++
						return nil
					}
					i.consecutivePathNotFound = 0
					pather.MoveThroughPath(path, helper.RandRng(7, 17), false)
					i.lastRun = time.Now()
					return nil
				}
				if time.Since(i.lastRun) < time.Millisecond*200 {
					return nil
				}

				objectX := o.Position.X - 2
				objectY := o.Position.Y - 2
				mX, mY := pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, objectX, objectY)

				x, y := helper.Spiral(i.mouseOverAttempts)
				hid.MovePointer(mX+x, mY+y)
				i.mouseOverAttempts++

				i.lastRun = time.Now()
				return nil
			}
		}
	}

	i.lastRun = time.Now()
	return fmt.Errorf("object %s not found", i.objectName)
}
