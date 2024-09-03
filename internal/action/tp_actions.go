package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ReturnTown() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ReturnTown"

	if ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	_ = step.OpenPortal()
	tp, found := ctx.Data.Objects.FindOne(object.TownPortal)
	if !found {
		return errors.New("town portal not found")
	}

	return InteractObject(tp, func() bool {
		return ctx.Data.PlayerUnit.Area.IsTown()
	})
}

func UsePortalInTown() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "UsePortalInTown"

	tpArea := town.GetTownByArea(ctx.Data.PlayerUnit.Area).TPWaitingArea(*ctx.Data)
	_ = MoveToCoords(tpArea)

	return UsePortalFrom(ctx.Data.PlayerUnit.Name)
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

	return nil
}
