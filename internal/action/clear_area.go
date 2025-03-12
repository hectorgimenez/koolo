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

	for {
		ctx.PauseIfNotPriority()

		// Check for closest monster within radius - monsters are already sorted by distance
		err := ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies(filter) {
				dist := pather.DistanceFromPoint(pos, m.Position)
				if ctx.Data.AreaData.IsWalkable(m.Position) && dist <= radius {
					return m.UnitID, true
				}
			}
			return 0, false
		}, nil)

		if err != nil {
			return err
		}

		// If no monsters found within radius, we're done
		found := false
		for _, m := range ctx.Data.Monsters.Enemies(filter) {
			if pather.DistanceFromPoint(pos, m.Position) <= radius {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
}
func ClearThroughPath(pos data.Position, radius int, filter data.MonsterFilter) error {
	ctx := context.Get()

	lastMovement := false
	for {
		ctx.PauseIfNotPriority()

		ClearAreaAroundPosition(ctx.Data.PlayerUnit.Position, radius, filter)

		if lastMovement {
			return nil
		}

		path, _, found := ctx.PathFinder.GetPath(pos)
		if !found {
			return fmt.Errorf("path could not be calculated")
		}

		movementDistance := radius
		if radius > len(path) {
			movementDistance = len(path)
		}

		dest := data.Position{
			X: path[movementDistance-1].X + ctx.Data.AreaData.OffsetX,
			Y: path[movementDistance-1].Y + ctx.Data.AreaData.OffsetY,
		}

		// Let's handle the last movement logic to MoveTo function, we will trust the pathfinder because
		// it can finish within a bigger distance than we expect (because blockers), so we will just check how far
		// we should be after the latest movement in a theoretical way
		if len(path)-movementDistance <= step.DistanceToFinishMoving {
			lastMovement = true
		}
		// Increasing DistanceToFinishMoving prevent not being to able to finish movement if our destination is center of a large object like Seal in diablo run.
		// is used only for pathing, attack.go will use default DistanceToFinishMoving
		err := step.MoveTo(dest, step.WithDistanceToFinish(7))
		if err != nil {
			return err
		}
	}
}
