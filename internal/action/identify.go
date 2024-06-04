package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/lxn/win"
)

func (b *Builder) IdentifyAll(skipIdentify bool) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		items := b.itemsToIdentify(d)

		b.Logger.Debug("Checking for items to identify...")
		if len(items) == 0 || skipIdentify {
			b.Logger.Debug("No items to identify...")
			return
		}

		idTome, found := d.Inventory.Find(item.TomeOfIdentify, item.LocationInventory)
		if !found {
			b.Logger.Warn("ID Tome not found, not identifying items")
			return
		}

		if st, statFound := idTome.FindStat(stat.Quantity, 0); !statFound || st.Value < len(items) {
			b.Logger.Info("Not enough ID scrolls, refilling...")
			actions = append(actions, b.VendorRefill(true, false))
		}

		b.Logger.Info(fmt.Sprintf("Identifying %d items...", len(items)))
		actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
			return []step.Step{
				step.SyncStepWithCheck(func(d game.Data) error {
					b.HID.PressKeyBinding(d.KeyBindings.Inventory)
					return nil
				}, func(d game.Data) step.Status {
					if d.OpenMenus.Inventory {
						return step.StatusCompleted
					}
					return step.StatusInProgress
				}),
				step.SyncStep(func(d game.Data) error {

					for _, i := range items {
						b.identifyItem(idTome, i)
					}

					b.HID.PressKey(win.VK_ESCAPE)

					return nil
				}),
			}
		}))

		return
	}, Resettable(), CanBeSkipped())
}

func (b *Builder) itemsToIdentify(d game.Data) (items []data.Item) {
	for _, i := range d.Inventory.ByLocation(item.LocationInventory) {
		if i.Identified || i.Quality == item.QualityNormal || i.Quality == item.QualitySuperior {
			continue
		}

		// Skip identifying items that fully match a rule when unid
		if _, result := b.CharacterCfg.Runtime.Rules.EvaluateAll(i); result == nip.RuleResultFullMatch {
			continue
		}

		items = append(items, i)
	}

	return
}

func (b *Builder) identifyItem(idTome data.Item, i data.Item) {
	screenPos := b.UIManager.GetScreenCoordsForItem(idTome)

	helper.Sleep(500)
	b.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
	helper.Sleep(1000)

	screenPos = b.UIManager.GetScreenCoordsForItem(i)

	b.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
	helper.Sleep(350)
}
