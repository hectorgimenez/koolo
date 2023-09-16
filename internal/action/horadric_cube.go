package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/ui"
	"go.uber.org/zap"
)

func (b *Builder) CubeAddItems(items ...data.Item) *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		cube, found := d.Items.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if !found {
			b.logger.Info("No Horadric Cube found in inventory")
			return nil
		}

		// Ensure stash is open
		if !d.OpenMenus.Stash {
			actions = append(actions, b.InteractObject(object.Bank, func(d data.Data) bool {
				return d.OpenMenus.Stash
			}))
		}

		b.logger.Info("Adding items to the Horadric Cube", zap.Any("items", items))

		// If items are on the Stash, pickup them to the inventory (only personal stash is supported for now)
		for _, itm := range items {
			nwIt := itm
			if nwIt.Location != item.LocationStash && nwIt.Location != item.LocationSharedStash1 && nwIt.Location != item.LocationSharedStash2 && nwIt.Location != item.LocationSharedStash3 {
				continue
			}

			b.logger.Debug("Item found on the stash, picking it up", zap.String("Item", string(nwIt.Name)))
			actions = append(actions, NewStepChain(func(d data.Data) []step.Step {
				screenPos := ui.GetScreenCoordsForItem(nwIt)
				hid.MovePointer(screenPos.X, screenPos.Y)

				hid.KeyDown("control")
				helper.Sleep(300)
				hid.Click(hid.LeftButton)
				helper.Sleep(200)
				hid.KeyUp("control")
				helper.Sleep(300)

				return nil
			}))
		}

		actions = append(actions, b.ensureCubeIsOpen(cube))

		for _, itm := range items {
			nwIt := itm
			actions = append(actions, NewStepChain(func(d data.Data) []step.Step {
				for _, updatedItem := range d.Items.AllItems {
					if nwIt.UnitID == updatedItem.UnitID {
						b.logger.Debug("Moving Item to the Horadric Cube", zap.String("Item", string(nwIt.Name)))
						screenPos := ui.GetScreenCoordsForItem(updatedItem)
						hid.MovePointer(screenPos.X, screenPos.Y)

						hid.KeyDown("control")
						helper.Sleep(300)
						hid.Click(hid.LeftButton)
						helper.Sleep(200)
						hid.KeyUp("control")
						helper.Sleep(300)
					}
				}

				return nil
			}))
		}

		return
	})
}

func (b *Builder) CubeTransmute() *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		cube, found := d.Items.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if !found {
			b.logger.Info("No Horadric Cube found in inventory")
			return nil
		}

		actions = append(actions, b.ensureCubeIsOpen(cube))

		actions = append(actions, NewStepChain(func(d data.Data) []step.Step {
			b.logger.Debug("Transmuting items in the Horadric Cube")
			hid.MovePointer(ui.CubeTransmuteBtnX, ui.CubeTransmuteBtnY)
			helper.Sleep(150)
			hid.Click(hid.LeftButton)
			helper.Sleep(3000)

			// Move the Item back to the inventory
			hid.MovePointer(238, 262)
			hid.KeyDown("control")
			helper.Sleep(300)
			hid.Click(hid.LeftButton)
			helper.Sleep(200)
			hid.KeyUp("control")
			helper.Sleep(300)

			return []step.Step{
				step.SyncStepWithCheck(func(d data.Data) error {
					hid.PressKey("esc")
					helper.Sleep(300)
					return nil
				}, func(d data.Data) step.Status {
					if d.OpenMenus.Inventory {
						return step.StatusInProgress
					}
					return step.StatusCompleted
				}),
			}
		}))

		return
	})
}

func (b *Builder) ensureCubeIsOpen(cube data.Item) Action {
	return NewStepChain(func(d data.Data) []step.Step {
		b.logger.Debug("Opening Horadric Cube...")
		return []step.Step{
			step.SyncStepWithCheck(func(d data.Data) error {
				screenPos := ui.GetScreenCoordsForItem(cube)
				hid.MovePointer(screenPos.X, screenPos.Y)
				helper.Sleep(300)
				hid.Click(hid.RightButton)
				helper.Sleep(200)
				return nil
			}, func(d data.Data) step.Status {
				if d.OpenMenus.Cube {
					b.logger.Debug("Horadric Cube window detected")
					return step.StatusCompleted
				}
				return step.StatusInProgress
			}),
		}
	})
}
