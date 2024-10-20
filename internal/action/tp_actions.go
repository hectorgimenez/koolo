package action

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ReturnTown() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ReturnTown"
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

	clearNearbyMonsters(portal.Position, 8)
	// Now that it is safe, interact with portal

	return InteractObject(portal, func() bool {
		return ctx.Data.PlayerUnit.Area.IsTown()
	})
}

func UsePortalInTown() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "UsePortalInTown"

	tpArea := town.GetTownByArea(ctx.Data.PlayerUnit.Area).TPWaitingArea(*ctx.Data)
	_ = MoveToCoords(tpArea)

	err := UsePortalFrom(ctx.Data.PlayerUnit.Name)
	if err != nil {
		return err
	}

	// Wait for the game to fully load after using the portal
	ctx.WaitForGameToLoad()

	// Refresh game data to ensure we have the latest information
	ctx.RefreshGameData()

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
	ctx.ContextDebug.LastAction = "UsePortalFrom"

	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	for _, obj := range ctx.Data.Objects {
		if obj.IsPortal() && obj.Owner == owner {
			return InteractObjectByID(obj.ID, func() bool {
				if !ctx.Data.PlayerUnit.Area.IsTown() {
					utils.Sleep(500)
					return true
				}
				return false
			})
		}
	}

	return errors.New("portal not found")
}
