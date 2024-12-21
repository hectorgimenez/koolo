package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
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

	return InteractObject(portal, nil)
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

	// Perform item pickup after re-entering the portal
	if err = ItemPickup(40); err != nil {
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
			return InteractObjectByID(obj.ID, nil)
		}
	}

	return errors.New("portal not found")
}
