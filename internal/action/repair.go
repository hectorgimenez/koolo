package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) Repair() *DynamicAction {
	repaired := false
	return BuildDynamic(func(d data.Data) (steps []step.Step, valid bool) {
		if repaired {
			if d.OpenMenus.NPCShop {
				b.logger.Info("ESC")
				steps = append(steps,
					step.SyncStep(func(_ data.Data) error {
						hid.PressKey("esc")
						return nil
					}),
				)

				return steps, true

			}

			return nil, false
		}

		shouldRepair := false
		for _, i := range d.Items.Equipped {
			if du, found := i.Stats[stat.Durability]; found && du.Value <= 1 {
				shouldRepair = true
				b.logger.Info(fmt.Sprintf("Repairing %s, durability is: %d", i.Name, du.Value))
				break
			}
		}

		if shouldRepair {
			repaired = true
			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).RepairNPC()),
				step.KeySequence("home", "down", "enter"),
				step.SyncStep(func(_ data.Data) error {
					helper.Sleep(100)
					hid.MovePointer(390, 515)
					hid.Click(hid.LeftButton)
					helper.Sleep(500)
					return nil
				}),
			)
			return steps, true
		}

		return
	}, CanBeSkipped())
}
