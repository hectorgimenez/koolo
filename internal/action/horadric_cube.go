package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/ui"
	"log/slog"
)

func (b *Builder) CubeAddItems(items ...data.Item) *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		cube, found := d.Items.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if !found {
			b.Logger.Info("No Horadric Cube found in inventory")
			return nil
		}

		// Ensure stash is open
		if !d.OpenMenus.Stash {
			actions = append(actions, b.InteractObject(object.Bank, func(d data.Data) bool {
				return d.OpenMenus.Stash
			}))
		}

		b.Logger.Info("Adding items to the Horadric Cube", slog.Any("items", items))

		// If items are on the Stash, pickup them to the inventory (only personal stash is supported for now)
		for _, itm := range items {
			nwIt := itm
			if nwIt.Location != item.LocationStash && nwIt.Location != item.LocationSharedStash1 && nwIt.Location != item.LocationSharedStash2 && nwIt.Location != item.LocationSharedStash3 {
				continue
			}

			b.Logger.Debug("Item found on the stash, picking it up", slog.String("Item", string(nwIt.Name)))
			actions = append(actions, NewStepChain(func(d data.Data) []step.Step {
				screenPos := ui.GetScreenCoordsForItem(nwIt)
				b.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
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
						b.Logger.Debug("Moving Item to the Horadric Cube", slog.String("Item", string(nwIt.Name)))
						screenPos := ui.GetScreenCoordsForItem(updatedItem)
						b.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
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
			b.Logger.Info("No Horadric Cube found in inventory")
			return nil
		}

		actions = append(actions, b.ensureCubeIsOpen(cube))

		actions = append(actions, NewStepChain(func(d data.Data) []step.Step {
			b.Logger.Debug("Transmuting items in the Horadric Cube")
			helper.Sleep(150)
			b.HID.Click(game.LeftButton, ui.CubeTransmuteBtnX, ui.CubeTransmuteBtnY)
			helper.Sleep(3000)

			// Move the Item back to the inventory
			b.HID.ClickWithModifier(game.LeftButton, 238, 262, game.CtrlKey)
			helper.Sleep(300)

			return []step.Step{
				step.SyncStepWithCheck(func(d data.Data) error {
					b.HID.PressKey("esc")
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
		b.Logger.Debug("Opening Horadric Cube...")
		return []step.Step{
			step.SyncStepWithCheck(func(d data.Data) error {
				screenPos := ui.GetScreenCoordsForItem(cube)
				helper.Sleep(300)
				b.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
				helper.Sleep(200)
				return nil
			}, func(d data.Data) step.Status {
				if d.OpenMenus.Cube {
					b.Logger.Debug("Horadric Cube window detected")
					return step.StatusCompleted
				}
				return step.StatusInProgress
			}),
		}
	})
}
