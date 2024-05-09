package action

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/game"
)

type CubeRecipie struct {
	Name  string
	Items []string
}

var (
	recipies = []CubeRecipie{
		{
			Name:  "Perfect Amethyst",
			Items: []string{"FlawlessAmethyst", "FlawlessAmethyst", "FlawlessAmethyst"},
		},
		{
			Name:  "Perfect Diamond",
			Items: []string{"FlawlessDiamond", "FlawlessDiamond", "FlawlessDiamond"},
		},
		{
			Name:  "Perfect Emerald",
			Items: []string{"FlawlessEmerald", "FlawlessEmerald", "FlawlessEmerald"},
		},
		{
			Name:  "Perfect Ruby",
			Items: []string{"FlawlessRuby", "FlawlessRuby", "FlawlessRuby"},
		},
		{
			Name:  "Perfect Sapphire",
			Items: []string{"FlawlessSapphire", "FlawlessSapphire", "FlawlessSapphire"},
		},
		{
			Name:  "Perfect Topaz",
			Items: []string{"FlawlessTopaz", "FlawlessTopaz", "FlawlessTopaz"},
		},
		{
			Name:  "Perfect Skull",
			Items: []string{"FlawlessSkull", "FlawlessSkull", "FlawlessSkull"},
		},
	}
)

func (b *Builder) CubeRecipes() *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		itemsInStash := d.Items.ByLocation(item.LocationStash, item.LocationSharedStash1, item.LocationSharedStash2, item.LocationSharedStash3)
		for _, recipe := range recipies {
			continueProcessing := true
			for continueProcessing {
				if items, hasItems := b.hasItemsForRecipie(itemsInStash, recipe); hasItems {
					// Add items to the cube and perform the transmutation
					actions = append(actions, b.CubeAddItems(items...))
					actions = append(actions, b.CubeTransmute())

					// Add items to the stash
					actions = append(actions, b.Stash(true))

					// Remove or decrement the used items from itemsInStash
					itemsInStash = removeUsedItems(itemsInStash, items)
				} else {
					continueProcessing = false
				}
			}
		}
		return actions
	})
}

func (b *Builder) hasItemsForRecipie(items []data.Item, recipie CubeRecipie) ([]data.Item, bool) {
	// Create a map of the items we need for the recipie.
	recipieItems := make(map[string]int)
	for _, item := range recipie.Items {
		recipieItems[item]++
		b.Logger.Debug("Increasing number for: ", slog.String("Item", item))
	}

	itemsForRecipie := []data.Item{}

	// Iterate over the items in our stash to see if we have the items for the recipie.
	for _, item := range items {
		// If we have the item, decrement the count in the map.
		if count, ok := recipieItems[string(item.Name)]; ok {

			// If the count is now 0, we have all the items for the recipie.
			if count == 0 {
				delete(recipieItems, string(item.Name))

				// If the map is now empty, we have all the items for the recipie.
				if len(recipieItems) == 0 {
					b.Logger.Debug("Found all items for recipie: ", slog.String("Recipie", recipie.Name))

					return itemsForRecipie, true
				}
			}

			recipieItems[string(item.Name)] = count - 1
			itemsForRecipie = append(itemsForRecipie, item)
		}
	}

	// We don't have all the items for the recipie.
	return nil, false
}

func removeUsedItems(stash []data.Item, usedItems []data.Item) []data.Item {
	remainingItems := make([]data.Item, 0)
	usedItemMap := make(map[string]int)

	// Populate a map with the count of used items
	for _, item := range usedItems {
		usedItemMap[string(item.Name)] += 1 // Assuming 'ID' uniquely identifies items in 'usedItems'
	}

	// Filter the stash by excluding used items based on the count in the map
	for _, item := range stash {
		if count, exists := usedItemMap[string(item.Name)]; exists && count > 0 {
			usedItemMap[string(item.Name)] -= 1
		} else {
			remainingItems = append(remainingItems, item)
		}
	}

	return remainingItems
}
