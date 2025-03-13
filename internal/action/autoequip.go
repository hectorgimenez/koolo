package action

import (
	"fmt"
	"slices"
	"sort"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

// Constants for equipment locations
const (
	EquipDelayMS = 500
)

var (
	classItems = map[data.Class][]string{
		data.Amazon:      {"ajav", "abow", "aspe"},
		data.Sorceress:   {"orb"},
		data.Necromancer: {"head"},
		data.Paladin:     {"ashd"},
		data.Barbarian:   {"phlm"},
		data.Druid:       {"pelt"},
		data.Assassin:    {"h2h"},
	}

	// shieldTypes defines items that should be equipped in right arm (technically they can be left or right arm but we don't want to try and equip two shields)
	shieldTypes = []string{"shie", "ashd", "head"}

	// mercBodyLocs defines valid mercenary equipment locations
	// No support for A3 and A5 mercs
	mercBodyLocs = []item.LocationType{item.LocHead, item.LocTorso, item.LocLeftArm}

	// questItems defines items that shouldn't be equipped
	// TODO Fix IsFromQuest() and remove
	questItems = []item.Name{
		"StaffOfKings",
		"HoradricStaff",
		"AmuletOfTheViper",
		"KhalimsFlail",
	}
)

// EvaluateAllItems evaluates and equips items for both player and mercenary
func AutoEquip() error {
	ctx := context.Get()

	allItems := ctx.Data.Inventory.ByLocation(
		item.LocationStash,
		item.LocationSharedStash,
		item.LocationInventory,
		item.LocationEquipped,
		item.LocationMercenary,
	)

	// Process player items
	playerItems := evaluateItems(allItems, item.LocationEquipped, PlayerScore)
	if err := equipBestItems(playerItems, item.LocationEquipped); err != nil {
		return fmt.Errorf("failed to equip player items: %w", err)
	}

	// Process mercenary items
	mercItems := evaluateItems(allItems, item.LocationMercenary, MercScore)
	if ctx.Data.MercHPPercent() > 0 {
		if err := equipBestItems(mercItems, item.LocationMercenary); err != nil {
			return fmt.Errorf("failed to equip mercenary items: %w", err)
		}
	}
	return nil
}

// isEquippable checks if an item meets the requirements for the given unit (player or NPC)
func isEquippable(i data.Item, target item.LocationType) bool {
	ctx := context.Get()

	bodyLoc := i.Desc().GetType().BodyLocs
	// Check item has valid equipment locations
	if bodyLoc == nil {
		return false
	}

	var str, dex, lvl int
	// Get required stats
	if target == item.LocationEquipped {
		str = ctx.Data.PlayerUnit.Stats[stat.Strength].Value
		dex = ctx.Data.PlayerUnit.Stats[stat.Dexterity].Value
		lvl = ctx.Data.PlayerUnit.Stats[stat.Level].Value
	} else if target == item.LocationMercenary {
		for _, m := range ctx.Data.Monsters {
			if m.IsMerc() {
				str = m.Stats[stat.Strength]
				dex = m.Stats[stat.Dexterity]
				lvl = m.Stats[stat.Level]
			}
		}
	}

	// Check for quest items
	isQuestItem := slices.Contains(questItems, i.Name)

	// Check for class specific
	for class, items := range classItems {
		if ctx.Data.PlayerUnit.Class != class && slices.Contains(items, i.Desc().Type) {
			return false
		}
	}

	// Check requirements
	return i.Identified &&
		str >= i.Desc().RequiredStrength &&
		dex >= i.Desc().RequiredDexterity &&
		lvl >= i.LevelReq &&
		!isQuestItem
}

func isValidLocation(i data.Item, bodyLoc item.LocationType, target item.LocationType) bool {
	ctx := context.Get()
	class := ctx.Data.PlayerUnit.Class
	itemType := i.Desc().Type
	isShield := slices.Contains(shieldTypes, itemType)

	if target == item.LocationMercenary {
		if slices.Contains(mercBodyLocs, bodyLoc) {
			if bodyLoc == item.LocLeftArm {
				// Merc can only use spears, polearms and javelins in left arm
				return itemType == "spea" || itemType == "pole" || itemType == "jave"
			}
			return true
		}
		return false
	}

	// Player validation
	if target == item.LocationEquipped {
		// Shields can only go in right arm
		if isShield {
			return bodyLoc == item.LocRightArm
		}

		if bodyLoc != item.LocRightArm {
			return true
		}

		// Dual wield
		switch class {
		case data.Barbarian:
			// Barbs can dual wield 1h weaps and 2h swords
			_, isOneHanded := i.FindStat(stat.MaxDamage, 0)
			_, isTwoHanded := i.FindStat(stat.TwoHandedMaxDamage, 0)
			return isOneHanded || (isTwoHanded && itemType == "swor")

		case data.Assassin:
			isClaws := itemType == "h2h" || itemType == "h2h2"

			// Only allow claws in right arm if there are claws in left arm
			if isClaws && bodyLoc == item.LocRightArm {
				for _, equippedItem := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
					if equippedItem.Location.BodyLocation == item.LocLeftArm {
						return equippedItem.Desc().Type == "h2h" || equippedItem.Desc().Type == "h2h2"
					}
				}
				return false
			}
			return isClaws
		default:
			// Everyone else can only have shields in right arm
			return false
		}
	}

	return false
}

// evaluateItems processes items for either player or merc
func evaluateItems(items []data.Item, target item.LocationType, scoreFunc func(data.Item) map[item.LocationType]float64) map[item.LocationType][]data.Item {
	ctx := context.Get()
	itemsByLoc := make(map[item.LocationType][]data.Item)
	itemScores := make(map[data.UnitID]map[item.LocationType]float64)

	for _, itm := range items {

		if !isEquippable(itm, target) {
			continue
		}

		// Get scores for all possible body locations
		bodyLocScores := scoreFunc(itm)

		if len(bodyLocScores) > 0 {
			if _, exists := itemScores[itm.UnitID]; !exists {
				itemScores[itm.UnitID] = make(map[item.LocationType]float64)
			}

			for bodyLoc, score := range bodyLocScores {
				isValid := isValidLocation(itm, bodyLoc, target)

				if isValid {
					itemScores[itm.UnitID][bodyLoc] = score
					itemsByLoc[bodyLoc] = append(itemsByLoc[bodyLoc], itm)
				}
			}
		}
	}

	// Sort items by score in each location
	for loc := range itemsByLoc {
		sort.Slice(itemsByLoc[loc], func(i, j int) bool {
			scoreI := itemScores[itemsByLoc[loc][i].UnitID][loc]
			scoreJ := itemScores[itemsByLoc[loc][j].UnitID][loc]
			return scoreI > scoreJ
		})

		ctx.Logger.Debug(fmt.Sprintf("*** Sorted items for %s ***", loc))
		for i, itm := range itemsByLoc[loc] {
			score := itemScores[itm.UnitID][loc]
			ctx.Logger.Debug(fmt.Sprintf("%d. %s (Score: %.1f)", i+1, itm.IdentifiedName, score))
		}
		ctx.Logger.Debug("**********************************")
	}

	// Check 1h weapon + shield score vs 2h weapons
	if target == item.LocationEquipped {
		class := ctx.Data.PlayerUnit.Class

		if items, ok := itemsByLoc[item.LocLeftArm]; ok && len(items) > 0 {
			// Check if the highest scoring left arm item is two-handed
			if _, found := items[0].FindStat(stat.TwoHandedMinDamage, 0); found {
				if class == data.Barbarian && items[0].Desc().Type == "swor" {
					// Skip the removal check for Barbarians with 2h swords
				} else {
					// Find best non-two-handed weapon score
					var bestComboScore float64
					for _, itm := range items {
						if _, isTwoHanded := itm.FindStat(stat.TwoHandedMinDamage, 0); !isTwoHanded {
							if score, exists := itemScores[itm.UnitID]["left_arm"]; exists {
								ctx.Logger.Debug(fmt.Sprintf("Best one-handed weapon score: %.1f", score))
								bestComboScore = score
								break
							}
						}
					}

					// Add best shield score if available
					if rightArmItems, ok := itemsByLoc[item.LocRightArm]; ok && len(rightArmItems) > 0 {
						if score, exists := itemScores[rightArmItems[0].UnitID][item.LocRightArm]; exists {
							ctx.Logger.Debug(fmt.Sprintf("Best shield score: %.1f", score))
							bestComboScore += score
							ctx.Logger.Debug(fmt.Sprintf("Best one-hand + shield combo score: %.1f", bestComboScore))
						}
					}

					// If one-hand + shield combo scores better, remove the two-handed weapon
					if twoHandedScore, exists := itemScores[items[0].UnitID][item.LocLeftArm]; exists && bestComboScore >= twoHandedScore {
						ctx.Logger.Debug(fmt.Sprintf("Removing two-handed weapon: %s", items[0].Name))
						itemsByLoc[item.LocLeftArm] = itemsByLoc[item.LocLeftArm][1:]
					}
				}
			}
		}
	}

	return itemsByLoc
}

// equipBestItems equips the highest scoring items for each location
func equipBestItems(itemsByLoc map[item.LocationType][]data.Item, target item.LocationType) error {
	ctx := context.Get()

	equippedItems := make(map[data.UnitID]bool)

	for loc, items := range itemsByLoc {
		if len(items) == 0 {
			continue
		}

		// Try each item in sorted order until we find one that can be equipped
		for _, itm := range items {

			// Skip if item is already equipped in the target location
			if itm.Location.LocationType == target {
				break
			}

			if equippedItems[itm.UnitID] {
				ctx.Logger.Debug(fmt.Sprintf("Skipping %s for %s as it was already equipped elsewhere", itm.Name, loc))
				continue
			}

			// Skip if item is equipped by the other target (player/merc)
			if (itm.Location.LocationType == item.LocationMercenary && target == item.LocationEquipped) || (itm.Location.LocationType == item.LocationEquipped && target == item.LocationMercenary) {
				continue
			}

			if err := equip(itm, loc, target); err != nil {
				ctx.Logger.Error(fmt.Sprintf("Failed to equip %s: %v", itm.Name, err))
				continue
			}

			// Mark this item as equipped
			equippedItems[itm.UnitID] = true
			break
		}

	}

	return nil
}

// passing in bodyloc as a parameter cos rings have 2 locations
func equip(itm data.Item, bodyloc item.LocationType, target item.LocationType) error {

	ctx := context.Get()
	ctx.SetLastAction("Equip")

	// Get coordinates for item and equipment location
	itemCoords := ui.GetScreenCoordsForItem(itm)

	if itm.Location.LocationType == item.LocationStash || itm.Location.LocationType == item.LocationSharedStash {
		OpenStash()
		utils.Sleep(EquipDelayMS)
		// Check in which tab the item is and switch to it
		switch itm.Location.LocationType {
		case item.LocationStash:
			SwitchStashTab(1)
		case item.LocationSharedStash:
			SwitchStashTab(itm.Location.Page + 1)
		}

		// We can't equip merc directly from stash using hotkeys, need to put it in inventory first
		if target == item.LocationMercenary {

			if itemFitsInventory(itm) {
				// Move from stash to inv
				ctx.HID.ClickWithModifier(game.LeftButton, itemCoords.X, itemCoords.Y, game.CtrlKey)
				step.CloseAllMenus() // Close Stash
				utils.Sleep(EquipDelayMS)

				inInventory := false
				for _, updatedItem := range ctx.Data.Inventory.AllItems {
					if itm.UnitID == updatedItem.UnitID {
						itemCoords = ui.GetScreenCoordsForItem(updatedItem)
						inInventory = true
						break
					}
				}
				if !inInventory || !itemFitsInventory(itm) {
					return fmt.Errorf("item not found in inventory")
				}
			}
		}
	}
	for !ctx.Data.OpenMenus.Inventory {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
		utils.Sleep(EquipDelayMS)
	}

	if target == item.LocationMercenary {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MercenaryScreen)
		utils.Sleep(EquipDelayMS)
		ctx.HID.ClickWithModifier(game.LeftButton, itemCoords.X, itemCoords.Y, game.CtrlKey)
	}

	if target == item.LocationEquipped {
		// We need to de-equip the item in the right ring or right arm slot first to prevent having to move cursor and click
		switch bodyloc {
		case item.LocRightRing:
			if !itemFitsInventory(itm) {
				return fmt.Errorf("not enough inventory space to unequip %s", itm.Name)
			}
			equippedRing := data.Position{X: ui.EquipRRinX, Y: ui.EquipRRinY}
			if ctx.Data.LegacyGraphics {
				equippedRing = data.Position{X: ui.EquipRRinClassicX, Y: ui.EquipRRinClassicY}
			}
			ctx.HID.ClickWithModifier(game.LeftButton, equippedRing.X, equippedRing.Y, game.ShiftKey)
			utils.Sleep(EquipDelayMS)

		case item.LocRightArm:
			// Check if there's something already equipped in the right arm
			for _, equippedItem := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
				if equippedItem.Location.BodyLocation == item.LocRightArm {
					if !itemFitsInventory(itm) {
						return fmt.Errorf("not enough inventory space to unequip %s", itm.Name)
					}

					// It will straight swap if the item type is the same
					// But item type needs to be different from left arm - TODO add the logic for this
					equippedRightArm := data.Position{X: ui.EquipRArmX, Y: ui.EquipRArmY}
					if ctx.Data.LegacyGraphics {
						equippedRightArm = data.Position{X: ui.EquipRArmClassicX, Y: ui.EquipRArmClassicY}
					}
					ctx.HID.ClickWithModifier(game.LeftButton, equippedRightArm.X, equippedRightArm.Y, game.ShiftKey)
					utils.Sleep(EquipDelayMS)
					break // Only need to un-equip one item
				}
			}
		}
		ctx.Logger.Debug(fmt.Sprintf("Equipping %s at %v to %s using hotkeys", itm.Name, itemCoords, bodyloc))
		ctx.HID.ClickWithModifier(game.LeftButton, itemCoords.X, itemCoords.Y, game.ShiftKey)
	}

	utils.Sleep(100)
	ctx.RefreshGameData()
	utils.Sleep(500)
	for _, inPlace := range ctx.Data.Inventory.AllItems {
		if itm.UnitID == inPlace.UnitID && inPlace.Location.BodyLocation == bodyloc {
			step.CloseAllMenus()
			ctx.Logger.Error(fmt.Sprintf("Failed to equip %s to %s using hotkeys, trying cursor", itm.Name, target))

			return equipCursor(itm, bodyloc, target)
		}
	}

	step.CloseAllMenus()
	return nil

}

// Fallback for when hotkey equip doesn't work
func equipCursor(itm data.Item, bodyloc item.LocationType, target item.LocationType) error {
	ctx := context.Get()
	ctx.SetLastAction("EquipCursor")

	// Get coordinates for item and equipment location
	itemCoords := ui.GetScreenCoordsForItem(itm)
	bodyCoords := ui.GetEquipCoords(bodyloc, target)

	if itm.Location.LocationType == item.LocationStash || itm.Location.LocationType == item.LocationSharedStash {
		OpenStash()
		utils.Sleep(EquipDelayMS) // Add small delay to allow the game to open the inventory
		// Check in which tab the item is and switch to it
		switch itm.Location.LocationType {
		case item.LocationStash:
			SwitchStashTab(1)
		case item.LocationSharedStash:
			SwitchStashTab(itm.Location.Page + 1)
		}
	} else {
		for !ctx.Data.OpenMenus.Inventory {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
			utils.Sleep(EquipDelayMS) // Add small delay to allow the game to open the inventory
		}
	}

	context.Get().HID.Click(game.LeftButton, itemCoords.X, itemCoords.Y)
	utils.Sleep(EquipDelayMS)
	if target == item.LocationMercenary {
		step.CloseAllMenus()
		utils.Sleep(EquipDelayMS) // Add small delay to allow the game to open the inventory
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MercenaryScreen)
	}
	context.Get().HID.Click(game.LeftButton, bodyCoords.X, bodyCoords.Y)
	for _, inPlace := range ctx.Data.Inventory.AllItems {
		if itm.UnitID == inPlace.UnitID && inPlace.Location.LocationType != target {
			step.CloseAllMenus()
			utils.Sleep(EquipDelayMS)
			DropMouseItem()

			return fmt.Errorf("failed %s to %s equip to using cursor", itm.Name, target)
		}
	}

	step.CloseAllMenus()
	return nil
}
