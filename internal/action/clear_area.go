package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func ClearAreaAroundPlayer(radius int, filter data.MonsterFilter) error {
	return ClearAreaAroundPosition(context.Get().Data.PlayerUnit.Position, radius, filter)
}

func ClearAreaAroundPosition(pos data.Position, radius int, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.SetLastAction("ClearAreaAroundPosition")

	return ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies(filter) {
			distanceToTarget := pather.DistanceFromPoint(pos, m.Position)
			if ctx.Data.AreaData.IsWalkable(m.Position) && distanceToTarget <= radius {
				return m.UnitID, true
			}
		}

		return 0, false
	}, nil)
}

// TODO handle repath when stuck with hidden stash/ shrines. this is root of hidden stash stuck issue in chaos
func ClearThroughPath(pos data.Position, radius int, filter data.MonsterFilter) error {
	ctx := context.Get()
	lastMovement := false
	currentArea := ctx.Data.PlayerUnit.Area
	canTeleport := ctx.Data.CanTeleport()

	for {
		ctx.PauseIfNotPriority()

		// Clear enemies at current position
		ClearAreaAroundPosition(ctx.Data.PlayerUnit.Position, radius, filter)

		if lastMovement {
			return nil
		}

		// Get current position for path calculation
		currentPos := ctx.Data.PlayerUnit.Position

		// Get path to destination, leveraging both context cache and global cache
		var path pather.Path
		var found bool

		// Try context cache first for UI consistency
		if ctx.CurrentGame.PathCache != nil &&
			ctx.CurrentGame.PathCache.DestPosition == pos &&
			ctx.CurrentGame.PathCache.IsPathValid(currentPos) {
			// Use existing path from context cache
			path = ctx.CurrentGame.PathCache.Path
			found = true
		} else {
			// Try to get path from global cache
			path, found = pather.GetCachedPath(currentPos, pos, currentArea, canTeleport)
			if !found {
				// Calculate new path if not in cache
				path, _, found = ctx.PathFinder.GetPath(pos)
				if !found {
					return fmt.Errorf("path could not be calculated")
				}

				// Store in global cache
				pather.StorePath(currentPos, pos, currentArea, canTeleport, path, currentPos)
			}

			// Update context cache for UI reference
			ctx.CurrentGame.PathCache = &context.PathCache{
				Path:             path,
				DestPosition:     pos,
				StartPosition:    currentPos,
				DistanceToFinish: step.DistanceToFinishMoving,
			}
		}

		// Calculate movement distance for this segment
		movementDistance := radius
		if movementDistance > len(path) {
			movementDistance = len(path)
		}

		// Set destination to next path segment
		dest := data.Position{
			X: path[movementDistance-1].X + ctx.Data.AreaOrigin.X,
			Y: path[movementDistance-1].Y + ctx.Data.AreaOrigin.Y,
		}

		// Check if this is the last movement segment
		if len(path)-movementDistance <= step.DistanceToFinishMoving {
			lastMovement = true
		}

		// Move to next segment
		if err := step.MoveTo(dest); err != nil {
			return err
		}
	}
}
