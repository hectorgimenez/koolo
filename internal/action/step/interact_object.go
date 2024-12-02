package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxInteractionAttempts = 5
	portalSyncDelay        = 200
	maxPortalSyncAttempts  = 15
)

func InteractObject(obj data.Object, isCompletedFn func() bool) error {
	interactionAttempts := 0
	mouseOverAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	ctx := context.Get()
	ctx.SetLastStep("InteractObject")

	// If no completion check provided, default to waiting for interaction
	if isCompletedFn == nil {
		isCompletedFn = func() bool {
			return waitingForInteraction
		}
	}

	// For portals, check JustPortaled state
	if obj.IsPortal() || obj.IsRedPortal() {
		isCompletedFn = func() bool {
			return ctx.Data.PlayerUnit.States.HasState(state.JustPortaled)
		}
	}

	for !isCompletedFn() {
		ctx.PauseIfNotPriority()

		if interactionAttempts >= maxInteractionAttempts || mouseOverAttempts >= 20 {
			return fmt.Errorf("failed interacting with object")
		}

		ctx.RefreshGameData()

		if waitingForInteraction && time.Since(lastRun) < time.Millisecond*200 {
			continue
		}

		var o data.Object
		var found bool
		if obj.ID != 0 {
			o, found = ctx.Data.Objects.FindByID(obj.ID)
		} else {
			o, found = ctx.Data.Objects.FindOne(obj.Name)
		}
		if !found {
			return fmt.Errorf("object %v not found", obj)
		}

		lastRun = time.Now()

		// Check portal states
		if o.IsPortal() || o.IsRedPortal() {
			if o.Mode == mode.ObjectModeOperating || o.Mode != mode.ObjectModeOpened {
				utils.Sleep(100)
				continue
			}
		}

		if o.IsHovered {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			waitingForInteraction = true
			interactionAttempts++

			// For portals, verify area transition
			if (o.IsPortal() || o.IsRedPortal()) && o.PortalData.DestArea != 0 {
				startTime := time.Now()
				for time.Since(startTime) < time.Second*2 {
					if ctx.Data.PlayerUnit.States.HasState(state.JustPortaled) {
						break
					}
					utils.Sleep(50)
				}

				utils.Sleep(500)
				for attempts := 0; attempts < maxPortalSyncAttempts; attempts++ {
					if ctx.Data.PlayerUnit.Area == o.PortalData.DestArea {
						if areaData, ok := ctx.Data.Areas[o.PortalData.DestArea]; ok {
							if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
								if o.PortalData.DestArea.IsTown() {
									return nil
								}
								if len(ctx.Data.Objects) > 0 {
									return nil
								}
							}
						}
					}
					utils.Sleep(portalSyncDelay)
				}
				return fmt.Errorf("portal sync timeout - expected area: %v, current: %v", o.PortalData.DestArea, ctx.Data.PlayerUnit.Area)
			}
			continue
		}

		objectX := o.Position.X - 2
		objectY := o.Position.Y - 2
		distance := ctx.PathFinder.DistanceFromMe(o.Position)
		if distance > 15 {
			return fmt.Errorf("object is too far away: %d. Current distance: %d", o.Name, distance)
		}

		mX, mY := ui.GameCoordsToScreenCords(objectX, objectY)
		if mouseOverAttempts == 2 && o.IsPortal() {
			mX, mY = ui.GameCoordsToScreenCords(objectX-4, objectY-4)
		}

		x, y := utils.Spiral(mouseOverAttempts)
		currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
		ctx.HID.MovePointer(mX+x, mY+y)
		mouseOverAttempts++
	}

	return nil
}
