package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) IdentifyAll(skipIdentify bool) *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		items := b.itemsToIdentify(data)

		if len(items) == 0 || skipIdentify {
			return
		}

		b.logger.Info("Identifying items...")
		steps = append(steps,
			step.SyncStepWithCheck(func(data game.Data) error {
				hid.PressKey(config.Config.Bindings.OpenInventory)
				return nil
			}, func(data game.Data) step.Status {
				if data.OpenMenus.Inventory {
					return step.StatusCompleted
				}
				return step.StatusInProgress
			}),
			step.SyncStep(func(data game.Data) error {
				idTome, found := getIDTome(data)
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

func (b Builder) itemsToIdentify(data game.Data) (items []game.Item) {
	for _, i := range data.Items.Inventory {
		if i.Identified || i.Quality == game.ItemQualityNormal || i.Quality == game.ItemQualitySuperior {
			continue
		}

		items = append(items, i)
	}

	return
}

func identifyItem(idTome game.Item, i game.Item) {
	xIDTome := int(float32(hid.GameAreaSizeX)/town.InventoryTopLeftX) + idTome.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	yIDTome := int(float32(hid.GameAreaSizeY)/town.InventoryTopLeftY) + idTome.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)

	hid.MovePointer(xIDTome, yIDTome)
	helper.Sleep(100)
	hid.Click(hid.RightButton)
	x := int(float32(hid.GameAreaSizeX)/town.InventoryTopLeftX) + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	y := int(float32(hid.GameAreaSizeY)/town.InventoryTopLeftY) + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
	hid.MovePointer(x, y)
	helper.Sleep(100)
	hid.Click(hid.LeftButton)
	helper.Sleep(350)
}

func getIDTome(data game.Data) (game.Item, bool) {
	for _, i := range data.Items.Inventory {
		if i.Name == game.ItemTomeOfIdentify {
			return i, true
		}
	}

	return game.Item{}, false
}
