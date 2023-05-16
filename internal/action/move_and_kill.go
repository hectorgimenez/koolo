package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) MoveToAreaAndKill(area area.Area) *Factory {
	pickupBeforeMoving := false
	return NewFactory(func(d data.Data) Action {
		if d.PlayerUnit.Area == area {
			b.logger.Debug("Already in area", zap.Any("area", area))
			return nil
		}

		for _, m := range d.Monsters.Enemies() {
			if d := pather.DistanceFromMe(d, m.Position); d < 5 {
				pickupBeforeMoving = true
				return b.ch.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
					return m.UnitID, true
				}, nil)
			}
		}

		if pickupBeforeMoving {
			pickupBeforeMoving = false
			return b.ItemPickup(false, 50)
		}

		return BuildStatic(func(d data.Data) []step.Step {
			for _, a := range d.AdjacentLevels {
				if a.Area == area {
					if d := pather.DistanceFromMe(d, a.Position); d < 10 {
						return []step.Step{step.MoveToLevel(area)}
					}

					return []step.Step{step.MoveTo(
						a.Position,
						step.ClosestWalkable(),
						step.WithTimeout(1),
					)}
				}
			}

			return nil
		})
	})
}

func (b Builder) MoveAndKill(toFunc func(d data.Data) (data.Position, bool)) *Factory {
	pickupBeforeMoving := false

	return NewFactory(func(d data.Data) Action {
		to, found := toFunc(d)
		if !found {
			return nil
		}

		if pather.DistanceFromMe(d, to) < 5 {
			return nil
		}

		for _, m := range d.Monsters.Enemies() {
			if d := pather.DistanceFromMe(d, m.Position); d < 5 {
				pickupBeforeMoving = true
				return b.ch.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
					return m.UnitID, true
				}, nil)
			}
		}

		if pickupBeforeMoving {
			pickupBeforeMoving = false
			return b.ItemPickup(false, 50)
		}

		return BuildStatic(func(d data.Data) []step.Step {
			return []step.Step{step.MoveTo(
				to,
				step.ClosestWalkable(),
				step.WithTimeout(1),
			)}
		})
	})
}
