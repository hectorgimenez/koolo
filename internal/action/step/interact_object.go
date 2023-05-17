package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type InteractObjectStep struct {
	pathingStep
	objectName            object.Name
	waitingForInteraction bool
	isCompleted           func(data.Data) bool
	mouseOverAttempts     int
}

func InteractObject(name object.Name, isCompleted func(data.Data) bool) *InteractObjectStep {
	return &InteractObjectStep{
		pathingStep: newPathingStep(),
		objectName:  name,
		isCompleted: isCompleted,
	}
}

func (i *InteractObjectStep) Status(d data.Data) Status {
	if i.status == StatusCompleted {
		return StatusCompleted
	}

	// If isCompleted is nil, we run it at least once, sometimes there is no good way to check interaction
	if time.Since(i.lastRun) > time.Second*1 && i.isCompleted == nil && i.waitingForInteraction {
		i.tryTransitionStatus(StatusCompleted)
	}

	// Give some extra time to render the UI
	if time.Since(i.lastRun) > time.Second*1 && i.isCompleted != nil && i.isCompleted(d) {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractObjectStep) Run(d data.Data) error {
	if i.isCompleted != nil && i.isCompleted(d) {
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
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second*1 {
		return nil
	}

	for _, o := range d.Objects {
		if o.Name == i.objectName {
			if o.IsHovered {
				hid.Click(hid.LeftButton)
				i.waitingForInteraction = true
				i.lastRun = time.Now()
				return nil
			} else {
				distance := pather.DistanceFromMe(d, o.Position)

				if distance > 15 {
					path, _, found := pather.GetPath(d, o.Position)
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
				mX, mY := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX, objectY)

				x, y := helper.Spiral(i.mouseOverAttempts)

				// In order to avoid the spiral (super slow and shitty) let's try to point the mouse to the top of the portal directly
				if i.mouseOverAttempts == 2 && i.objectName == object.TownPortal {
					mX, mY = pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX-4, objectY-4)
				}

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
