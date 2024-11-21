package action

import (
	"errors"
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
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

	// Remember original area for area correction
	originalArea := ctx.Data.PlayerUnit.Area

	// Open portal and wait for it to appear
	err := step.OpenPortal()
	if err != nil {
		return err
	}

	portal, found := ctx.Data.Objects.FindOne(object.TownPortal)
	if !found {
		return errors.New("portal not found")
	}

	// Clear the area around portal before using it
	if err = ClearAreaAroundPosition(portal.Position, 8, data.MonsterAnyFilter()); err != nil {
		ctx.Logger.Warn("Error clearing area around portal", "error", err)
	}

	// Get the expected town area based on current location
	expectedTownArea := town.GetTownByArea(originalArea).TownArea()

	// Disable area correction during portal transition
	ctx.CurrentGame.AreaCorrection.Enabled = false

	// Interact with portal and verify transition
	err = step.InteractObject(portal, func() bool {
		if !ctx.Data.PlayerUnit.Area.IsTown() {
			return false
		}

		// Verify area data exists and is loaded
		if townData, ok := ctx.Data.Areas[expectedTownArea]; ok {
			return townData.IsInside(ctx.Data.PlayerUnit.Position)
		}
		return false
	})

	if err != nil {
		ctx.CurrentGame.AreaCorrection.Enabled = true
		return err
	}

	// Set expected area to town
	ctx.CurrentGame.AreaCorrection.ExpectedArea = expectedTownArea
	ctx.CurrentGame.AreaCorrection.Enabled = true

	return nil
}

func UsePortalInTown() error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalInTown")

	// Store current area for verification
	townArea := ctx.Data.PlayerUnit.Area
	if !townArea.IsTown() {
		return errors.New("must be in town to use town portal")
	}

	// Move to town's portal waiting area
	tpArea := town.GetTownByArea(townArea).TPWaitingArea(*ctx.Data)
	if err := MoveToCoords(tpArea); err != nil {
		return fmt.Errorf("failed to move to portal area: %w", err)
	}

	// Temporarily disable area correction during portal use
	ctx.CurrentGame.AreaCorrection.Enabled = false

	err := UsePortalFrom(ctx.Data.PlayerUnit.Name)
	if err != nil {
		ctx.CurrentGame.AreaCorrection.Enabled = true
		return err
	}

	// Wait for area transition and verify
	if err = ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		ctx.CurrentGame.AreaCorrection.Enabled = true
		return err
	}

	// Verify we're not in town
	if ctx.Data.PlayerUnit.Area.IsTown() {
		ctx.CurrentGame.AreaCorrection.Enabled = true
		return errors.New("failed to leave town area")
	}

	// Set area correction to new area
	ctx.CurrentGame.AreaCorrection.ExpectedArea = ctx.Data.PlayerUnit.Area
	ctx.CurrentGame.AreaCorrection.Enabled = true

	return ItemPickup(40)
}

func UsePortalFrom(owner string) error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalFrom")

	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	// Find the correct portal and store destination area if available
	var targetPortal data.Object
	var found bool
	for _, obj := range ctx.Data.Objects {
		if obj.IsPortal() && obj.Owner == owner {
			targetPortal = obj
			found = true
			break
		}
	}

	if !found {
		return errors.New("portal not found")
	}

	// Wait for portal to be fully opened
	if targetPortal.Mode != mode.ObjectModeOpened {
		return errors.New("portal is not ready")
	}

	// Use portal and verify transition
	return step.InteractObject(targetPortal, func() bool {
		// Successful when we're no longer in town and area data is synced
		if ctx.Data.PlayerUnit.Area.IsTown() {
			return false
		}

		if areaData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
			return areaData.IsInside(ctx.Data.PlayerUnit.Position) && len(ctx.Data.Objects) > 0
		}
		return false
	})
}
