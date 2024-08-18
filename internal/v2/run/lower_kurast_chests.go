package run

import (
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/v2/action"
	"github.com/hectorgimenez/koolo/internal/v2/context"
)

var minChestDistanceFromBonfire = 25
var maxChestDistanceFromBonfire = 45

type LowerKurastChests struct {
	ctx *context.Status
}

func NewLowerKurastChest() *LowerKurastChests {
	return &LowerKurastChests{
		ctx: context.Get(),
	}
}

func (run LowerKurastChests) Name() string {
	return string(config.LowerKurastRun)
}

func (run LowerKurastChests) Run() error {
	run.ctx.Logger.Debug("Running a Lower Kurast Chest run")
	var bonFirePositions []data.Position
	// Use Waypoint to Lower Kurast
	err := action.WayPoint(area.LowerKurast)
	if err != nil {
		return err
	}

	// Find the bonfires
	for _, o := range run.ctx.Data.Objects {
		if o.Name == object.SmallFire {
			bonFirePositions = append(bonFirePositions, o.Position)
		}
	}

	run.ctx.Logger.Debug("Found bonfires", "bonfires", bonFirePositions)

	var chestsIds = []object.Name{object.JungleMediumChestLeft, object.JungleChest}

	// Move to each of the bonfires one by one
	for _, bonfirePos := range bonFirePositions {
		// Move to the bonfire
		err = action.MoveToCoords(bonfirePos)
		if err != nil {
			return err
		}
		// Find the chests
		var chests []data.Object
		for _, o := range run.ctx.Data.Objects {
			if slices.Contains(chestsIds, o.Name) && isChestWithinBonfireRange(o, bonfirePos) {
				chests = append(chests, o)
			}
		}

		// Open the chests
		for _, chest := range chests {
			err = action.InteractObject(chest, func() bool {
				chest, _ := run.ctx.Data.Objects.FindByID(chest.ID)
				return !chest.Selectable
			})
			if err != nil {
				run.ctx.Logger.Warn("Failed interacting with chest: %v", err)
			}
		}
	}

	// Return to town
	if err = action.ReturnTown(); err != nil {
		return err
	}

	// Move to A4 if possible to shorten the run time
	err = action.WayPoint(area.ThePandemoniumFortress)
	if err != nil {
		return err
	}

	// Done
	return nil
}

func isChestWithinBonfireRange(chest data.Object, bonfirePosition data.Position) bool {
	distance := pather.DistanceFromPoint(chest.Position, bonfirePosition)
	return distance >= minChestDistanceFromBonfire && distance <= maxChestDistanceFromBonfire
}
