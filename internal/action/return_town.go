package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

func (b Builder) ReturnTown() *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		if d.PlayerUnit.Area.IsTown() {
			return
		}

		return []step.Step{
			step.OpenPortal(),
			step.InteractObject(object.TownPortal, func(d data.Data) bool {
				return d.PlayerUnit.Area.IsTown()
			}),
		}
	}, Resettable())
}
