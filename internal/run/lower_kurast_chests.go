package run

import (
	"fmt"
	"slices"
	"sort"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
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
	return string(config.LowerKurastChestRun)
}

func (run LowerKurastChests) Run() error {
	run.ctx.Logger.Debug("Running a Lower Kurast Chest run")

	// Use Waypoint to Lower Kurast
	err := action.WayPoint(area.LowerKurast)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	// Get bonfires from cached map data
	var bonFirePositions []data.Position
	if areaData, ok := run.ctx.GameReader.GetData().Areas[area.LowerKurast]; ok {
		for _, obj := range areaData.Objects {
			if obj.Name == object.Name(160) { // SmallFire
				run.ctx.Logger.Debug("Found bonfire at:", "position", obj.Position)
				bonFirePositions = append(bonFirePositions, obj.Position)
			}
		}
	}

	run.ctx.Logger.Debug("Total bonfires found", "count", len(bonFirePositions))

	// Define objects to interact with : chests + weapon racks/armor stands (if enabled)
	interactableObjects := []object.Name{object.JungleMediumChestLeft, object.JungleChest}

	if run.ctx.CharacterCfg.Game.LowerKurastChest.OpenRacks {
		interactableObjects = append(interactableObjects,
			object.ArmorStandRight,
			object.ArmorStandLeft,
			object.WeaponRackRight,
			object.WeaponRackLeft,
		)
	}

	// Move to each of the bonfires one by one
	for _, bonfirePos := range bonFirePositions {
		// Move to the bonfire
		err = action.MoveToCoords(bonfirePos)
		if err != nil {
			return err
		}

		// Find the interactable objects
		var objects []data.Object
		for _, o := range run.ctx.Data.Objects {
			if slices.Contains(interactableObjects, o.Name) && isChestWithinBonfireRange(o, bonfirePos) {
				objects = append(objects, o)
			}
		}

		// Interact with objects in the order of shortest travel
		for len(objects) > 0 {

			playerPos := run.ctx.Data.PlayerUnit.Position

			sort.Slice(objects, func(i, j int) bool {
				return pather.DistanceFromPoint(objects[i].Position, playerPos) <
					pather.DistanceFromPoint(objects[j].Position, playerPos)
			})

			// Interact with the closest object
			closestObject := objects[0]
			err = action.InteractObject(closestObject, func() bool {
				object, _ := run.ctx.Data.Objects.FindByID(closestObject.ID)
				return !object.Selectable
			})
			if err != nil {
				run.ctx.Logger.Warn(fmt.Sprintf("[%s] failed interacting with object [%v] in Area: [%s]", run.ctx.Name, closestObject.Name, run.ctx.Data.PlayerUnit.Area.Area().Name), err)
			}
			utils.Sleep(500) // Add small delay to allow the game to open the object and drop the content

			// Remove the interacted container from the list
			objects = objects[1:]
		}
	}

	// Done
	return nil
}

func isChestWithinBonfireRange(chest data.Object, bonfirePosition data.Position) bool {
	distance := pather.DistanceFromPoint(chest.Position, bonfirePosition)
	return distance >= minChestDistanceFromBonfire && distance <= maxChestDistanceFromBonfire
}
