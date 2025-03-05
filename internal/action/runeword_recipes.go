package action

import (
	"fmt"
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

type RunewordRecipe struct {
	Name          string
	Inserts       []string
	BaseItemTypes []string
	BaseSortOrder []stat.ID
	AllowEth      bool
}

var (
	RunewordRecipes = []RunewordRecipe{
		// Add recipes in order of priority. If we have inserts for two recipes, it will process the first one in the list first.
		// Add the inserts in order required for the runeword
		{
			Name:          "Chains of Honor",
			Inserts:       []string{"DolRune", "UmRune", "BerRune", "IstRune"},
			BaseItemTypes: []string{"Armor"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      true,
		},
		{
			Name:          "Fortitude",
			Inserts:       []string{"ElRune", "SolRune", "DolRune", "LoRune"},
			BaseItemTypes: []string{"Armor"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      true,
		},
		{
			Name:          "Call to Arms",
			Inserts:       []string{"AmnRune", "RalRune", "MalRune", "IstRune", "OhmRune"},
			BaseItemTypes: []string{"Mace"},
			BaseSortOrder: []stat.ID{},
			AllowEth:      true,
		},
		{
			Name:          "Heart of the Oak",
			Inserts:       []string{"KoRune", "VexRune", "PulRune", "ThulRune"},
			BaseItemTypes: []string{"Mace"},
			BaseSortOrder: []stat.ID{},
			AllowEth:      true,
		},
		{
			Name:          "Flickering Flame",
			Inserts:       []string{"NefRune", "PulRune", "VexRune"},
			BaseItemTypes: []string{"Helm"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      false,
		},
		{
			Name:          "Smoke",
			Inserts:       []string{"NefRune", "LumRune"},
			BaseItemTypes: []string{"Armor"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      false,
		},
		{
			Name:          "Rhyme",
			Inserts:       []string{"ShaelRune", "EthRune"},
			BaseItemTypes: []string{"Shield"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      false,
		},
		{
			Name:          "Spirit",
			Inserts:       []string{"TalRune", "ThulRune", "OrtRune", "AmnRune"},
			BaseItemTypes: []string{"Sword", "Shield"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      false,
		},
		{
			Name:          "Insight",
			Inserts:       []string{"RalRune", "TirRune", "TalRune", "SolRune"},
			BaseItemTypes: []string{"Polearm"},
			BaseSortOrder: []stat.ID{stat.TwoHandedMaxDamage},
			AllowEth:      true,
		},
		{
			Name:          "Cure",
			Inserts:       []string{"ShaelRune", "IoRune", "TalRune"},
			BaseItemTypes: []string{"Helm"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      true,
		},
		{
			Name:          "Lore",
			Inserts:       []string{"OrtRune", "SolRune"},
			BaseItemTypes: []string{"Helm"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      false,
		},
		{
			Name:          "Ancients Pledge",
			Inserts:       []string{"RalRune", "OrtRune", "TalRune"},
			BaseItemTypes: []string{"Shield"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      false,
		},
		{
			Name:          "Leaf",
			Inserts:       []string{"TirRune", "RalRune"},
			BaseItemTypes: []string{"Staff"},
			BaseSortOrder: []stat.ID{},
			AllowEth:      true,
		},
		{
			Name:          "Stealth",
			Inserts:       []string{"TalRune", "EthRune"},
			BaseItemTypes: []string{"Armor"},
			BaseSortOrder: []stat.ID{stat.Defense},
			AllowEth:      false,
		},
	}
)

func MakeRunewords() error {
	ctx := context.Get()
	ctx.SetLastAction("SocketAddItems")

	insertItems := ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash, item.LocationInventory)
	baseItems := ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash, item.LocationInventory)

	for _, recipe := range RunewordRecipes {
		if !slices.Contains(ctx.CharacterCfg.Game.Leveling.EnabledRunewordRecipes, recipe.Name) {
			continue
		}

		ctx.Logger.Debug("Socket recipe is enabled, processing", "recipe", recipe.Name)

		continueProcessing := true
		for continueProcessing {

			if baseItem, hasBase := hasBaseForRunewordRecipe(baseItems, recipe); hasBase {

				existingTier, hasExisting := currentRunewordBaseTier(ctx, recipe, baseItem.Type().Name)
				// Prevent creating runeword multiple times if we don't care about damage / def
				if hasExisting && (len(recipe.BaseSortOrder) == 0 || baseItem.Desc().Tier() <= existingTier) {
					ctx.Logger.Debug("Skipping recipe - existing runeword has equal or better tier in same base type",
						"recipe", recipe.Name,
						"baseType", baseItem.Type().Name,
						"existingTier", existingTier,
						"newBaseTier", baseItem.Desc().Tier())
					continueProcessing = false
					continue
				}
				if inserts, hasInserts := hasItemsForRunewordRecipe(insertItems, recipe); hasInserts {
					err := SocketItems(ctx, recipe, baseItem, inserts...)
					if err != nil {
						return err
					}

					insertItems = removeUsedItems(insertItems, inserts)
				} else {
					continueProcessing = false
				}
				baseItems = removeUsedItems(baseItems, []data.Item{baseItem})
			} else {
				continueProcessing = false
			}
		}
	}
	return nil
}
func SocketItems(ctx *context.Status, recipe RunewordRecipe, base data.Item, items ...data.Item) error {

	ctx.SetLastAction("SocketItem")

	ins := ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash, item.LocationInventory)

	for _, itm := range items {
		if itm.Location.LocationType == item.LocationStash || itm.Location.LocationType == item.LocationSharedStash {
			OpenStash()
			break
		}
	}
	if !ctx.Data.OpenMenus.Stash && (base.Location.LocationType == item.LocationStash || base.Location.LocationType == item.LocationSharedStash) {
		err := OpenStash()
		if err != nil {
			return err
		}
	}

	if base.Location.LocationType == item.LocationSharedStash {
		ctx.Logger.Debug("Base in shared - checking it fits")
		if !itemFitsInventory(base) {
			ctx.Logger.Error("Base item does not fit in inventory", "item", base.Name)
			return step.CloseAllMenus()
		} else {
			ctx.Logger.Debug("Base in shared stash but fits in inv, switching to correct tab")
			SwitchStashTab(base.Location.Page + 1)
			ctx.Logger.Debug("Switched to correct tab")
			utils.Sleep(500)
			screenPos := ui.GetScreenCoordsForItem(base)
			ctx.Logger.Debug(fmt.Sprintf("Clicking after 5s at %d:%d", screenPos.X, screenPos.Y))
			ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
		}
	}

	requiredCounts := make(map[string]int)
	for _, insert := range recipe.Inserts {
		requiredCounts[insert]++
	}

	usedItems := make(map[*data.Item]bool)
	orderedItems := make([]data.Item, 0)

	// Process each required insert in order
	for _, requiredInsert := range recipe.Inserts {
		for i := range ins {
			item := &ins[i]
			if string(item.Name) == requiredInsert && !usedItems[item] {
				orderedItems = append(orderedItems, *item)
				usedItems[item] = true
				break
			}
		}
	}
	previousPage := -1 // Initialize to invalid page number
	for _, itm := range orderedItems {
		if itm.Location.LocationType == item.LocationSharedStash || itm.Location.LocationType == item.LocationStash {
			currentPage := itm.Location.Page + 1
			if previousPage != currentPage || currentPage != base.Location.Page {
				SwitchStashTab(currentPage)
			}
			previousPage = currentPage
		}

		screenPos := ui.GetScreenCoordsForItem(itm)
		ctx.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
		utils.Sleep(300)

		for _, movedBase := range ctx.Data.Inventory.AllItems {
			if base.UnitID == movedBase.UnitID {
				if (base.Location.LocationType == item.LocationStash) && base.Location.Page != itm.Location.Page {
					SwitchStashTab(base.Location.Page + 1)
				}

				basescreenPos := ui.GetScreenCoordsForItem(movedBase)
				ctx.HID.Click(game.LeftButton, basescreenPos.X, basescreenPos.Y)
				utils.Sleep(300)
				if itm.Location.LocationType == item.LocationCursor {
					DropMouseItem()
					return fmt.Errorf("failed to insert item %s into base %s", itm.Name, base.Name)
				}
			}
		}
		utils.Sleep(300)
	}
	return step.CloseAllMenus()
}

func currentRunewordBaseTier(ctx *context.Status, recipe RunewordRecipe, baseType string) (item.Tier, bool) {

	items := ctx.Data.Inventory.ByLocation(
		item.LocationInventory,
		item.LocationEquipped,
		item.LocationStash,
		item.LocationSharedStash,
	)

	for _, itm := range items {
		if string(itm.RunewordName) == recipe.Name && itm.Type().Name == baseType {
			return itm.Desc().Tier(), true
		}
	}
	return 0, false
}

func hasBaseForRunewordRecipe(items []data.Item, rwrecipe RunewordRecipe) (data.Item, bool) {
	var validBases []data.Item
	for _, itm := range items {
		itemType := itm.Type().Name

		isValidType := false
		for _, baseType := range rwrecipe.BaseItemTypes {
			if itemType == baseType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			continue
		}

		sockets, found := itm.FindStat(stat.NumSockets, 0)
		if !found || sockets.Value != len(rwrecipe.Inserts) {
			continue
		}

		if itm.Ethereal && !rwrecipe.AllowEth {
			continue
		}

		if itm.HasSocketedItems() {
			continue
		}

		if itm.Quality > item.QualitySuperior {
			continue
		}

		validBases = append(validBases, itm)
	}

	if len(validBases) == 0 {
		return data.Item{}, false
	}

	// Sort bases by BaseSortOrder if provided
	if len(rwrecipe.BaseSortOrder) > 0 {
		slices.SortFunc(validBases, func(a, b data.Item) int {
			for _, statID := range rwrecipe.BaseSortOrder {
				statA, foundA := a.FindStat(statID, 0)
				statB, foundB := b.FindStat(statID, 0)

				if !foundA && !foundB {
					continue
				}
				if !foundA {
					return -1
				}
				if !foundB {
					return 1
				}

				if statA.Value != statB.Value {
					return statB.Value - statA.Value // Higher values first
				}
			}
			return 0
		})
	} else {
		// When no BaseSortOrder specified, sort by lowest combined requirements
		slices.SortFunc(validBases, func(a, b data.Item) int {
			aStr := a.Desc().RequiredStrength
			aDex := a.Desc().RequiredDexterity
			bStr := b.Desc().RequiredStrength
			bDex := b.Desc().RequiredDexterity

			aTotal := aStr + aDex
			bTotal := bStr + bDex

			return aTotal - bTotal // Lower requirements first
		})
	}

	return validBases[0], true
}

func hasItemsForRunewordRecipe(items []data.Item, rwrecipe RunewordRecipe) ([]data.Item, bool) {

	RunewordRecipeItems := make(map[string]int)
	for _, item := range rwrecipe.Inserts {
		RunewordRecipeItems[item]++
	}

	itemsForRecipe := []data.Item{}

	for _, item := range items {
		if count, ok := RunewordRecipeItems[string(item.Name)]; ok {

			itemsForRecipe = append(itemsForRecipe, item)

			// Check if we now have exactly the needed count before decrementing
			count -= 1
			if count == 0 {
				delete(RunewordRecipeItems, string(item.Name))
				if len(RunewordRecipeItems) == 0 {
					return itemsForRecipe, true
				}
			} else {
				RunewordRecipeItems[string(item.Name)] = count
			}
		}
	}

	return nil, false
}
