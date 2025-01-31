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
	portalSyncDelay        = 200
	maxPortalSyncAttempts  = 15
)

func InteractObject(obj data.Object, isCompletedFn func() bool) error {
    // Enhanced hidden stash check
    if string(obj.Name) == "hidden stash" || obj.ID == 125 || obj.ID == 127 || obj.ID == 128 {  
        return fmt.Errorf("interaction blocked for hidden stash")
    }

    // Add position validation before interaction
    ctx := context.Get()
    if !ctx.Data.AreaData.IsWalkable(obj.Position) {
        if walkablePos, found := ctx.PathFinder.FindNearbyWalkablePosition(obj.Position); found {
            if err := MoveTo(walkablePos); err != nil {
                return fmt.Errorf("failed to move to safe interaction position: %v", err)
            }
        }
    }

    interactionAttempts := 0
    mouseOverAttempts := 0
    waitingForInteraction := false
    currentMouseCoords := data.Position{}
    lastRun := time.Time{}

    ctx.SetLastStep("InteractObject")

    // Rest of the original function remains unchanged
    if isCompletedFn == nil {
        isCompletedFn = func() bool {
            return waitingForInteraction
        }
    }

    expectedArea := area.ID(0)
    if obj.IsRedPortal() {
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
                    ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) &&
                    len(ctx.Data.Objects) > 0
            }
        }
    }

    for !isCompletedFn() {
        ctx.PauseIfNotPriority()

        if interactionAttempts >= maxInteractionAttempts || mouseOverAttempts >= 20 {
            return fmt.Errorf("[%s] failed interacting with object [%v] in Area: [%s]", ctx.Name, obj.Name, ctx.Data.PlayerUnit.Area.Area().Name)
        }

        ctx.RefreshGameData()

        if waitingForInteraction && time.Since(lastRun) < time.Millisecond*200 {
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
                utils.Sleep(500)
                for attempts := 0; attempts < maxPortalSyncAttempts; attempts++ {
                    ctx.RefreshGameData()
                    if ctx.Data.PlayerUnit.Area == expectedArea {
                        if areaData, ok := ctx.Data.Areas[expectedArea]; ok {
                            if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
                                if expectedArea.IsTown() {
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
                return fmt.Errorf("portal sync timeout - expected area: %v, current: %v", expectedArea, ctx.Data.PlayerUnit.Area)
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
