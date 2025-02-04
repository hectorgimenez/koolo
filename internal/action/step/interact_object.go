package step

import (
	"fmt"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxInteractionAttempts = 10
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
	lastRun := time.Now()

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
				// If in loading screen or just portaled state, consider interaction incomplete
				if ctx.Data.OpenMenus.LoadingScreen || ctx.Data.PlayerUnit.States.HasState(state.JustPortaled) {
					return false
				}

				// Check if we're in the destination area
				if ctx.Data.PlayerUnit.Area == obj.PortalData.DestArea {
					if areaData, ok := ctx.Data.Areas[obj.PortalData.DestArea]; ok {
						return areaData.IsInside(ctx.Data.PlayerUnit.Position)
					}
					// If area data not available but we're in correct area, consider it complete
					return true
				}
			}

			return waitingForInteraction
		}
	}

	for !isCompletedFn() {
		ctx.PauseIfNotPriority()

		if interactionAttempts >= maxInteractionAttempts || mouseOverAttempts >= 20 {
			if obj.IsPortal() {
				// For portals, log warning and continue instead of erroring . It will retry
				ctx.Logger.Warn(fmt.Sprintf("Portal interaction attempts exceeded for %s [ID: %d], continuing...", obj.Name, obj.ID))
				return nil
			}
			return fmt.Errorf("failed interacting with object: %s [ID: %d] after %d attempts", obj.Name, obj.ID, interactionAttempts)
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
			if obj.IsPortal() || obj.IsRedPortal() {
				// For portals, missing object may mean we're transitioning - continue
				utils.Sleep(100)
				continue
			}
			return fmt.Errorf("object %v not found", obj)
		}

		lastRun = time.Now()

		// Handle portal states
		if o.IsPortal() || o.IsRedPortal() {
			// Detect JustPortaled state and wait for loading screen if it's active
			if ctx.Data.PlayerUnit.States.HasState(state.JustPortaled) {
				// Check for loading screen during portal transition
				if ctx.Data.OpenMenus.LoadingScreen {
					ctx.WaitForGameToLoad()
					continue
				}
			}

			if o.Mode != mode.ObjectModeOpened {
				utils.Sleep(100)
				continue
			}
		}

		// Handle chest states
		if o.IsChest() && o.Mode == mode.ObjectModeOperating {
			continue // Skip if chest is already being opened
		}

		// Handle object interaction
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
								return nil
							}
						} else if o.PortalData.DestArea.IsTown() {
							return nil
						}
					}
					utils.Sleep(portalSyncDelay)
				}
				ctx.Logger.Warn(fmt.Sprintf("Portal sync timeout - expected area: %v, current: %v", o.PortalData.DestArea, ctx.Data.PlayerUnit.Area))
				continue
			}
			continue
		}

		// Get object description for spiral
		desc := object.Desc[int(o.Name)]
		mX, mY := ui.GameCoordsToScreenCords(o.Position.X, o.Position.Y)
		x, y := utils.ObjectSpiral(mouseOverAttempts, desc)

		currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
		ctx.HID.MovePointer(currentMouseCoords.X, currentMouseCoords.Y)
		mouseOverAttempts++
		utils.Sleep(100)
	}

	return nil
}
