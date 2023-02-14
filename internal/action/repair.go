package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) Repair() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		shouldRepair := false
		for _, i := range data.Items.Equipped {
			if d, found := i.Stats[stat.Durability]; found && d < 3 {
				shouldRepair = true
				b.logger.Info(fmt.Sprintf("Repairing %s, durability is: %d", i.Name, d))
				break
			}
		}

		if shouldRepair {
			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(data.PlayerUnit.Area).RepairNPC()),
				step.KeySequence("home", "down", "enter"),
				step.SyncStep(func(_ game.Data) error {
					helper.Sleep(100)
					hid.MovePointer(390, 515)
					hid.Click(hid.LeftButton)
					helper.Sleep(500)
					hid.PressKey("esc")
					return nil
				}),
			)
		}

		return
	}, Resettable(), CanBeSkipped())
}
