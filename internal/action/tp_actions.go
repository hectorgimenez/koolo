package action

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	portalTransitionTimeout = 3 * time.Second
	portalSyncDelay         = 100
	initialTransitionDelay  = 200
)

func ReturnTown() error {
	ctx := context.Get()
	ctx.SetLastAction("ReturnTown")
	ctx.PauseIfNotPriority()

	if ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	// Store initial state
	fromArea := ctx.Data.PlayerUnit.Area
	townArea := town.GetTownByArea(fromArea).TownArea()

	ctx.Logger.Debug("Starting town return sequence",
		slog.String("from_area", fromArea.Area().Name),
		slog.String("to_area", townArea.Area().Name))

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

	// Disable area correction before portal interaction
	ctx.CurrentGame.AreaCorrection.Enabled = false
	ctx.SwitchPriority(context.PriorityHigh)
	defer func() {
		ctx.CurrentGame.AreaCorrection.Enabled = true
		ctx.SwitchPriority(context.PriorityNormal)
	}()

	ctx.Logger.Debug("Interacting with portal")
	// Now that it is safe, interact with portal
	err = InteractObject(portal, func() bool {
		return ctx.Data.PlayerUnit.Area.IsTown()
	})
	if err != nil {
		return err
	}

	// Initial delay to let the game process the transition request
	utils.Sleep(initialTransitionDelay)
	ctx.RefreshGameData()

	// Verify we've actually started the transition
	if ctx.Data.PlayerUnit.Area == fromArea {
		ctx.Logger.Debug("Still in source area after initial delay, starting transition wait")
	}

	// Wait for proper town transition
	deadline := time.Now().Add(portalTransitionTimeout)
	transitionStartTime := time.Now()

	for time.Now().Before(deadline) {
		ctx.RefreshGameData()
		currentArea := ctx.Data.PlayerUnit.Area

		ctx.Logger.Debug("Waiting for transition",
			slog.String("current_area", currentArea.Area().Name),
			slog.Duration("elapsed", time.Since(transitionStartTime)))

		if currentArea == townArea {
			if areaData, ok := ctx.Data.Areas[townArea]; ok {
				if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
					ctx.Logger.Debug("Successfully reached town area")
					utils.Sleep(300) // Extra wait to ensure everything is loaded
					ctx.RefreshGameData()
					return nil
				}
			}
		}
		utils.Sleep(portalSyncDelay)
	}

	return fmt.Errorf("failed to verify town transition within timeout (start area: %s)", fromArea.Area().Name)
}
func UsePortalInTown() error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalInTown")

	tpArea := town.GetTownByArea(ctx.Data.PlayerUnit.Area).TPWaitingArea(*ctx.Data)
	if err := MoveToCoords(tpArea); err != nil {
		return err
	}

	// Store initial state
	fromArea := ctx.Data.PlayerUnit.Area

	// Disable area correction and raise priority during transition
	ctx.CurrentGame.AreaCorrection.Enabled = false
	ctx.SwitchPriority(context.PriorityHigh)
	defer func() {
		ctx.CurrentGame.AreaCorrection.Enabled = true
		ctx.SwitchPriority(context.PriorityNormal)
	}()

	err := UsePortalFrom(ctx.Data.PlayerUnit.Name)
	if err != nil {
		return err
	}

	// Initial delay to let the game process the transition request
	utils.Sleep(initialTransitionDelay)
	ctx.RefreshGameData()

	// Wait for proper area transition
	deadline := time.Now().Add(portalTransitionTimeout)
	transitionStartTime := time.Now()

	for time.Now().Before(deadline) {
		ctx.RefreshGameData()
		currentArea := ctx.Data.PlayerUnit.Area

		ctx.Logger.Debug("Waiting for transition",
			slog.String("current_area", currentArea.Area().Name),
			slog.Duration("elapsed", time.Since(transitionStartTime)))

		if currentArea != fromArea {
			if areaData, ok := ctx.Data.Areas[currentArea]; ok {
				if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
					ctx.Logger.Debug("Successfully reached destination area")
					utils.Sleep(300) // Extra wait to ensure everything is loaded
					ctx.RefreshGameData()

					// Perform item pickup after re-entering the portal
					err = ItemPickup(40)
					if err != nil {
						ctx.Logger.Warn("Error during item pickup after portal use", "error", err)
					}
					return nil
				}
			}
		}
		utils.Sleep(portalSyncDelay)
	}

	return fmt.Errorf("failed to verify area transition within timeout (start area: %s)", fromArea.Area().Name)
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
					// Initial delay to let the game process the transition
					utils.Sleep(initialTransitionDelay)
					ctx.RefreshGameData()

					// Verify we're no longer in town
					if ctx.Data.PlayerUnit.Area.IsTown() {
						return false
					}

					if areaData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
						return areaData.IsInside(ctx.Data.PlayerUnit.Position)
					}
				}
				return false
			})
		}
	}

	return errors.New("portal not found")
}
