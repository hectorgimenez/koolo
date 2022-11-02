package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

func (b Builder) ReturnTown() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		if data.PlayerUnit.Area.IsTown() {
			return
		}

		lastRun := time.Time{}
		steps = append(steps,
			step.SyncStepWithCheck(func(data game.Data) error {
				// Give some time to portal to popup before retrying...
				if time.Since(lastRun) < time.Second*3 {
					return nil
				}

				hid.PressKey(config.Config.Bindings.TP)
				helper.Sleep(50)
				hid.Click(hid.RightButton)
				lastRun = time.Now()
				return nil
			}, func(data game.Data) step.Status {
				// Give some extra time, sometimes if we move the mouse over the portal before is shown
				// and there is an intractable entity behind it, will keep it focused
				if time.Since(lastRun) > time.Second*1 {
					for _, o := range data.Objects {
						if o.IsPortal() {
							return step.StatusCompleted
						}
					}
				}

				return step.StatusInProgress
			}),
			step.InteractObject(object.TownPortal, func(data game.Data) bool {
				return data.PlayerUnit.Area.IsTown()
			}),
		)
		return
	}, Resettable())
}
