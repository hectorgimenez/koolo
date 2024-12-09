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
	ctx := context.Get()
	ctx.SetLastStep("InteractObject")

	interactionAttempts := 0
	mouseOverAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastHoverCoords := data.Position{} // Track last successful hover position for portals
	lastRun := time.Now()

	// If no completion check provided, default to waiting for interaction
	if isCompletedFn == nil {
		isCompletedFn = func() bool {

			if strings.EqualFold(string(obj.Name), "Bank") {
				return ctx.Data.OpenMenus.Stash
			}
			// For chests, check mode and selectability
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
			// For chests, we can check more frequently since the state changes fast
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

		// Check portal states
		if o.IsPortal() || o.IsRedPortal() {
			if o.Mode == mode.ObjectModeOperating || o.Mode != mode.ObjectModeOpened {
				utils.Sleep(100)
				continue
			}
		}

		if o.IsChest() {
			if o.Mode == mode.ObjectModeOperating {
				continue // Skip if chest is already being opened
			}
		}

		// Store successful hover position for portals
		if o.IsHovered && currentMouseCoords != (data.Position{}) && (o.IsPortal() || o.IsRedPortal()) {
			lastHoverCoords = currentMouseCoords
		}

		if o.IsHovered {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			waitingForInteraction = true
			interactionAttempts++

			// For portals, verify area transition
			if (o.IsPortal() || o.IsRedPortal()) && o.PortalData.DestArea != 0 {
				startTime := time.Now()
				for time.Since(startTime) < time.Second*2 {
					// Check for loading screen during portal transition
					if ctx.Data.OpenMenus.LoadingScreen {
						ctx.Logger.Debug("Loading screen detected during portal transition...")
						ctx.WaitForGameToLoad()
						break
					}

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

		// Try last successful hover position first for portals
		if mouseOverAttempts == 0 && lastHoverCoords != (data.Position{}) && (o.IsPortal() || o.IsRedPortal()) {
			currentMouseCoords = lastHoverCoords
			ctx.HID.MovePointer(lastHoverCoords.X, lastHoverCoords.Y)
			mouseOverAttempts++
			utils.Sleep(100)
			continue
		}

		mX, mY := ui.GameCoordsToScreenCords(objectX, objectY)
		if mouseOverAttempts == 2 && (o.IsPortal() || o.IsRedPortal()) {
			mX, mY = ui.GameCoordsToScreenCords(objectX-4, objectY-4)
		}

		x, y := utils.Spiral(mouseOverAttempts)
		x = x / 3
		y = y / 3
		currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
		ctx.HID.MovePointer(mX+x, mY+y)
		mouseOverAttempts++
		utils.Sleep(100)
	}

	// After successful portal transition
	if (obj.IsPortal() || obj.IsRedPortal()) && ctx.Data.PlayerUnit.Area == obj.PortalData.DestArea {
		if areaData, ok := ctx.Data.Areas[obj.PortalData.DestArea]; ok {
			if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
				// Enable area correction with this area as expected
				ctx.CurrentGame.AreaCorrection.Enabled = true
				ctx.CurrentGame.AreaCorrection.ExpectedArea = obj.PortalData.DestArea
				return nil
			}
		}
	}

	return nil
}
