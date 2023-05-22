package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
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

func (b Builder) UsePortalInTown() *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		if !d.PlayerUnit.Area.IsTown() {
			return
		}

		tpArea := town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)
		return []step.Step{
			step.MoveTo(tpArea),
			step.InteractObject(object.TownPortal, func(d data.Data) bool {
				if !d.PlayerUnit.Area.IsTown() {
					helper.Sleep(500)
					return true
				}

				return false
			}),
		}

	}, Resettable())
}
