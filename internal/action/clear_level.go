package action

import (
	"errors"
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ClearCurrentLevel(openChests bool, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ClearCurrentLevel"

	rooms := ctx.PathFinder.OptimizeRoomsTraverseOrder()
	for _, r := range rooms {
		err := clearRoom(r, filter)
		if err != nil {
			ctx.Logger.Warn("Failed to clear room: %v", err)
		}

		if !openChests {
			continue
		}

		// Handle chests after room is cleared
		for _, o := range ctx.Data.Objects {
			if o.IsChest() && o.Selectable && r.IsInside(o.Position) {
				err = MoveToCoords(o.Position)
				if err != nil {
					ctx.Logger.Warn("Failed moving to chest: %v", err)
					continue
				}
				err = InteractObject(o, func() bool {
					chest, _ := ctx.Data.Objects.FindByID(o.ID)
					return !chest.Selectable
				})
				if err != nil {
					ctx.Logger.Warn("Failed interacting with chest: %v", err)
				}
				utils.Sleep(500) // Add small delay to allow the game to open the chest and drop the content
			}
		}
	}

	return nil
}

func clearRoom(room data.Room, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "clearRoom"

	// Initial room positioning
	path, _, found := ctx.PathFinder.GetClosestWalkablePath(room.GetCenter())
	if !found {
		return errors.New("failed to find a path to the room center")
	}

	to := data.Position{
		X: path.To().X + ctx.Data.AreaOrigin.X,
		Y: path.To().Y + ctx.Data.AreaOrigin.Y,
	}

	// Move to room center first
	if err := MoveToCoords(to); err != nil {
		return fmt.Errorf("failed moving to room center: %w", err)
	}

	// Allow time for monsters to become visible/targetable

	for {
		monsters := getMonstersInRoom(room, filter)
		if len(monsters) == 0 {
			return nil
		}

		// Prioritize monster raisers
		targetMonster := selectTargetMonster(monsters)

		// Let attack.go handle all positioning and attack logic
		err := ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			// Only return monsters that are still alive
			m, found := d.Monsters.FindByID(targetMonster.UnitID)
			if found && m.Stats[stat.Life] > 0 {
				return targetMonster.UnitID, true
			}
			return 0, false
		}, nil)

		if err != nil {
			ctx.Logger.Debug("Failed to kill monster, moving to next target", "error", err)
		}
	}
}

// selectTargetMonster prioritizes monsters based on their type and threat level
func selectTargetMonster(monsters []data.Monster) data.Monster {
	// Start with first monster as default target
	targetMonster := monsters[0]

	// Look for priority targets
	for _, m := range monsters {
		// Prioritize monster raisers first
		if m.IsMonsterRaiser() {
			return m
		}

		// Could add more prioritization logic here if needed
		// For example:
		// - Prioritize elites
		// - Prioritize ranged monsters
		// - Prioritize based on immunities
	}

	return targetMonster
}

func getMonstersInRoom(room data.Room, filter data.MonsterFilter) []data.Monster {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "getMonstersInRoom"

	monstersInRoom := make([]data.Monster, 0)
	for _, m := range ctx.Data.Monsters.Enemies(filter) {
		// Consider monsters either in the room or close to the player
		if m.Stats[stat.Life] > 0 && (room.IsInside(m.Position) || ctx.PathFinder.DistanceFromMe(m.Position) < 30) {
			monstersInRoom = append(monstersInRoom, m)
		}
	}

	return monstersInRoom
}
