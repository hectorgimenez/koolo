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
	basicStep
	objectName            object.Name
	waitingForInteraction bool
	isCompleted           func(data.Data) bool
	mouseOverAttempts     int
}

func InteractObject(name object.Name, isCompleted func(data.Data) bool) *InteractObjectStep {
	return &InteractObjectStep{
		basicStep:   newBasicStep(),
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
	i.tryTransitionStatus(StatusInProgress)

	if i.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("object %d could not be interacted", i.objectName)
	}

	// Give some time before retrying the interaction
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second*1 {
		return nil
	}

	i.lastRun = time.Now()
	if o, found := d.Objects.FindOne(i.objectName); found {
		if o.IsHovered {
			hid.Click(hid.LeftButton)
			i.waitingForInteraction = true
			return nil
		} else {
			distance := pather.DistanceFromMe(d, o.Position)
			if distance > 15 {
				return fmt.Errorf("object is too far away: %d. Current distance: %d", o.Name, distance)
			}

			objectX := o.Position.X - 2
			objectY := o.Position.Y - 2
			mX, mY := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX, objectY)

			// In order to avoid the spiral (super slow and shitty) let's try to point the mouse to the top of the portal directly
			if i.mouseOverAttempts == 2 && i.objectName == object.TownPortal {
				mX, mY = pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX-4, objectY-4)
			}

			x, y := helper.Spiral(i.mouseOverAttempts)
			hid.MovePointer(mX+x, mY+y)
			i.mouseOverAttempts++

			return nil
		}
	}

	return fmt.Errorf("object %d not found", i.objectName)
}
