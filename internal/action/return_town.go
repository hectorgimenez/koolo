package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
)

func (b Builder) ReturnTown() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if data.Area.IsTown() {
			return
		}

		steps = append(steps,
			step.SyncAction(func(data game.Data) error {
				hid.PressKey(config.Config.Bindings.TP)
				helper.Sleep(50)
				hid.Click(hid.RightButton)
				return nil
			}),
			step.InteractObject("TownPortal", func(data game.Data) bool {
				return data.Area.IsTown()
			}),
		)
		return
	})
}
