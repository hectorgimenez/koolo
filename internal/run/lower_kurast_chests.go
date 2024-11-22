package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/pather"
	"slices"
	"sort"
)

const (
	minChestDistanceFromBonfire = 25
	maxChestDistanceFromBonfire = 45
)

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
	run.ctx.Logger.Debug("Starting Lower Kurast Chest run")

	// Use Waypoint to Lower Kurast
	err := action.WayPoint(area.LowerKurast)
	if err != nil {
		return err
	}

	// Get bonfires from cached map data
	cachedLKData := run.ctx.GameReader.GetData()
	var bonfirePositions []data.Position
	for _, obj := range cachedLKData.Objects {
		if obj.Name == object.Name(160) { // SmallFire
			run.ctx.Logger.Debug("Found bonfire at:", "position", obj.Position)
			bonfirePositions = append(bonfirePositions, obj.Position)
		}
	}

	run.ctx.Logger.Debug("Total bonfires found ", "count", len(bonfirePositions))

	// Define chest types we want to look for
	var chestsIds = []object.Name{
		object.JungleMediumChestLeft,
		object.JungleChest,
		object.GoodChest,
		object.NotSoGoodChestName,
	}

	// Weapon Racks and Armor Stands if enabled
	var rackIds []object.Name
	if run.ctx.CharacterCfg.Game.LowerKurastChest.OpenRacks {
		rackIds = []object.Name{
			object.ArmorStandRight,
			object.ArmorStandLeft,
			object.WeaponRackRight,
			object.WeaponRackLeft,
		}
	}

	processedBonfires := make(map[data.Position]bool)

	// Process each bonfire one by one
	for {
		// Sort remaining bonfires by distance from player
		unprocessedBonfires := make([]data.Position, 0)
		for _, pos := range bonfirePositions {
			if !processedBonfires[pos] {
				unprocessedBonfires = append(unprocessedBonfires, pos)
			}
		}

		if len(unprocessedBonfires) == 0 {
			break
		}

		// Sort by distance from current position
		sort.Slice(unprocessedBonfires, func(i, j int) bool {
			return run.ctx.PathFinder.DistanceFromMe(unprocessedBonfires[i]) <
				run.ctx.PathFinder.DistanceFromMe(unprocessedBonfires[j])
		})

		bonfirePos := unprocessedBonfires[0]

		// Move to bonfire
		if walkablePos, found := run.ctx.PathFinder.FindNearbyWalkablePosition(bonfirePos); found {
			err = action.MoveToCoords(walkablePos)
			if err != nil {
				run.ctx.Logger.Warn("Failed to move to bonfire, skipping", "error", err)
				processedBonfires[bonfirePos] = true
				continue
			}

			// Refresh game data after movement
			run.ctx.RefreshGameData()

			// Find all valid interactable objects near this bonfire
			var interactables []data.Object
			for _, o := range run.ctx.Data.Objects {
				if (slices.Contains(chestsIds, o.Name) || slices.Contains(rackIds, o.Name)) &&
					isObjectWithinBonfireRange(o, bonfirePos) {
					interactables = append(interactables, o)
				}
			}

			// Process objects in order of shortest path
			for len(interactables) > 0 {
				playerPos := run.ctx.Data.PlayerUnit.Position

				// Sort remaining objects by distance
				sort.Slice(interactables, func(i, j int) bool {
					return pather.DistanceFromPoint(interactables[i].Position, playerPos) <
						pather.DistanceFromPoint(interactables[j].Position, playerPos)
				})

				// Open closest object
				closestObject := interactables[0]
				err = action.InteractObject(closestObject, func() bool {
					obj, _ := run.ctx.Data.Objects.FindByID(closestObject.ID)
					return !obj.Selectable
				})
				if err != nil {
					run.ctx.Logger.Warn("Failed interacting with object", "error", err)
				}

				interactables = interactables[1:]
			}

			// Mark this bonfire as processed
			processedBonfires[bonfirePos] = true

		} else {
			run.ctx.Logger.Warn("Could not find walkable position near bonfire, skipping", "bonfire", bonfirePos)
			processedBonfires[bonfirePos] = true
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

	return nil
}

func isObjectWithinBonfireRange(obj data.Object, bonfirePosition data.Position) bool {
	distance := pather.DistanceFromPoint(obj.Position, bonfirePosition)
	return distance >= minChestDistanceFromBonfire && distance <= maxChestDistanceFromBonfire
}
