package action

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
)

func (b *Builder) Gamble() *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		stashedGold, _ := d.PlayerUnit.FindStat(stat.StashGold, 0)
		if d.CharacterCfg.Gambling.Enabled && stashedGold.Value >= 2500000 {
			b.Logger.Info("Time to gamble! Visiting vendor...")

			openShopStep := step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_DOWN, win.VK_RETURN)
			vendorNPC := town.GetTownByArea(d.PlayerUnit.Area).GamblingNPC()

			// Jamella gamble button is the second one
			if vendorNPC == npc.Jamella {
				openShopStep = step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
			}

			// Fix for Anya position
			if vendorNPC == npc.Drehya {
				actions = append(actions, b.MoveToCoords(data.Position{
					X: 5107,
					Y: 5119,
				}))
			}

			return append(actions,
				b.InteractNPC(vendorNPC,
					openShopStep,
					step.Wait(time.Second),
				),
				b.gambleItems(),
			)
		}

		return nil
	})
}

func (b *Builder) gambleItems() *StepChainAction {
	var itemBought data.Item
	currentIdx := 0
	lastStep := false
	return NewStepChain(func(d game.Data) []step.Step {
		if lastStep {
			if d.OpenMenus.Inventory {
				return []step.Step{step.SyncStep(func(d game.Data) error {
					b.HID.PressKey(win.VK_ESCAPE)
					return nil
				})}
			}

			b.Logger.Info("Finished gambling", slog.Int("currentGold", d.PlayerUnit.TotalPlayerGold()))

			return nil
		}

		if itemBought.Name != "" {
			for _, itm := range d.Inventory.ByLocation(item.LocationInventory) {
				if itm.UnitID == itemBought.UnitID {
					itemBought = itm
					b.Logger.Debug("Gambled for item", slog.Any("item", itemBought))
					break
				}
			}

			if _, result := d.CharacterCfg.Runtime.Rules.EvaluateAll(itemBought); result == nip.RuleResultFullMatch {
				lastStep = true
				return []step.Step{step.Wait(time.Millisecond * 200)}
			} else {
				// Filter not pass, selling the item
				return []step.Step{step.SyncStep(func(d game.Data) error {
					b.sm.SellItem(itemBought)
					itemBought = data.Item{}
					return nil
				})}
			}
		}

		if d.PlayerUnit.TotalPlayerGold() < 500000 {
			lastStep = true
			return []step.Step{step.Wait(time.Millisecond * 200)}
		}

		for idx, itmName := range d.CharacterCfg.Gambling.Items {
			// Let's try to get one of each every time
			if currentIdx == len(d.CharacterCfg.Gambling.Items) {
				currentIdx = 0
			}

			if currentIdx > idx {
				continue
			}

			itm, found := d.Inventory.Find(itmName, item.LocationVendor)
			if !found {
				b.Logger.Debug("Item not found in gambling window, refreshing...", slog.String("item", string(itmName)))

				return []step.Step{step.SyncStep(func(d game.Data) error {
					if d.LegacyGraphics {
						b.HID.Click(game.LeftButton, ui.GambleRefreshButtonXClassic, ui.GambleRefreshButtonYClassic)
					} else {
						b.HID.Click(game.LeftButton, ui.GambleRefreshButtonX, ui.GambleRefreshButtonY)
					}
					return nil
				}),
					step.Wait(time.Millisecond * 500),
				}
			}

			return []step.Step{step.SyncStep(func(d game.Data) error {
				b.sm.BuyItem(itm, 1)
				itemBought = itm
				currentIdx++
				return nil
			})}
		}

		return nil
	}, RepeatUntilNoSteps())
}
