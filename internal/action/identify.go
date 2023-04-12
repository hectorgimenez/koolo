package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) IdentifyAll(skipIdentify bool) *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		items := b.itemsToIdentify(d)

		if len(items) == 0 || skipIdentify {
			return
		}

		b.logger.Info("Identifying items...")
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
				idTome, found := getIDTome(d)
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

func (b Builder) itemsToIdentify(d data.Data) (items []data.Item) {
	for _, i := range d.Items.Inventory {
		if i.Identified || i.Quality == item.QualityNormal || i.Quality == item.QualitySuperior {
			continue
		}

		items = append(items, i)
	}

	return
}

func identifyItem(idTome data.Item, i data.Item) {
	xIDTome := town.InventoryTopLeftX + idTome.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	yIDTome := town.InventoryTopLeftY + idTome.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)

	hid.MovePointer(xIDTome, yIDTome)
	helper.Sleep(200)
	hid.Click(hid.RightButton)
	helper.Sleep(200)
	x := town.InventoryTopLeftX + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	y := town.InventoryTopLeftY + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
	hid.MovePointer(x, y)
	helper.Sleep(300)
	hid.Click(hid.LeftButton)
	helper.Sleep(350)
}

func getIDTome(d data.Data) (data.Item, bool) {
	for _, i := range d.Items.Inventory {
		if i.Name == item.TomeOfIdentify {
			return i, true
		}
	}

	return data.Item{}, false
}
