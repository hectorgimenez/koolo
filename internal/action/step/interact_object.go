package step

import (
	"fmt"
	"strings"
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

	// If no completion check provided and not defined here default to waiting for interaction
	if isCompletedFn == nil {
		isCompletedFn = func() bool {
			// For stash if we have open menu we can return early
			if strings.EqualFold(string(obj.Name), "Bank") {
				return ctx.Data.OpenMenus.Stash
			}
			if obj.IsChest() {
				chest, found := ctx.Data.Objects.FindByID(obj.ID)
				// Since opening a chest is immediate and the mode changes right away,
				// we can return true as soon as we see these states
				if !found || chest.Mode == mode.ObjectModeOperating || chest.Mode == mode.ObjectModeOpened {
					return true
				}
				// Also return true if no longer selectable (as a fallback)
				return !chest.Selectable
			}

			// For portals, check if the player has entered the portal's destination area
			if obj.IsPortal() || obj.IsRedPortal() {
				if ctx.Data.PlayerUnit.Area == obj.PortalData.DestArea {
					if areaData, ok := ctx.Data.Areas[obj.PortalData.DestArea]; ok {
						if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
							return true
						}
					}
				}
			}

			return waitingForInteraction
		}
	}

	for !isCompletedFn() {
		ctx.PauseIfNotPriority()

		if interactionAttempts >= maxInteractionAttempts || mouseOverAttempts >= 20 {
			return fmt.Errorf("failed interacting with object")
		}

		ctx.RefreshGameData()

		// Give some time before retrying the interaction
		if waitingForInteraction && time.Since(lastRun) < time.Millisecond*200 {
			// for chest we can check more often status is almost instant
			if !obj.IsChest() || time.Since(lastRun) < time.Millisecond*50 {
				continue
			}
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

		// If portal is still being created, wait
		if o.IsPortal() || o.IsRedPortal() {
			// Detect JustPortaled state and wait for loading screen if it's active
			if ctx.Data.PlayerUnit.States.HasState(state.JustPortaled) {
				// Check for loading screen during portal transition
				if ctx.Data.OpenMenus.LoadingScreen {
					ctx.WaitForGameToLoad()
					break
				}
			}

			if o.Mode != mode.ObjectModeOpened {
				utils.Sleep(100)
				continue
			}
		}

		if o.IsChest() && o.Mode == mode.ObjectModeOperating {
			continue // Skip if chest is already being opened
		}

		if o.IsHovered {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			waitingForInteraction = true
			interactionAttempts++

			if (o.IsPortal() || o.IsRedPortal()) && o.PortalData.DestArea != 0 {
				startTime := time.Now()
				for time.Since(startTime) < time.Second*2 {
					// Check for loading screen during portal transition
					if ctx.Data.OpenMenus.LoadingScreen {
						ctx.WaitForGameToLoad()
						break
					}
					utils.Sleep(50)
				}

				for attempts := 0; attempts < maxPortalSyncAttempts; attempts++ {
					if ctx.Data.PlayerUnit.Area == o.PortalData.DestArea {
						if areaData, ok := ctx.Data.Areas[o.PortalData.DestArea]; ok {
							if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
								if o.PortalData.DestArea.IsTown() {
									return nil
								}
								// For special areas, ensure we have proper object data loaded
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
		} else {

			objectX := o.Position.X - 2
			objectY := o.Position.Y - 2
			distance := ctx.PathFinder.DistanceFromMe(o.Position)
			if distance > 15 {
				return fmt.Errorf("object is too far away: %d. Current distance: %d", o.Name, distance)
			}

			mX, mY := ui.GameCoordsToScreenCords(objectX, objectY)
			// In order to avoid the spiral (super slow and shitty) let's try to point the mouse to the top of the portal directly
			if mouseOverAttempts == 2 && o.IsPortal() {
				mX, mY = ui.GameCoordsToScreenCords(objectX-4, objectY-4)
			}

			x, y := utils.Spiral(mouseOverAttempts)
			currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
			ctx.HID.MovePointer(mX+x, mY+y)
			mouseOverAttempts++
		}
	}

	return nil
}
