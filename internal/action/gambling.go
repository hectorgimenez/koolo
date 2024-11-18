package action

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

const maxGamblingDuration = 10 * time.Minute

func checkTimeLimit(gameStartedAt time.Time, ctx *context.Status) error {
	if time.Since(gameStartedAt) > maxGamblingDuration {
		ctx.Logger.Info("Max gambling duration reached, cleaning up...",
			slog.Float64("duration_seconds", time.Since(gameStartedAt).Seconds()))

		if err := step.CloseAllMenus(); err != nil {
			ctx.Logger.Error("Failed to close menus during timeout cleanup", slog.String("error", err.Error()))
		}

		return fmt.Errorf(
			"max gambling duration reached: %0.2f seconds",
			time.Since(gameStartedAt).Seconds(),
		)
	}
	return nil
}

func Gamble(gameStartedAt time.Time) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "Gamble"

	if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
		return err
	}

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
		// Check time before interacting with NPC
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			return err
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
		// Check time before gambling
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			return err
		}
		return gambleItems(gameStartedAt)
	}

	return nil
}

func GambleSingleItem(items []string, desiredQuality item.Quality, gameStartedAt time.Time) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "GambleSingleItem"

	cleanup := func(itemBought data.Item) {
		// If we have a bought item during timeout, try to sell it
		if itemBought.Name != "" {
			ctx.Logger.Info("Selling item before timeout cleanup", slog.Any("item", itemBought))
			town.SellItem(itemBought)
		}

		if err := step.CloseAllMenus(); err != nil {
			ctx.Logger.Error("Failed to close menus during cleanup", slog.String("error", err.Error()))
		}
	}

	if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
		cleanup(data.Item{})
		return err
	}

	charGold := ctx.Data.PlayerUnit.TotalPlayerGold()
	var itemBought data.Item

	// Check if we have enough gold to gamble
	if charGold >= 150000 {
		ctx.Logger.Info("Gambling for items", slog.Any("items", items))

		vendorNPC := town.GetTownByArea(ctx.Data.PlayerUnit.Area).GamblingNPC()

		// Fix for Anya position
		if vendorNPC == npc.Drehya {
			_ = MoveToCoords(data.Position{
				X: 5107,
				Y: 5119,
			})
		}
		// Check time before interacting with NPC
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			cleanup(data.Item{})
			return err
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
	}

	for {
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			cleanup(itemBought)
			return err
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
				ctx.Logger.Info("Found item matching nip rules, will be kept", slog.Any("item", itemBought))
				itemBought = data.Item{}
				continue
			} else {
				if itemBought.Quality == desiredQuality {
					ctx.Logger.Info("Found item matching desired quality, will be kept", slog.Any("item", itemBought))
					return step.CloseAllMenus()
				} else {
					town.SellItem(itemBought)
					itemBought = data.Item{}
				}
			}
		}

		if ctx.Data.PlayerUnit.TotalPlayerGold() < 150000 {
			cleanup(data.Item{})
			return errors.New("gold is below 150000, stopping gamble")
		}

		// Check time before interacting with NPC
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			cleanup(data.Item{})
			return err
		}

		for _, itmName := range items {
			itm, found := ctx.Data.Inventory.Find(item.Name(itmName), item.LocationVendor)
			if found {
				town.BuyItem(itm, 1)
				itemBought = itm
				break
			}
		}

		if itemBought.Name == "" {
			ctx.Logger.Debug("Desired items not found in gambling window, refreshing...", slog.Any("items", items))
			RefreshGamblingWindow(ctx)
			utils.Sleep(500)
		}

		// Check time before interacting with NPC
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			cleanup(data.Item{})
			return err
		}
	}
}

func gambleItems(gameStartedAt time.Time) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "gambleItems"

	cleanup := func(itemBought data.Item) {
		// If we have a bought item during timeout, try to sell it
		if itemBought.Name != "" {
			ctx.Logger.Info("Selling item before timeout cleanup", slog.Any("item", itemBought))
			town.SellItem(itemBought)
		}

		if err := step.CloseAllMenus(); err != nil {
			ctx.Logger.Error("Failed to close menus during cleanup", slog.String("error", err.Error()))
		}
	}

	if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
		cleanup(data.Item{})
		return err
	}

	var itemBought data.Item
	lastStep := false

	for {
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			cleanup(data.Item{})
			return err
		}

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
				ctx.Logger.Info("Found item matching NIP rules, keeping", slog.Any("item", itemBought))
				lastStep = true
			} else {
				ctx.Logger.Debug("Item doesn't match NIP rules, selling", slog.Any("item", itemBought))
				town.SellItem(itemBought)
			}
			itemBought = data.Item{}
			continue
		}

		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			cleanup(data.Item{})
			return err
		}

		if ctx.Data.PlayerUnit.TotalPlayerGold() < 500000 {
			lastStep = true
			continue
		}

		// Check for any desired item in the current gambling window
		var foundItem bool
		for _, itmName := range ctx.Data.CharacterCfg.Gambling.Items {
			if itm, found := ctx.Data.Inventory.Find(itmName, item.LocationVendor); found {
				town.BuyItem(itm, 1)
				itemBought = itm
				foundItem = true
				ctx.Logger.Debug("Found and bought gambling item", slog.String("item", string(itmName)))
				break
			}
		}

		// Only refresh if no desired items were found
		if !foundItem {
			ctx.Logger.Debug("No desired items found in gambling window, refreshing...",
				slog.Any("searching_for", ctx.Data.CharacterCfg.Gambling.Items))
			RefreshGamblingWindow(ctx)
			utils.Sleep(500)
		}
	}
}

func RefreshGamblingWindow(ctx *context.Status) {
	if ctx.Data.LegacyGraphics {
		ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonXClassic, ui.GambleRefreshButtonYClassic)
	} else {
		ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonX, ui.GambleRefreshButtonY)
	}
}
