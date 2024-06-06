package action

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
)

func (b *Builder) CubeAddItems(items ...data.Item) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		cube, found := d.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if !found {
			b.Logger.Info("No Horadric Cube found in inventory")
			return nil
		}

		// Ensure stash is open
		if !d.OpenMenus.Stash {
			actions = append(actions, b.InteractObject(object.Bank, func(d game.Data) bool {
				return d.OpenMenus.Stash
			}))
		}

		b.Logger.Info("Adding items to the Horadric Cube", slog.Any("items", items))

		// If items are on the Stash, pickup them to the inventory
		for _, itm := range items {
			nwIt := itm
			if nwIt.Location.LocationType != item.LocationStash && nwIt.Location.LocationType != item.LocationSharedStash {
				continue
			}

			// Check in which tab the item is and switch to it
			switch nwIt.Location.LocationType {
			case item.LocationStash:
				actions = append(actions, b.SwitchStashTab(1))
			case item.LocationSharedStash:
				actions = append(actions, b.SwitchStashTab(nwIt.Location.Page+1))
			}

			b.Logger.Debug("Item found on the stash, picking it up", slog.String("Item", string(nwIt.Name)))
			actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
				screenPos := b.UIManager.GetScreenCoordsForItem(nwIt)

				b.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
				helper.Sleep(300)

				return nil
			}))
		}

		actions = append(actions, b.ensureCubeIsOpen(cube))

		for _, itm := range items {
			nwIt := itm
			actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
				for _, updatedItem := range d.Inventory.AllItems {
					if nwIt.UnitID == updatedItem.UnitID {
						b.Logger.Debug("Moving Item to the Horadric Cube", slog.String("Item", string(nwIt.Name)))

						screenPos := b.UIManager.GetScreenCoordsForItem(updatedItem)

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
	return NewChain(func(d game.Data) (actions []Action) {
		cube, found := d.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if !found {
			b.Logger.Info("No Horadric Cube found in inventory")
			return nil
		}

		actions = append(actions, b.ensureCubeIsOpen(cube))

		actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
			b.Logger.Debug("Transmuting items in the Horadric Cube")
			helper.Sleep(150)

			if d.LegacyGraphics {
				b.HID.Click(game.LeftButton, ui.CubeTransmuteBtnXClassic, ui.CubeTransmuteBtnYClassic)
			} else {
				b.HID.Click(game.LeftButton, ui.CubeTransmuteBtnX, ui.CubeTransmuteBtnY)
			}

			helper.Sleep(2000)

			if d.LegacyGraphics {
				b.HID.ClickWithModifier(game.LeftButton, ui.CubeTakeItemXClassic, ui.CubeTakeItemYClassic, game.CtrlKey)
			} else {
				b.HID.ClickWithModifier(game.LeftButton, ui.CubeTakeItemX, ui.CubeTakeItemY, game.CtrlKey)
			}

			helper.Sleep(300)

			return []step.Step{
				step.SyncStepWithCheck(func(d game.Data) error {
					b.HID.PressKey(win.VK_ESCAPE)
					helper.Sleep(300)
					return nil
				}, func(d game.Data) step.Status {
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
	return NewStepChain(func(d game.Data) []step.Step {
		b.Logger.Debug("Opening Horadric Cube...")
		return []step.Step{
			step.SyncStepWithCheck(func(d game.Data) error {
				// Switch to the tab
				b.switchTab(cube.Location.Page + 1)

				screenPos := b.UIManager.GetScreenCoordsForItem(cube)

				helper.Sleep(300)
				b.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
				helper.Sleep(200)
				return nil
			}, func(d game.Data) step.Status {
				if d.OpenMenus.Cube {
					b.Logger.Debug("Horadric Cube window detected")
					return step.StatusCompleted
				}
				return step.StatusInProgress
			}),
		}
	})
}
