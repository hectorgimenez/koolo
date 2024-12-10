package action

import (
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type SocketRecipe struct {
	Name          string
	Inserts       []string
	BaseItemTypes []string
	BaseSockets   int
}

var (
	SocketRecipes = []SocketRecipe{
		// Add recipes in order of priority. If we have inserts for two recipes, it will process the first one in the list first.
		// Add the inserts in order required for the runeword
		{
			Name:          "TirTir",
			Inserts:       []string{"TirRune", "TirRune"},
			BaseItemTypes: []string{"Helm"},
			BaseSockets:   2,
		},
		{
			Name:          "Stealth",
			Inserts:       []string{"TalRune", "EthRune"},
			BaseItemTypes: []string{"Armor"},
			BaseSockets:   2,
		},
		{
			Name:          "Lore",
			Inserts:       []string{"OrtRune", "SolRune"},
			BaseItemTypes: []string{"Helm"},
			BaseSockets:   2,
		},
		{
			Name:          "Ancients Pledge",
			Inserts:       []string{"RalRune", "OrtRune", "TalRune"},
			BaseItemTypes: []string{"Shield"},
			BaseSockets:   3,
		},
		{
			Name:          "Smoke",
			Inserts:       []string{"NefRune", "LumRune"},
			BaseItemTypes: []string{"Armor"},
			BaseSockets:   2,
		},
		{
			Name:          "Spirit sword",
			Inserts:       []string{"TalRune", "ThulRune", "OrtRune", "AmnRune"},
			BaseItemTypes: []string{"Sword"},
			BaseSockets:   4,
		},
		{
			Name:          "Spirit shield",
			Inserts:       []string{"TalRune", "ThulRune", "OrtRune", "AmnRune"},
			BaseItemTypes: []string{"Shield", "Auric Shields"},
			BaseSockets:   4,
		},
		{
			Name:          "Insight",
			Inserts:       []string{"RalRune", "TirRune", "TalRune", "SolRune"},
			BaseItemTypes: []string{"Polearm"},
			BaseSockets:   4,
		},
		{
			Name:          "Leaf",
			Inserts:       []string{"TirRune", "RalRune"},
			BaseItemTypes: []string{"Staff"},
			BaseSockets:   2,
		},
	}
)

func alreadyFilled(item data.Item) bool {

	// List of things that can appear on white items
	allowedStats := []stat.ID{
		stat.Defense,
		stat.MinDamage,
		stat.TwoHandedMinDamage,
		stat.MaxDamage,
		stat.TwoHandedMaxDamage,
		stat.AttackRate,
		stat.AttackRating,
		stat.EnhancedDamage,
		stat.EnhancedDamageMax,
		stat.Durability,
		stat.EnhancedDefense,
		stat.MaxDurabilityPercent,
		stat.MaxDurability,
		stat.ChanceToBlock,
		stat.FasterBlockRate,
		stat.NumSockets,
		stat.AddClassSkills,
		stat.NonClassSkill,
		stat.AddSkillTab,
		stat.AllSkills,
	}

	// Check each stat on the item
	for _, itemStat := range item.Stats {

		// Special case for Auric Shields with ColdResist
		if (itemStat.ID == stat.ColdResist || itemStat.ID == stat.FireResist || itemStat.ID == stat.LightningResist || itemStat.ID == stat.PoisonResist) && item.Type().Name == "Auric Shields" {
			continue // Allow AllResists on Auric Shields
		}

		statAllowed := false
		for _, allowed := range allowedStats {
			if itemStat.ID == allowed {
				statAllowed = true
				break
			}
		}

		if !statAllowed {
			return true // Item has unwanted stats
		}
	}

	return false // All stats are allowed
}

func hasBaseForSocketRecipe(items []data.Item, sockrecipe SocketRecipe) (data.Item, bool) {
	// Iterate through items to find matching base
	for _, item := range items {
		// Get item type
		itemType := item.Type().Name

		// Check if item type matches any of the allowed base types
		// TODO Allow for multiple bases in inventory and select the best one
		isValidType := false
		for _, baseType := range sockrecipe.BaseItemTypes {
			if itemType == baseType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			continue
		}

		// Check socket count
		sockets, found := item.FindStat(stat.NumSockets, 0)
		if !found || sockets.Value != sockrecipe.BaseSockets {
			continue
		}

		// Check if item has unwanted stats (already socketed/modified)
		if alreadyFilled(item) {
			continue
		}
		// Found valid base item
		return item, true
	}

	// No valid base found
	return data.Item{}, false
}

func hasItemsForSocketRecipe(items []data.Item, sockrecipe SocketRecipe) ([]data.Item, bool) {

	socketrecipeItems := make(map[string]int)
	for _, item := range sockrecipe.Inserts {
		socketrecipeItems[item]++
	}

	itemsForRecipe := []data.Item{}

	// Iterate over the items in our stash to see if we have the items for the recipie.
	for _, item := range items {
		if count, ok := socketrecipeItems[string(item.Name)]; ok {

			itemsForRecipe = append(itemsForRecipe, item)

			// Check if we now have exactly the needed count before decrementing
			count -= 1
			if count == 0 {
				delete(socketrecipeItems, string(item.Name))
				if len(socketrecipeItems) == 0 {
					return itemsForRecipe, true
				}
			} else {
				socketrecipeItems[string(item.Name)] = count
			}
		}
	}

	// We don't have all the items for the recipie.
	return nil, false
}
func SetSocketRecipes() error {
	ctx := context.Get()
	ctx.SetLastAction("SocketAddItems")

	insertsInStash := ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash)
	basesInStash := ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash)

	for _, recipe := range SocketRecipes {
		// Check if the current recipe is Enabled
		if !slices.Contains(ctx.CharacterCfg.Game.Leveling.EnabledSocketRecipes, recipe.Name) {
			continue
		}

		ctx.Logger.Debug("Socket recipe is enabled, processing", "recipe", recipe.Name)

		continueProcessing := true
		for continueProcessing {
			for _, recipe := range SocketRecipes {

				if baseItems, hasBase := hasBaseForSocketRecipe(basesInStash, recipe); hasBase {
					if inserts, hasInserts := hasItemsForSocketRecipe(insertsInStash, recipe); hasInserts {

						TakeItemsFromStash([]data.Item{baseItems})
						// Refresh game state
						baseToUse, _ := ctx.Data.Inventory.FindByID(baseItems.UnitID)
						TakeItemsFromStash(inserts)
						itemsToUse := inserts
						err := SocketItems(ctx, recipe, baseToUse, itemsToUse...)
						if err != nil {
							return err
						}

						insertsInStash = RemoveUsedItems(insertsInStash, inserts)
					} else {
						continueProcessing = false
					}
					// Remove or decrement the used items from basesInStash
					basesInStash = RemoveUsedItems(basesInStash, []data.Item{baseItems})
				} else {
					continueProcessing = false
				}
			}
		}
	}
	return nil
}

func SocketItems(ctx *context.Status, recipe SocketRecipe, base data.Item, items ...data.Item) error {

	ctx.SetLastAction("SocketItem")
	itemsInInv := ctx.Data.Inventory.ByLocation(item.LocationInventory)

	// Count required inserts
	requiredCounts := make(map[string]int)
	for _, insert := range recipe.Inserts {
		requiredCounts[insert]++
	}

	// Track which items we've used
	usedItems := make(map[*data.Item]bool)
	orderedItems := make([]data.Item, 0)

	// Process each required insert in order
	for _, requiredInsert := range recipe.Inserts {
		for i := range itemsInInv {
			item := &itemsInInv[i]
			if string(item.Name) == requiredInsert && !usedItems[item] {
				orderedItems = append(orderedItems, *item)
				usedItems[item] = true
				break
			}
		}
	}

	// Process items in correct order
	for _, itm := range orderedItems {

		basescreenPos := ui.GetScreenCoordsForItem(base)
		screenPos := ui.GetScreenCoordsForItem(itm)
		ctx.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
		utils.Sleep(300)

		ctx.HID.Click(game.LeftButton, basescreenPos.X, basescreenPos.Y)
	}

	utils.Sleep(300)

	return step.CloseAllMenus()
}
