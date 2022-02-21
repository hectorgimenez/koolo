package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type InteractObjectStep struct {
	pathingStep
	objectName            string
	waitingForInteraction bool
	isCompleted           func(game.Data) bool
}

func InteractObject(objectName string, isCompleted func(game.Data) bool) *InteractObjectStep {
	return &InteractObjectStep{
		pathingStep: newPathingStep(),
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
	if i.consecutivePathNotFound >= maxPathNotFoundRetries {
		return fmt.Errorf("error moving to %s: %w", i.objectName, errPathNotFound)
	}

	i.tryTransitionStatus(StatusInProgress)
	// Throttle movement clicks
	if time.Since(i.lastRun) < time.Millisecond*350 {
		return nil
	}

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
					pather.MoveThroughPath(path, 12, false)
					i.lastRun = time.Now()
					return nil
				}
				if time.Since(i.lastRun) < time.Second {
					return nil
				}
				x, y := pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, o.Position.X-2, o.Position.Y-2)
				hid.MovePointer(x, y)

				i.lastRun = time.Now()
				return nil
			}
		}
	}

	i.lastRun = time.Now()
	return fmt.Errorf("object %s not found", i.objectName)
}
