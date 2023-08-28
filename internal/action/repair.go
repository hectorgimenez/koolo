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

func (b Builder) Repair() *Factory {
	return NewFactory(func(d data.Data) Action {
		for _, i := range d.Items.ByLocation(item.LocationEquipped) {
			du, found := i.Stats[stat.Durability]
			if _, maxDurabilityFound := i.Stats[stat.MaxDurability]; maxDurabilityFound && !found || (found && du.Value <= 1) {
				b.logger.Info(fmt.Sprintf("Repairing %s, durability is: %d", i.Name, du.Value))
				return b.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).RepairNPC(),
					step.KeySequence("home", "down", "enter"),
					step.SyncStep(func(_ data.Data) error {
						helper.Sleep(100)
						hid.MovePointer(390, 515)
						hid.Click(hid.LeftButton)
						helper.Sleep(500)
						return nil
					}),
				)
			}
		}

		return nil
	})
}
