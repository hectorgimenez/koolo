package action

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
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
	ctx.SetLastAction("Gamble")

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
	ctx.SetLastAction("GambleSingleItem")

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

// ItemGambleCount tracks the number of times each item has been gambled
type ItemGambleCount struct {
	counts map[item.Name]int
	mu     sync.Mutex
}

// NewItemGambleCount creates a new ItemGambleCount instance
func NewItemGambleCount() *ItemGambleCount {
	return &ItemGambleCount{
		counts: make(map[item.Name]int),
	}
}

// Increment increases the count for an item and returns true if the item is still eligible for gambling
func (i *ItemGambleCount) Increment(itemName item.Name) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.counts[itemName]++
	return i.counts[itemName] <= 10
}

// ShouldReset checks if all items have reached the threshold and resets if necessary
func (i *ItemGambleCount) ShouldReset(items []item.Name) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, name := range items {
		if i.counts[name] < 10 {
			return false
		}
	}

	// All items have reached 10 attempts, reset the counts
	i.counts = make(map[item.Name]int)
	return true
}

// GetCount returns the current count for an item
func (i *ItemGambleCount) GetCount(itemName item.Name) int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.counts[itemName]
}

func gambleItems(gameStartedAt time.Time) error {
	ctx := context.Get()
	ctx.SetLastAction("gambleItems")

	cleanup := func(itemBought data.Item) {
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
	itemCounts := NewItemGambleCount()

	for {
		if err := checkTimeLimit(gameStartedAt, ctx); err != nil {
			cleanup(itemBought)
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
					ctx.Logger.Debug("Gambled for item",
						slog.Any("item", itemBought),
						slog.Int("attempts", itemCounts.GetCount(itemBought.Name)))
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
			cleanup(itemBought)
			return err
		}

		if ctx.Data.PlayerUnit.TotalPlayerGold() < 500000 {
			lastStep = true
			continue
		}

		// Check if we need to reset the counts
		if itemCounts.ShouldReset(ctx.Data.CharacterCfg.Gambling.Items) {
			ctx.Logger.Info("Reset gambling counts - all items reached 10 attempts")
		}

		var foundItem bool
		for _, itmName := range ctx.Data.CharacterCfg.Gambling.Items {
			if itm, found := ctx.Data.Inventory.Find(itmName, item.LocationVendor); found {
				// Only buy if we haven't reached the limit for this item
				if itemCounts.Increment(itmName) {
					town.BuyItem(itm, 1)
					itemBought = itm
					foundItem = true
					ctx.Logger.Debug("Found and bought gambling item",
						slog.String("item", string(itmName)),
						slog.Int("attempt", itemCounts.GetCount(itmName)))
					break
				}
			}
		}

		if !foundItem {
			ctx.Logger.Debug("No eligible items found in gambling window, refreshing...",
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
