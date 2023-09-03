package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func (b *Builder) IdentifyAll(skipIdentify bool) *StepChainAction {
	return NewStepChain(func(d data.Data) (steps []step.Step) {
		items := b.itemsToIdentify(d)

		b.logger.Debug("Checking for items to identify...")
		if len(items) == 0 || skipIdentify {
			b.logger.Debug("No items to identify...")
			return
		}

		b.logger.Info(fmt.Sprintf("Identifying %d items...", len(items)))
		steps = append(steps,
			step.SyncStepWithCheck(func(d data.Data) error {
				hid.PressKey(config.Config.Bindings.OpenInventory)
				return nil
			}, func(d data.Data) step.Status {
				if d.OpenMenus.Inventory {
					return step.StatusCompleted
				}
				return step.StatusInProgress
			}),
			step.SyncStep(func(d data.Data) error {
				idTome, found := d.Items.Find(item.TomeOfIdentify, item.LocationInventory)
				if !found {
					b.logger.Warn("ID Tome not found, not identifying items")
					return nil
				}

				for _, i := range items {
					identifyItem(idTome, i)
				}

				hid.PressKey("esc")

				return nil
			}),
		)

		return
	}, Resettable(), CanBeSkipped())
}

func (b *Builder) itemsToIdentify(d data.Data) (items []data.Item) {
	for _, i := range d.Items.ByLocation(item.LocationInventory) {
		if i.Identified || i.Quality == item.QualityNormal || i.Quality == item.QualitySuperior {
			continue
		}

		items = append(items, i)
	}

	return
}

func identifyItem(idTome data.Item, i data.Item) {
	screenPos := ui.GetScreenCoordsForItem(idTome)
	hid.MovePointer(screenPos.X, screenPos.Y)
	helper.Sleep(350)
	hid.Click(hid.RightButton)
	helper.Sleep(200)

	screenPos = ui.GetScreenCoordsForItem(i)
	hid.MovePointer(screenPos.X, screenPos.Y)
	helper.Sleep(300)
	hid.Click(hid.LeftButton)
	helper.Sleep(350)
}
