package action

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func WayPoint(dest area.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("WayPoint")

	if !ctx.Data.PlayerUnit.Area.IsTown() {
		if err := ReturnTown(); err != nil {
			return err
		}
	}

	if ctx.Data.PlayerUnit.Area == dest {
		ctx.WaitForGameToLoad()
		return nil
	}

	wpCoords, found := area.WPAddresses[dest]
	if !found {
		return fmt.Errorf("area destination %s is not mapped to a WayPoint (waypoint.go)", area.Areas[dest].Name)
	}

	for _, o := range ctx.Data.Objects {
		if o.IsWaypoint() {

			err := InteractObject(o, func() bool {
				return ctx.Data.OpenMenus.Waypoint
			})
			if err != nil {
				return err
			}
			if ctx.Data.LegacyGraphics {
				actTabX := ui.WpTabStartXClassic + (wpCoords.Tab-1)*ui.WpTabSizeXClassic + (ui.WpTabSizeXClassic / 2)
				ctx.HID.Click(game.LeftButton, actTabX, ui.WpTabStartYClassic)
			} else {
				actTabX := ui.WpTabStartX + (wpCoords.Tab-1)*ui.WpTabSizeX + (ui.WpTabSizeX / 2)
				ctx.HID.Click(game.LeftButton, actTabX, ui.WpTabStartY)
			}
			utils.Sleep(200)
			// Just to make sure no message like TZ change or public game spam prevent bot from clicking on waypoint
			ClearMessages()
		}
	}

	err := useWP(dest)
	if err != nil {
		return err
	}

	// Wait for the game to load after using the waypoint
	ctx.WaitForGameToLoad()

	// Verify that we've reached the destination
	ctx.RefreshGameData()
	if ctx.Data.PlayerUnit.Area != dest {
		return fmt.Errorf("failed to reach destination area %s using waypoint", area.Areas[dest].Name)
	}

	return nil
}
func useWP(dest area.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("useWP")

	finalDestination := dest
	traverseAreas := make([]area.ID, 0)
	currentWP := area.WPAddresses[dest]
	if !slices.Contains(ctx.Data.PlayerUnit.AvailableWaypoints, dest) {
		for {
			traverseAreas = append(currentWP.LinkedFrom, traverseAreas...)

			if currentWP.LinkedFrom != nil {
				dest = currentWP.LinkedFrom[0]
			}

			currentWP = area.WPAddresses[currentWP.LinkedFrom[0]]

			if slices.Contains(ctx.Data.PlayerUnit.AvailableWaypoints, dest) {
				break
			}
		}
	}

	currentWP = area.WPAddresses[dest]

	// First use the previous available waypoint that we have discovered
	if ctx.Data.LegacyGraphics {
		areaBtnY := ui.WpListStartYClassic + (currentWP.Row-1)*ui.WpAreaBtnHeightClassic + (ui.WpAreaBtnHeightClassic / 2)
		ctx.HID.Click(game.LeftButton, ui.WpListPositionXClassic, areaBtnY)
	} else {
		areaBtnY := ui.WpListStartY + (currentWP.Row-1)*ui.WpAreaBtnHeight + (ui.WpAreaBtnHeight / 2)
		ctx.HID.Click(game.LeftButton, ui.WpListPositionX, areaBtnY)
	}
	utils.Sleep(1000)

	// We have the WP discovered, just use it
	if len(traverseAreas) == 0 {
		return nil
	}

	traverseAreas = append(traverseAreas, finalDestination)

	// Next keep traversing all the areas from the previous available waypoint until we reach the destination, trying to discover WPs during the way
	ctx.Logger.Info("Traversing areas to reach destination", slog.Any("areas", traverseAreas))

	for i, dst := range traverseAreas {
		if i > 0 {
			err := MoveToArea(dst)
			if err != nil {
				return err
			}

			err = DiscoverWaypoint()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
