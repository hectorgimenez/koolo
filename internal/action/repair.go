package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b *Builder) Repair() *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		for _, i := range d.Items.ByLocation(item.LocationEquipped) {
			du, found := i.Stats[stat.Durability]
			if _, maxDurabilityFound := i.Stats[stat.MaxDurability]; maxDurabilityFound && !found || (found && du.Value <= 1) {
				b.logger.Info(fmt.Sprintf("Repairing %s, durability is: %d", i.Name, du.Value))
				repairNPC := town.GetTownByArea(d.PlayerUnit.Area).RepairNPC()
				if repairNPC == npc.Hratli {
					actions = append(actions, b.MoveToCoords(data.Position{X: 5224, Y: 5045}))
				}
				keys := make([]string, 0)
				keys = append(keys, "home")
				if repairNPC != npc.Halbu {
					keys = append(keys, "down")
				}
				keys = append(keys, "enter")

				return append(actions, b.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).RepairNPC(),
					step.KeySequence(keys...),
					step.SyncStep(func(_ data.Data) error {
						helper.Sleep(100)
						hid.MovePointer(390, 515)
						hid.Click(hid.LeftButton)
						helper.Sleep(500)
						return nil
					}),
					step.KeySequence("esc"),
				))
			}
		}

		return nil
	})
}
