package action

import (
	"errors"
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/action/step"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/town"
	"github.com/hectorgimenez/koolo/internal/v2/ui"
	"github.com/hectorgimenez/koolo/internal/v2/utils"
	"github.com/lxn/win"
)

func Gamble() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "Gamble"

	stashedGold, _ := ctx.Data.PlayerUnit.FindStat(stat.StashGold, 0)
	if ctx.CharacterCfg.Gambling.Enabled && stashedGold.Value >= 2500000 {
		ctx.Logger.Info("Time to gamble! Visiting vendor...")

		vendorNPC := town.GetTownByArea(ctx.Data.PlayerUnit.Area).GamblingNPC()

		// Fix for Anya position
		if vendorNPC == npc.Drehya {
			_ = MoveToCoords(data.Position{
				X: 5107,
				Y: 5119,
			})
		}

		InteractNPC(vendorNPC)
		// Jamella gamble button is the second one
		if vendorNPC == npc.Jamella {
			ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
		} else {
			ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_DOWN, win.VK_RETURN)
		}

		if !ctx.Data.OpenMenus.NPCShop {
			return errors.New("failed opening gambling window")
		}

		return gambleItems()
	}

	return nil
}

func gambleItems() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "gambleItems"

	var itemBought data.Item
	currentIdx := 0
	lastStep := false
	for {
		if lastStep {
			utils.Sleep(200)
			ctx.Logger.Info("Finished gambling", slog.Int("currentGold", ctx.Data.PlayerUnit.TotalPlayerGold()))

			return step.CloseAllMenus()
		}

		if itemBought.Name != "" {
			for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
				if itm.UnitID == itemBought.UnitID {
					itemBought = itm
					ctx.Logger.Debug("Gambled for item", slog.Any("item", itemBought))
					break
				}
			}

			if _, result := ctx.Data.CharacterCfg.Runtime.Rules.EvaluateAll(itemBought); result == nip.RuleResultFullMatch {
				lastStep = true

			} else {
				// Filter not pass, selling the item
				town.SellItem(itemBought)
				itemBought = data.Item{}
			}
			continue
		}

		if ctx.Data.PlayerUnit.TotalPlayerGold() < 500000 {
			lastStep = true
			continue
		}

		for idx, itmName := range ctx.Data.CharacterCfg.Gambling.Items {
			// Let's try to get one of each every time
			if currentIdx == len(ctx.CharacterCfg.Gambling.Items) {
				currentIdx = 0
			}

			if currentIdx > idx {
				continue
			}

			itm, found := ctx.Data.Inventory.Find(itmName, item.LocationVendor)
			if !found {
				ctx.Logger.Debug("Item not found in gambling window, refreshing...", slog.String("item", string(itmName)))

				if ctx.Data.LegacyGraphics {
					ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonXClassic, ui.GambleRefreshButtonYClassic)
				} else {
					ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonX, ui.GambleRefreshButtonY)
				}

				utils.Sleep(500)
				continue
			}

			town.BuyItem(itm, 1)
			itemBought = itm
			currentIdx++
		}
	}
}
