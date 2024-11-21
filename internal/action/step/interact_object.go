package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxInteractionAttempts = 5
	portalSyncDelay        = 150
	maxPortalSyncAttempts  = 25
)

func findPortalPosition(obj data.Object, ctx *context.Status) (data.Position, error) {
	if ctx.Data.AreaData.IsWalkable(obj.Position) {
		return obj.Position, nil
	}
	for radius := 1; radius <= 5; radius++ {
		for x := -radius; x <= radius; x++ {
			for y := -radius; y <= radius; y++ {
				if x == 0 && y == 0 {
					continue
				}
				pos := data.Position{X: obj.Position.X + x, Y: obj.Position.Y + y}
				if ctx.Data.AreaData.IsWalkable(pos) {
					if path, _, found := ctx.PathFinder.GetPath(pos); found && len(path) > 0 {
						return pos, nil
					}
				}
			}
		}
	}
	return data.Position{}, fmt.Errorf("no walkable position found near portal at %v", obj.Position)
}

func handlePortalSync(ctx *context.Status, expectedArea area.ID) error {
	utils.Sleep(500)
	for attempts := 0; attempts < maxPortalSyncAttempts; attempts++ {
		ctx.RefreshGameData()
		if ctx.Data.PlayerUnit.Area == expectedArea {
			if areaData, ok := ctx.Data.Areas[expectedArea]; ok {
				if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
					if expectedArea.IsTown() {
						return nil
					}
					if (len(ctx.Data.Objects) > 0 || expectedArea == area.TheWorldstoneChamber ||
						expectedArea == area.DurielsLair) && ctx.Data.PlayerUnit.Mode != mode.Dead {
						return nil
					}
				}
			}
		}
		if ctx.Data.OpenMenus.LoadingScreen {
			utils.Sleep(portalSyncDelay * 2)
			continue
		}
		utils.Sleep(portalSyncDelay)
	}
	return fmt.Errorf("portal sync timeout - expected: %v, current: %v, mode: %v",
		expectedArea, ctx.Data.PlayerUnit.Area, ctx.Data.PlayerUnit.Mode)
}

func InteractObject(obj data.Object, isCompletedFn func() bool) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractObject")

	// For portals, we need to ensure we can reach them
	if obj.IsPortal() || obj.IsRedPortal() {
		accessPos, err := findPortalPosition(obj, ctx)
		if err == nil {
			if err := MoveTo(accessPos); err != nil {
				ctx.Logger.Debug("Failed to move to portal access position, continuing with default behavior")
			}
		}
	}

	interactionAttempts := 0
	mouseOverAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	// If there is no completion check, just assume the interaction is completed after clicking
	if isCompletedFn == nil {
		isCompletedFn = func() bool { return waitingForInteraction }
	}

	// For portals, we need to ensure proper area sync
	expectedArea := area.ID(0)
	if obj.IsRedPortal() {
		// For red portals, we need to determine the expected destination
		switch {
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.StonyField:
			expectedArea = area.Tristram
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.RogueEncampment:
			expectedArea = area.MooMooFarm
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.Harrogath:
			expectedArea = area.NihlathaksTemple
		case obj.Name == object.PermanentTownPortal && ctx.Data.PlayerUnit.Area == area.ArcaneSanctuary:
			expectedArea = area.CanyonOfTheMagi
		case obj.Name == object.BaalsPortal && ctx.Data.PlayerUnit.Area == area.ThroneOfDestruction:
			expectedArea = area.TheWorldstoneChamber
		case obj.Name == object.DurielsLairPortal && (ctx.Data.PlayerUnit.Area >= area.TalRashasTomb1 && ctx.Data.PlayerUnit.Area <= area.TalRashasTomb7):
			expectedArea = area.DurielsLair
		}
	} else if obj.IsPortal() {
		fromArea := ctx.Data.PlayerUnit.Area
		if !fromArea.IsTown() {
			expectedArea = town.GetTownByArea(fromArea).TownArea()
		} else {
			isCompletedFn = func() bool {
				return !ctx.Data.PlayerUnit.Area.IsTown() &&
					ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) && len(ctx.Data.Objects) > 0
			}
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

		if o.IsPortal() || o.IsRedPortal() {
			if o.Mode == mode.ObjectModeOperating {
				utils.Sleep(100)
				continue
			}
			if o.Mode != mode.ObjectModeOpened {
				utils.Sleep(100)
				continue
			}
		}

		if o.IsHovered {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			waitingForInteraction = true
			interactionAttempts++

			if expectedArea != 0 {
				return handlePortalSync(ctx, expectedArea)
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
