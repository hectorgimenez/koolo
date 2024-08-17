package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/utils"
)

func InteractObject(obj data.Object, isCompletedFn func() bool) error {
	maxInteractionAttempts := 5
	interactionAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	// If there is no check, just assume the interaction is completed after clicking
	if isCompletedFn == nil {
		isCompletedFn = func() bool {
			return waitingForInteraction
		}
	}

	ctx := context.Get()
	ctx.ContextDebug.LastStep = "InteractObject"

	for !isCompletedFn() {
		ctx.RefreshGameData()

		// Pause the execution if the priority is not the same as the execution priority
		if ctx.ExecutionPriority != ctx.Priority {
			continue
		}

		if interactionAttempts >= maxInteractionAttempts {
			return errors.New("failed interacting with object")
		}

		// Give some time before retrying the interaction
		if waitingForInteraction && time.Since(lastRun) < time.Millisecond*500 {
			continue
		}

		var o data.Object
		var found bool
		if obj.ID != 0 {
			o, found = ctx.Data.Objects.FindByID(obj.ID)
			if !found {
				return fmt.Errorf("object %v not found", obj)
			}
		} else {
			o, found = ctx.Data.Objects.FindOne(obj.Name)
			if !found {
				return fmt.Errorf("object %v not found", obj)
			}
		}

		lastRun = time.Now()
		if o.IsHovered {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			waitingForInteraction = true
			interactionAttempts++
			continue
		} else {
			objectX := o.Position.X - 2
			objectY := o.Position.Y - 2
			distance := ctx.PathFinder.DistanceFromMe(o.Position)
			if distance > 15 {
				return fmt.Errorf("object is too far away: %d. Current distance: %d", o.Name, distance)
			}

			mX, mY := utils.GameCoordsToScreenCords(objectX, objectY)
			// In order to avoid the spiral (super slow and shitty) let's try to point the mouse to the top of the portal directly
			if interactionAttempts == 2 && o.Name == object.TownPortal {
				mX, mY = utils.GameCoordsToScreenCords(objectX-4, objectY-4)
			}

			x, y := utils.Spiral(interactionAttempts)
			currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
			ctx.HID.MovePointer(mX+x, mY+y)
			interactionAttempts++
		}
	}

	return nil
}
