package action

import (
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/game"
)

type CubeRecipe struct {
	Name  string
	Items []string
}

var (
	recipies = []CubeRecipe{

		// Perfects
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

		// Token
		{
			Name:  "Token of Absolution",
			Items: []string{"TwistedEssenceOfSuffering", "ChargedEssenceOfHatred", "BurningEssenceOfTerror", "FesteringEssenceOfDestruction"},
		},

		// Runes
		{
			Name:  "Upgrade El",
			Items: []string{"ElRune", "ElRune", "ElRune"},
		},
		{
			Name:  "Upgrade Eld",
			Items: []string{"EldRune", "EldRune", "EldRune"},
		},
		{
			Name:  "Upgrade Tir",
			Items: []string{"TirRune", "TirRune", "TirRune"},
		},
		{
			Name:  "Upgrade Nef",
			Items: []string{"NefRune", "NefRune", "NefRune"},
		},
		{
			Name:  "Upgrade Eth",
			Items: []string{"EthRune", "EthRune", "EthRune"},
		},
		{
			Name:  "Upgrade Ith",
			Items: []string{"IthRune", "IthRune", "IthRune"},
		},
		{
			Name:  "Upgrade Tal",
			Items: []string{"TalRune", "TalRune", "TalRune"},
		},
		{
			Name:  "Upgrade Ral",
			Items: []string{"RalRune", "RalRune", "RalRune"},
		},
		{
			Name:  "Upgrade Ort",
			Items: []string{"OrtRune", "OrtRune", "OrtRune"},
		},
		{
			Name:  "Upgrade Thul",
			Items: []string{"ThulRune", "ThulRune", "ThulRune", "ChippedTopaz"},
		},
		{
			Name:  "Upgrade Amn",
			Items: []string{"AmnRune", "AmnRune", "AmnRune", "ChippedAmethyst"},
		},
		{
			Name:  "Upgrade Sol",
			Items: []string{"SolRune", "SolRune", "SolRune", "ChippedSapphire"},
		},
		{
			Name:  "Upgrade Shael",
			Items: []string{"ShaelRune", "ShaelRune", "ShaelRune", "ChippedRuby"},
		},
		{
			Name:  "Upgrade Dol",
			Items: []string{"DolRune", "DolRune", "DolRune", "ChippedEmerald"},
		},
		{
			Name:  "Upgrade Hel",
			Items: []string{"HelRune", "HelRune", "HelRune", "ChippedDiamond"},
		},
		{
			Name:  "Upgrade Io",
			Items: []string{"IoRune", "IoRune", "IoRune", "FlawedTopaz"},
		},
		{
			Name:  "Upgrade Lum",
			Items: []string{"LumRune", "LumRune", "LumRune", "FlawedAmethyst"},
		},
		{
			Name:  "Upgrade Ko",
			Items: []string{"KoRune", "KoRune", "KoRune", "FlawedSapphire"},
		},
		{
			Name:  "Upgrade Fal",
			Items: []string{"FalRune", "FalRune", "FalRune", "FlawedRuby"},
		},
		{
			Name:  "Upgrade Lem",
			Items: []string{"LemRune", "LemRune", "LemRune", "FlawedEmerald"},
		},
		{
			Name:  "Upgrade Pul",
			Items: []string{"PulRune", "PulRune", "FlawedDiamond"},
		},
		{
			Name:  "Upgrade Um",
			Items: []string{"UmRune", "UmRune", "Topaz"},
		},
		{
			Name:  "Upgrade Mal",
			Items: []string{"MalRune", "MalRune", "Amethyst"},
		},
		{
			Name:  "Upgrade Ist",
			Items: []string{"IstRune", "IstRune", "Sapphire"},
		},
		{
			Name:  "Upgrade Gul",
			Items: []string{"GulRune", "GulRune", "Ruby"},
		},
		{
			Name:  "Upgrade Vex",
			Items: []string{"VexRune", "VexRune", "Emerald"},
		},
		{
			Name:  "Upgrade Ohm",
			Items: []string{"OhmRune", "OhmRune", "Diamond"},
		},
		{
			Name:  "Upgrade Lo",
			Items: []string{"LoRune", "LoRune", "FlawlessTopaz"},
		},
		{
			Name:  "Upgrade Sur",
			Items: []string{"SurRune", "SurRune", "FlawlessAmethyst"},
		},
		{
			Name:  "Upgrade Ber",
			Items: []string{"BerRune", "BerRune", "FlawlessSapphire"},
		},
		{
			Name:  "Upgrade Jah",
			Items: []string{"JahRune", "JahRune", "FlawlessRuby"},
		},
		{
			Name:  "Upgrade Cham",
			Items: []string{"ChamRune", "ChamRune", "FlawlessEmerald"},
		},
	}
)

func (b *Builder) CubeRecipes() *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		// If cubing is disabled from settings just return nil
		if !b.CharacterCfg.CubeRecipes.Enabled {
			return nil
		}

		itemsInStash := d.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash)
		for _, recipe := range recipies {

			// Check if the current recipe is Enabled
			if !slices.Contains(b.CharacterCfg.CubeRecipes.EnabledRecipes, recipe.Name) {
				continue
			}

			continueProcessing := true
			for continueProcessing {
				if items, hasItems := b.hasItemsForRecipe(itemsInStash, recipe); hasItems {
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

func (b *Builder) hasItemsForRecipe(items []data.Item, recipe CubeRecipe) ([]data.Item, bool) {
	// Create a map of the items we need for the recipie.
	recipeItems := make(map[string]int)
	for _, item := range recipe.Items {
		recipeItems[item]++
	}

	itemsForRecipe := []data.Item{}

	// Iterate over the items in our stash to see if we have the items for the recipie.
	for _, item := range items {
		if count, ok := recipeItems[string(item.Name)]; ok {
			itemsForRecipe = append(itemsForRecipe, item)

			// Check if we now have exactly the needed count before decrementing
			count -= 1
			if count == 0 {
				delete(recipeItems, string(item.Name))
				if len(recipeItems) == 0 {
					return itemsForRecipe, true
				}
			} else {
				recipeItems[string(item.Name)] = count
			}
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
