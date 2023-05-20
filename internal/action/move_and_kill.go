package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) MoveToAreaAndKill(area area.Area) *Factory {
	pickupBeforeMoving := false
	openedDoors := make(map[object.Name]data.Position)

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

		// Check if there is a door blocking our way
		if !step.CanTeleport(d) {
			for _, o := range d.Objects {
				if o.IsDoor() && pather.DistanceFromMe(d, o.Position) < 10 && openedDoors[o.Name] != o.Position {
					return BuildStatic(func(d data.Data) []step.Step {
						b.logger.Info("Door detected and teleport is not available, trying to open it...")
						openedDoors[o.Name] = o.Position
						return []step.Step{step.InteractObject(o.Name, nil)}
					})
				}
			}
		}

		for _, a := range d.AdjacentLevels {
			if a.Area == area {
				return BuildStatic(func(d data.Data) []step.Step {
					if d := pather.DistanceFromMe(d, a.Position); d < 10 {
						return []step.Step{step.MoveToLevel(area)}
					}

					return []step.Step{step.MoveTo(
						a.Position,
						step.ClosestWalkable(),
						step.WithTimeout(1),
					)}
				})

			}
		}

		b.logger.Debug("Destination area not found", zap.Any("area", area))
		return nil
	})
}

func (b Builder) MoveAndKill(toFunc func(d data.Data) (data.Position, bool)) *Factory {
	pickupBeforeMoving := false
	openedDoors := make(map[object.Name]data.Position)

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

		// Check if there is a door blocking our way
		if !step.CanTeleport(d) {
			for _, o := range d.Objects {
				if o.IsDoor() && pather.DistanceFromMe(d, o.Position) < 10 && openedDoors[o.Name] != o.Position {
					return BuildStatic(func(d data.Data) []step.Step {
						b.logger.Info("Door detected and teleport is not available, trying to open it...")
						openedDoors[o.Name] = o.Position
						return []step.Step{step.InteractObject(o.Name, nil)}
					})
				}
			}
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
