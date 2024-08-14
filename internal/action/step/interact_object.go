package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type InteractObjectStep struct {
	basicStep
	objectName            object.Name
	objectID              data.UnitID
	waitingForInteraction bool
	isCompleted           func(game.Data) bool
	mouseOverAttempts     int
	currentMouseCoords    data.Position
}

func InteractObject(name object.Name, isCompleted func(game.Data) bool) *InteractObjectStep {
	return &InteractObjectStep{
		basicStep:   newBasicStep(),
		objectName:  name,
		isCompleted: isCompleted,
	}
}

func InteractObjectByID(ID data.UnitID, isCompleted func(game.Data) bool) *InteractObjectStep {
	return &InteractObjectStep{
		basicStep:   newBasicStep(),
		objectID:    ID,
		isCompleted: isCompleted,
	}
}

func (i *InteractObjectStep) Status(d game.Data, _ container.Container) Status {
	if i.status == StatusCompleted {
		return StatusCompleted
	}

	// If isCompleted is nil, we run it at least once, sometimes there is no good way to check interaction
	if time.Since(i.lastRun) > time.Second*1 && i.isCompleted == nil && i.waitingForInteraction {
		i.tryTransitionStatus(StatusCompleted)
	}

	if i.isCompleted != nil && i.isCompleted(d) {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractObjectStep) Run(d game.Data, container container.Container) error {
	i.tryTransitionStatus(StatusInProgress)

	if i.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("object %d could not be interacted", i.objectName)
	}

	// Give some time before retrying the interaction
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Millisecond*500 {
		return nil
	}

	i.lastRun = time.Now()
	var o data.Object
	var found bool

	if i.objectID != 0 {
		for _, obj := range d.Objects {
			if obj.ID == i.objectID {
				o = obj
				found = true
				break
			}
		}
	} else {
		o, found = d.Objects.FindOne(i.objectName)
		// Let's try to use our own portal instead of any random portal we find
		if i.objectName == object.TownPortal {
			for _, obj := range d.Objects {
				if obj.Owner == d.PlayerUnit.Name {
					o = obj
					found = true
				}
			}
		}
	}

	if found {
		objectX := o.Position.X - 1
		objectY := o.Position.Y - 1
		if o.IsHovered {
			container.HID.Click(game.LeftButton, i.currentMouseCoords.X, i.currentMouseCoords.Y)
			i.waitingForInteraction = true
			return nil
		} else {
			distance := pather.DistanceFromMe(d, o.Position)
			if distance > 15 {
				return fmt.Errorf("object is too far away: %d. Current distance: %d", o.Name, distance)
			}

			mX, mY := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX, objectY)
			// In order to avoid the spiral (super slow and shitty) let's try to point the mouse to the top of the portal directly
			if i.mouseOverAttempts == 2 && o.Name == object.TownPortal {
				mX, mY = container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX-4, objectY-4)
			}

			x, y := helper.Spiral(i.mouseOverAttempts)
			i.currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
			container.HID.MovePointer(mX+x, mY+y)
			i.mouseOverAttempts++

			return nil
		}
	}

	return fmt.Errorf("object %d not found", i.objectName)
}
