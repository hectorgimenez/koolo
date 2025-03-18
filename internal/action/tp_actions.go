package action

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ReturnTown() error {
	ctx := context.Get()
	ctx.SetLastAction("ReturnTown")
	ctx.PauseIfNotPriority()

	if ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	err := step.OpenPortal()
	if err != nil {
		return err
	}
	portal, found := ctx.Data.Objects.FindOne(object.TownPortal)
	if !found {
		return errors.New("portal not found")
	}

	if err = ClearAreaAroundPosition(portal.Position, 8, data.MonsterAnyFilter()); err != nil {
		ctx.Logger.Warn("Error clearing area around portal", "error", err)
	}

	// Now that it is safe, interact with portal
	err = InteractObject(portal, func() bool {
		return ctx.Data.PlayerUnit.Area.IsTown()
	})
	if err != nil {
		return err
	}

	// Wait for area transition and data sync
	utils.Sleep(1000)
	ctx.RefreshGameData()

	// Wait for town area data to be fully loaded
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if ctx.Data.PlayerUnit.Area.IsTown() {
			// Verify area data exists and is loaded
			if townData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
				if townData.IsInside(ctx.Data.PlayerUnit.Position) {
					return nil
				}
			}
		}
		utils.Sleep(100)
		ctx.RefreshGameData()
	}

	return fmt.Errorf("failed to verify town area data after portal transition")
}

func UsePortalInTown() error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalInTown")

	tpArea := town.GetTownByArea(ctx.Data.PlayerUnit.Area).TPWaitingArea(*ctx.Data)
	_ = MoveToCoords(tpArea)

	err := UsePortalFrom(ctx.Data.PlayerUnit.Name)
	if err != nil {
		return err
	}

	// Wait for area sync before attempting any movement
	utils.Sleep(500)
	ctx.RefreshGameData()
	if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		return err
	}

	// Ensure we're not in town
	if ctx.Data.PlayerUnit.Area.IsTown() {
		return fmt.Errorf("failed to leave town area")
	}

	// Perform item pickup after re-entering the portal
	err = ItemPickup(40)
	if err != nil {
		ctx.Logger.Warn("Error during item pickup after portal use", "error", err)
	}

	return nil
}

func UsePortalFrom(owner string) error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalFrom")

	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	for _, obj := range ctx.Data.Objects {
		if obj.IsPortal() && obj.Owner == owner {
			return InteractObjectByID(obj.ID, func() bool {
				if !ctx.Data.PlayerUnit.Area.IsTown() {
					// Ensure area data is synced after portal transition
					utils.Sleep(500)
					ctx.RefreshGameData()

					if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
						return false
					}
					return true
				}
				return false
			})
		}
	}

	return errors.New("portal not found")
}

func ReturnToTownWithOwnedPortal() error {
	ctx := context.Get()
	ctx.SetLastAction("ReturnTown")
	ctx.PauseIfNotPriority()

	if ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	// Move slightly if we're right next to a waypoint to prevent fail to hover portal
	for _, obj := range ctx.Data.Objects {
		if obj.IsWaypoint() && ctx.PathFinder.DistanceFromMe(obj.Position) < 3 {
			// Try a few different positions until we find one that works
			for i := 0; i < 4; i++ {
				newPos := data.Position{
					X: ctx.Data.PlayerUnit.Position.X + 3 - i,
					Y: ctx.Data.PlayerUnit.Position.Y + 3 - i,
				}
				if ctx.Data.AreaData.IsWalkable(newPos) && ctx.PathFinder.DistanceFromMe(obj.Position) >= 3 {
					MoveToCoords(newPos)
					break
				}
			}
			break
		}
	}

	err := step.OpenNewPortal()
	if err != nil {
		return err
	}

	portal, found := ctx.Data.Objects.FindOne(object.TownPortal)
	if !found {
		return errors.New("portal not found")
	}

	return InteractObject(portal, nil)
}
