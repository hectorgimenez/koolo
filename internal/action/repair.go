package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) Repair() *Chain {
	return NewChain(func(d data.Data) []Action {
		for _, i := range d.Items.ByLocation(item.LocationEquipped) {
			if du, found := i.Stats[stat.Durability]; found && du.Value <= 1 {
				b.logger.Info(fmt.Sprintf("Repairing %s, durability is: %d", i.Name, du.Value))
				return []Action{
					b.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).RepairNPC(),
						step.KeySequence("home", "down", "enter"),
						step.SyncStep(func(_ data.Data) error {
							helper.Sleep(100)
							hid.MovePointer(390, 515)
							hid.Click(hid.LeftButton)
							helper.Sleep(500)
							return nil
						}),
					),
				}
			}
		}

		return nil
	})
}
