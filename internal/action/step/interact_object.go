package step

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxInteractionAttempts = 5
	maxMouseOverAttempts   = 20
	interactionDelay       = 200
	maxPortalSyncAttempts  = 25
	portalSyncDelay        = 150
	maxObjectDistance      = 15
	hoverVerifyDelay       = 50
)

func InteractObject(obj data.Object, isCompletedFn func() bool) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractObject")

	interactionAttempts := 0
	mouseOverAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	// Default completion check if none provided
	if isCompletedFn == nil {
		isCompletedFn = func() bool { return waitingForInteraction }
	}

	// Determine expected area and potentially modify completion function for portals
	expectedArea := determineExpectedArea(obj, ctx, &isCompletedFn)

	for !isCompletedFn() {
		ctx.PauseIfNotPriority()

		// Check for max attempts
		if interactionAttempts >= maxInteractionAttempts || mouseOverAttempts >= maxMouseOverAttempts {
			return fmt.Errorf("failed to interact with object after %d attempts", maxInteractionAttempts)
		}

		// Throttle interaction attempts
		if waitingForInteraction && time.Since(lastRun) < interactionDelay {
			continue
		}

		// Find and verify object
		o, found := findObject(obj, ctx)
		if !found {
			return fmt.Errorf("object %v not found", obj)
		}

		lastRun = time.Now()

		// Handle portal states
		if shouldWaitForPortal(o) {
			utils.Sleep(100)
			continue
		}

		// Verify hover state with a small delay
		if o.IsHovered {
			// Double check hover state after a small delay
			utils.Sleep(hoverVerifyDelay)
			o, found = findObject(obj, ctx)
			if !found || !o.IsHovered {
				continue
			}

			// Click multiple times for portals to ensure interaction
			if o.IsPortal() || o.IsRedPortal() {
				for i := 0; i < 2; i++ {
					ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
					utils.Sleep(50)
				}
			} else {
				ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			}

			waitingForInteraction = true
			interactionAttempts++

			if expectedArea != 0 {
				utils.Sleep(750)
				return handlePortalTransition(ctx, expectedArea)
			}
			continue
		}

		// Check distance and try to interact
		if err := tryInteraction(o, ctx, &mouseOverAttempts, &currentMouseCoords); err != nil {
			return err
		}

		// Small delay after moving mouse before next iteration
		utils.Sleep(hoverVerifyDelay)
	}

	return nil
}

func findObject(obj data.Object, ctx *context.Status) (data.Object, bool) {
	if obj.ID != 0 {
		return ctx.Data.Objects.FindByID(obj.ID)
	}
	return ctx.Data.Objects.FindOne(obj.Name)
}

func shouldWaitForPortal(o data.Object) bool {
	if !o.IsPortal() && !o.IsRedPortal() {
		return false
	}
	return o.Mode == mode.ObjectModeOperating || o.Mode != mode.ObjectModeOpened
}

func tryInteraction(o data.Object, ctx *context.Status, attempts *int, coords *data.Position) error {
	distance := ctx.PathFinder.DistanceFromMe(o.Position)
	if distance > maxObjectDistance {
		return fmt.Errorf("object too far: %d (distance: %d)", o.Name, distance)
	}

	objectX := o.Position.X - 2
	objectY := o.Position.Y - 2

	mX, mY := ui.GameCoordsToScreenCords(objectX, objectY)
	if *attempts == 2 && o.Name == object.TownPortal {
		mX, mY = ui.GameCoordsToScreenCords(objectX-4, objectY-4)
	}

	x, y := utils.Spiral(*attempts)
	*coords = data.Position{X: mX + x, Y: mY + y}
	ctx.HID.MovePointer(mX+x, mY+y)
	*attempts++

	return nil
}

func determineExpectedArea(obj data.Object, ctx *context.Status, isCompletedFn *func() bool) area.ID {
	if obj.IsRedPortal() {
		// For red portals, we need to determine the expected destination
		switch {
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.StonyField:
			return area.Tristram
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.RogueEncampment:
			return area.MooMooFarm
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.Harrogath:
			return area.NihlathaksTemple
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.ArcaneSanctuary:
			return area.CanyonOfTheMagi
		case obj.Name == object.BaalsPortal && ctx.Data.PlayerUnit.Area == area.ThroneOfDestruction:
			return area.TheWorldstoneChamber
		case obj.Name == object.DurielsLairPortal && (ctx.Data.PlayerUnit.Area >= area.TalRashasTomb1 && ctx.Data.PlayerUnit.Area <= area.TalRashasTomb7):
			return area.DurielsLair
		}
	} else if obj.IsPortal() {
		fromArea := ctx.Data.PlayerUnit.Area
		if !fromArea.IsTown() {
			return town.GetTownByArea(fromArea).TownArea()
		} else {
			// For town portals going out of town, modify the completion function
			*isCompletedFn = func() bool {
				return !ctx.Data.PlayerUnit.Area.IsTown() &&
					ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) && len(ctx.Data.Objects) > 0
			}
		}
	}

	return 0
}

func handlePortalTransition(ctx *context.Status, expectedArea area.ID) error {
	attempts := 0
	maxAttempts := maxPortalSyncAttempts

	for attempts < maxAttempts {
		// Check if we're in a loading screen
		if ctx.Data.OpenMenus.LoadingScreen {
			utils.Sleep(portalSyncDelay * 2)
			continue
		}

		// Check if we've reached the expected area
		if ctx.Data.PlayerUnit.Area == expectedArea {
			if areaData, ok := ctx.Data.Areas[expectedArea]; ok {
				if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
					// Additional checks for special areas
					if expectedArea.IsTown() ||
						len(ctx.Data.Objects) > 0 ||
						expectedArea == area.TheWorldstoneChamber ||
						expectedArea == area.DurielsLair {
						return nil
					}
				}
			}
		}

		attempts++
		utils.Sleep(portalSyncDelay)
	}

	return fmt.Errorf("portal transition failed - expected: %v, current: %v (after %d attempts)",
		expectedArea, ctx.Data.PlayerUnit.Area, maxAttempts)
}
