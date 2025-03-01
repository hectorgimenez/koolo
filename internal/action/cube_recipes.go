package action

import (
	"slices"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type CubeRecipe struct {
	Name             string
	Items            []string
	PurchaseRequired bool
	PurchaseItems    []string
}

var (
	Recipes = []CubeRecipe{

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

		// Crafting
		{
			Name:  "Reroll GrandCharms",
			Items: []string{"GrandCharm", "Perfect", "Perfect", "Perfect"}, // Special handling in hasItemsForRecipe
		},

		// Caster Amulet
		{
			Name:             "Caster Amulet",
			Items:            []string{"RalRune", "PerfectAmethyst", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Amulet"},
		},

		// Caster Ring
		{
			Name:             "Caster Ring",
			Items:            []string{"AmnRune", "PerfectAmethyst", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Ring"},
		},

		// Blood Gloves
		{
			Name:             "Blood Gloves",
			Items:            []string{"NefRune", "PerfectRuby", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"HeavyGloves", "SharkskinGloves", "VampireboneGloves"},
		},

		// Blood Boots
		{
			Name:             "Blood Boots",
			Items:            []string{"EthRune", "PerfectRuby", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"LightPlatedBoots", "BattleBoots", "MirroredBoots"},
		},

		// Blood Belt
		{
			Name:             "Blood Belt",
			Items:            []string{"TalRune", "PerfectRuby", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Belt", "MeshBelt", "MithrilCoil"},
		},

		// Blood Helm
		{
			Name:             "Blood Helm",
			Items:            []string{"RalRune", "PerfectRuby", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Helm", "Casque", "Armet"},
		},

		// Blood Armor
		{
			Name:             "Blood Armor",
			Items:            []string{"ThulRune", "PerfectRuby", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"PlateMail", "TemplarPlate", "HellforgePlate"},
		},

		// Blood Weapon
		{
			Name:             "Blood Weapon",
			Items:            []string{"OrtRune", "PerfectRuby", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Axe"},
		},

		// Safety Shield
		{
			Name:             "Safety Shield",
			Items:            []string{"NefRune", "PerfectEmerald", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"KiteShield", "DragonShield", "Monarch"},
		},

		// Safety Armor
		{
			Name:             "Safety Armor",
			Items:            []string{"EthRune", "PerfectEmerald", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"BreastPlate", "Curiass", "GreatHauberk"},
		},

		// Safety Boots
		{
			Name:             "Safety Boots",
			Items:            []string{"OrtRune", "PerfectEmerald", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Greaves", "WarBoots", "MyrmidonBoots"},
		},

		// Safety Gloves
		{
			Name:             "Safety Gloves",
			Items:            []string{"RalRune", "PerfectEmerald", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Gauntlets", "WarGauntlets", "OgreGauntlets"},
		},

		// Safety Belt
		{
			Name:             "Safety Belt",
			Items:            []string{"TalRune", "PerfectEmerald", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Sash", "DemonhideSash", "SpiderwebSash"},
		},

		// Safety Helm
		{
			Name:             "Safety Helm",
			Items:            []string{"IthRune", "PerfectEmerald", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"Crown", "GrandCrown", "Corona"},
		},

		// Hitpower Gloves
		{
			Name:             "Hitpower Gloves",
			Items:            []string{"OrtRune", "PerfectSapphire", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"ChainGloves", "HeavyBracers", "Vambraces"},
		},

		// Hitpower Boots
		{
			Name:             "Hitpower Boots",
			Items:            []string{"RalRune", "PerfectSapphire", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"ChainBoots", "MeshBoots", "Boneweave"},
		},

		// Hitpower Belt
		{
			Name:             "Hitpower Belt",
			Items:            []string{"TalRune", "PerfectSapphire", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"HeavyBelt", "BattleBelt", "TrollBelt"},
		},

		// Hitpower Helm
		{
			Name:             "Hitpower Helm",
			Items:            []string{"NefRune", "PerfectSapphire", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"FullHelm", "Basinet", "GiantConch"},
		},

		// Hitpower Armor
		{
			Name:             "Hitpower Armor",
			Items:            []string{"EthRune", "PerfectSapphire", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"FieldPlate", "Sharktooth", "KrakenShell"},
		},

		// Hitpower Shield
		{
			Name:             "Hitpower Shield",
			Items:            []string{"IthRune", "PerfectSapphire", "Jewel"},
			PurchaseRequired: true,
			PurchaseItems:    []string{"GothicShield", "AncientShield", "Ward"},
		},
	}
)

func CubeRecipes() error {
	ctx := context.Get()
	ctx.SetLastAction("CubeRecipes")

	// If cubing is disabled from settings just return nil
	if !ctx.CharacterCfg.CubeRecipes.Enabled {
		ctx.Logger.Debug("Cube recipes are disabled, skipping")
		return nil
	}

	ingredientSources := []item.LocationType{item.LocationSharedStash}
	if ctx.CharacterCfg.CubeRecipes.IncludePersonalStashForCubing {
		ingredientSources = append(ingredientSources, item.LocationStash)
	}

	itemsInStash := ctx.Data.Inventory.ByLocation(ingredientSources...)

	for _, recipe := range Recipes {
		// Check if the current recipe is Enabled
		if !slices.Contains(ctx.CharacterCfg.CubeRecipes.EnabledRecipes, recipe.Name) {
			// is this really needed ? making huge logs
			//		ctx.Logger.Debug("Cube recipe is not enabled, skipping", "recipe", recipe.Name)
			continue
		}

		ctx.Logger.Debug("Cube recipe is enabled, processing", "recipe", recipe.Name)

		continueProcessing := true
		for continueProcessing {
			if items, hasItems := hasItemsForRecipe(ctx, recipe); hasItems {

				// TODO: Check if we have the items in our storage and if not, purchase them, else take the item from the storage
				if recipe.PurchaseRequired {
					err := GambleSingleItem(recipe.PurchaseItems, item.QualityMagic)
					if err != nil {
						ctx.Logger.Error("Error gambling item, skipping recipe", "error", err, "recipe", recipe.Name)
						break
					}

					purchasedItem := getPurchasedItem(ctx, recipe.PurchaseItems)
					if purchasedItem.Name == "" {
						ctx.Logger.Debug("Could not find purchased item. Skipping recipe", "recipe", recipe.Name)
						break
					}

					// Add the purchased item the list of items to cube
					items = append(items, purchasedItem)
				}

				// Add items to the cube and perform the transmutation
				err := CubeAddItems(items...)
				if err != nil {
					return err
				}
				if err = CubeTransmute(); err != nil {
					return err
				}

				// Get a list of items that are in our inventory
				itemsInInv := ctx.Data.Inventory.ByLocation(item.LocationInventory)

				stashingRequired := false
				stashingGrandCharm := false

				// Check if the items that are not in the protected inventory slots should be stashed
				for _, item := range itemsInInv {
					// If item is not in the protected slots, check if it should be stashed
					if ctx.CharacterCfg.Inventory.InventoryLock[item.Position.Y][item.Position.X] == 1 {
						if item.Name == "Key" {
							continue
						}

						shouldStash, reason, _ := shouldStashIt(item, false)

						if shouldStash {
							ctx.Logger.Debug("Stashing item after cube recipe.", "item", item.Name, "recipe", recipe.Name, "reason", reason)
							stashingRequired = true
						} else if item.Name == "GrandCharm" {
							ctx.Logger.Debug("Checking if we need to stash a GrandCharm that doesn't match any NIP rules.", "recipe", recipe.Name)
							// Check if we have a GrandCharm in stash that doesn't match any NIP rules
							hasUnmatchedGrandCharm := false
							for _, stashItem := range itemsInStash {
								if stashItem.Name == "GrandCharm" {
									if _, result := ctx.CharacterCfg.Runtime.Rules.EvaluateAll(stashItem); result != nip.RuleResultFullMatch {
										hasUnmatchedGrandCharm = true
										break
									}
								}
							}
							if !hasUnmatchedGrandCharm {

								ctx.Logger.Debug("GrandCharm doesn't match any NIP rules and we don't have any in stash to be used for this recipe. Stashing it.", "recipe", recipe.Name)
								stashingRequired = true
								stashingGrandCharm = true

							} else {
								DropInventoryItem(item)
								utils.Sleep(500)
							}
						} else {
							DropInventoryItem(item)
							utils.Sleep(500)
						}
					}
				}

				// Add items to the stash if needed
				if stashingRequired && !stashingGrandCharm {
					_ = Stash(false)
				} else if stashingGrandCharm {
					// Force stashing of the invetory
					_ = Stash(true)
				}

				// Remove or decrement the used items from itemsInStash
				itemsInStash = removeUsedItems(itemsInStash, items)
			} else {
				continueProcessing = false
			}
		}
	}

	return nil
}

func hasItemsForRecipe(ctx *context.Status, recipe CubeRecipe) ([]data.Item, bool) {

	ctx.RefreshGameData()

	ingredientSources := []item.LocationType{item.LocationSharedStash}
	if ctx.CharacterCfg.CubeRecipes.IncludePersonalStashForCubing {
		ingredientSources = append(ingredientSources, item.LocationStash)
	}

	items := ctx.Data.Inventory.ByLocation(ingredientSources...)
	// Special handling for "Reroll GrandCharms" recipe
	if recipe.Name == "Reroll GrandCharms" {
		return hasItemsForGrandCharmReroll(ctx, items)
	}

	recipeItems := make(map[string]int)
	for _, item := range recipe.Items {
		recipeItems[item]++
	}

	itemsForRecipe := []data.Item{}

	// Figure out all of the items we have in our stashes as well as their count
	itemsByType := map[string][]data.Item{}
	for _, item := range items {
		// Let's make sure we don't use an item we don't want to. Add more if needed (depending on the recipes we have)
		if item.Name == "Jewel" {
			if _, result := ctx.CharacterCfg.Runtime.Rules.EvaluateAll(item); result == nip.RuleResultFullMatch {
				continue
			}
		}
		itemsByType[string(item.Name)] = append(itemsByType[string(item.Name)], item)
	}

	// Check to see if we have all of the items needed for the recipe
	for itemName, countNeededForRecipe := range recipeItems {
		if itemList, ok := itemsByType[itemName]; ok {
			countInStash := len(itemList)

			bufferCount := 0
			if strings.HasPrefix(recipe.Name, "Upgrade") {
				// Imprecise check to see if this is a rune upgrade recipe
				bufferCount = ctx.CharacterCfg.CubeRecipes.BufferRunes
			}

			if countInStash >= (countNeededForRecipe + bufferCount) {
				itemsForRecipe = append(itemsForRecipe, itemList[:countNeededForRecipe]...)
				delete(recipeItems, itemName)
				if len(recipeItems) == 0 { // Was this the last item-type for the recipe?
					return itemsForRecipe, true
				}
			}
			// We don't have enough for this item-type in the recipe
		}
		return nil, false
	}
	return nil, false
}

func hasItemsForGrandCharmReroll(ctx *context.Status, items []data.Item) ([]data.Item, bool) {
	var grandCharm data.Item
	perfectGems := make([]data.Item, 0, 3)

	for _, itm := range items {
		if itm.Name == "GrandCharm" {
			if _, result := ctx.CharacterCfg.Runtime.Rules.EvaluateAll(itm); result != nip.RuleResultFullMatch && itm.Quality == item.QualityMagic {
				grandCharm = itm
			}
		} else if isPerfectGem(itm) && len(perfectGems) < 3 {
			// Skip perfect amethysts and rubies if configured
			if (ctx.CharacterCfg.CubeRecipes.SkipPerfectAmethysts && itm.Name == "PerfectAmethyst") ||
				(ctx.CharacterCfg.CubeRecipes.SkipPerfectRubies && itm.Name == "PerfectRuby") {
				continue
			}
			perfectGems = append(perfectGems, itm)
		}

		if grandCharm.Name != "" && len(perfectGems) == 3 {
			return append([]data.Item{grandCharm}, perfectGems...), true
		}
	}

	return nil, false
}

func isPerfectGem(item data.Item) bool {
	perfectGems := []string{"PerfectAmethyst", "PerfectDiamond", "PerfectEmerald", "PerfectRuby", "PerfectSapphire", "PerfectTopaz", "PerfectSkull"}
	for _, gemName := range perfectGems {
		if string(item.Name) == gemName {
			return true
		}
	}
	return false
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

func getPurchasedItem(ctx *context.Status, purchaseItems []string) data.Item {
	itemsInInv := ctx.Data.Inventory.ByLocation(item.LocationInventory)
	for _, citem := range itemsInInv {
		for _, pi := range purchaseItems {
			if string(citem.Name) == pi && citem.Quality == item.QualityMagic {
				return citem
			}
		}
	}
	return data.Item{}
}
