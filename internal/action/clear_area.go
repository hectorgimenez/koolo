package action

import (
	"fmt"
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
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
	ctx.Logger.Debug("Clearing area around position...", slog.Int("radius", radius))

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

	for {
		ctx.PauseIfNotPriority()

		ClearAreaAroundPosition(ctx.Data.PlayerUnit.Position, radius, filter)

		path, distance, found := ctx.PathFinder.GetPath(pos)
		if !found {
			return fmt.Errorf("path could not be calculated")
		}

		if distance <= distanceToFinishMoving {
			return nil
		}

		movementDistance := radius
		if radius > len(path) {
			movementDistance = len(path)
		}

		dest := data.Position{
			X: path[movementDistance-1].X + ctx.Data.AreaData.OffsetX,
			Y: path[movementDistance-1].Y + ctx.Data.AreaData.OffsetY,
		}
		err := MoveToCoords(dest)
		if err != nil {
			return err
		}
	}
}
