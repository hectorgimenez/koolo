package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) MoveToArea(area area.Area) Action {
	toFun := func(d data.Data) (data.Position, bool) {
		if d.PlayerUnit.Area == area {
			b.logger.Debug("Already in area", zap.Any("area", area))
			return data.Position{}, false
		}

		for _, a := range d.AdjacentLevels {
			if a.Area == area {
				return a.Position, true
			}
		}

		b.logger.Debug("Destination area not found", zap.Any("area", area))

		return data.Position{}, false
	}

	return NewChain(func(d data.Data) []Action {
		return []Action{
			b.MoveTo(toFun),
			BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.MoveToLevel(area),
				}
			}),
		}
	})
}

func (b Builder) MoveToCoords(to data.Position) *Factory {
	return b.MoveTo(func(d data.Data) (data.Position, bool) {
		return to, true
	})
}

func (b Builder) MoveTo(toFunc func(d data.Data) (data.Position, bool)) *Factory {
	pickupBeforeMoving := false
	openedDoors := make(map[object.Name]data.Position)

	return NewFactory(func(d data.Data) Action {
		to, found := toFunc(d)
		if !found {
			return nil
		}

		// To stop the movement, not very accurate
		if pather.DistanceFromMe(d, to) < 5 {
			return nil
		}

		// If we can teleport, just return the normal MoveTo step and stop here
		if step.CanTeleport(d) {
			return BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{step.MoveTo(to)}
			})
		}

		// Detect if there are monsters close to the player
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

		// Check if there is a door blocking our path
		for _, o := range d.Objects {
			if o.IsDoor() && pather.DistanceFromMe(d, o.Position) < 7 && openedDoors[o.Name] != o.Position {
				return BuildStatic(func(d data.Data) []step.Step {
					b.logger.Info("Door detected and teleport is not available, trying to open it...")
					openedDoors[o.Name] = o.Position
					return []step.Step{step.InteractObject(o.Name, nil)}
				}, CanBeSkipped())
			}
		}

		// Continue moving
		return BuildStatic(func(d data.Data) []step.Step {
			return []step.Step{step.MoveTo(
				to,
				step.ClosestWalkable(),
				step.WithTimeout(1),
			)}
		})
	})
}
