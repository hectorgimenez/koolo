package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/object"
)

func (b Builder) ReturnTown() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		if data.PlayerUnit.Area.IsTown() {
			return
		}

		return []step.Step{
			step.OpenPortal(),
			step.InteractObject(object.TownPortal, func(data game.Data) bool {
				return data.PlayerUnit.Area.IsTown()
			}),
		}
	}, Resettable())
}
