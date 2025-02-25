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
func ClearThroughPath(pos data.Position, radius int, filter data.MonsterFilter) error {
	ctx := context.Get()
	lastMovement := false

	for {
		ctx.PauseIfNotPriority()

		ClearAreaAroundPosition(ctx.Data.PlayerUnit.Position, radius, filter)

		if lastMovement {
			return nil
		}

		var path []data.Position
		var found bool
		if ctx.CurrentGame.PathCache != nil &&
			ctx.CurrentGame.PathCache.DestPosition == pos &&
			step.IsPathValid(ctx.Data.PlayerUnit.Position, ctx.CurrentGame.PathCache) {
			path = ctx.CurrentGame.PathCache.Path
			found = true
		} else {

			path, _, found = ctx.PathFinder.GetPath(pos)
			if found {
				ctx.CurrentGame.PathCache = &context.PathCache{
					Path:             path,
					DestPosition:     pos,
					StartPosition:    ctx.Data.PlayerUnit.Position,
					DistanceToFinish: step.DistanceToFinishMoving,
				}
			}
		}

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

		if len(path)-movementDistance <= step.DistanceToFinishMoving {
			lastMovement = true
		}

		if err := step.MoveTo(dest); err != nil {
			return err
		}
	}
}
