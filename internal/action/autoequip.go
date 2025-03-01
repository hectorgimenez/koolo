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

	if target == item.LocationMercenary {
		if slices.Contains(mercBodyLocs, bodyLoc) {
			if bodyLoc == item.LocLeftArm {
				// Merc can only use spears, polearms and javelins in left arm
				return i.Desc().Type == "spea" ||
					i.Desc().Type == "pole" ||
					i.Desc().Type == "jave"
			}
			return true
		}
		return false
	}

	// Player validation
	if target == item.LocationEquipped {
		isShield := slices.Contains(shieldTypes, i.Desc().Type)

		// Shields can only go in right arm
		if isShield {
			return bodyLoc == item.LocRightArm
		}

		// For non-shield items, all locations are valid except right arm if it's reserved for shields
		if bodyLoc == item.LocRightArm {
			return !isShield
		}

		return true
	}

	return false
}

// evaluateItems processes items for either player or merc
func evaluateItems(items []data.Item, target item.LocationType, scoreFunc func(data.Item) float64) map[item.LocationType][]data.Item {
	itemsByLoc := make(map[item.LocationType][]data.Item)
	for _, itm := range items {

		if !isEquippable(itm, target) {
			continue
		}

		locations := itm.Desc().GetType().BodyLocs
		for _, loc := range locations {
			if isValidLocation(itm, loc, target) {
				itemsByLoc[loc] = append(itemsByLoc[loc], itm)
			}
		}
	}

	// Sort items by score in each location
	for loc := range itemsByLoc {
		sort.Slice(itemsByLoc[loc], func(i, j int) bool {
			scoreI := scoreFunc(itemsByLoc[loc][i])
			scoreJ := scoreFunc(itemsByLoc[loc][j])
			return scoreI > scoreJ
		})
	}

	// Special handling for two-handed weapons
	// If a two-handed weapon is equipped, check if a one-handed weapon + shield combo scores better
	if target == item.LocationEquipped {
		var bestComboScore float64
		if items, ok := itemsByLoc[item.LocLeftArm]; ok {
			// Check if the highest scoring left arm item is two-handed
			if _, found := items[0].FindStat(stat.TwoHandedMinDamage, 0); found {
				// Find best non-two-handed weapon score
				for _, item := range items {
					if _, isTwoHanded := item.FindStat(stat.TwoHandedMinDamage, 0); !isTwoHanded {
						bestComboScore = scoreFunc(item)
						break
					}
				}
				// Add best shield score if available
				if rightArmItems, ok := itemsByLoc[item.LocRightArm]; ok {
					bestComboScore += scoreFunc(rightArmItems[0])
				}

				// If one-hand + shield combo scores better, remove the two-handed weapon
				if bestComboScore >= scoreFunc(items[0]) {
					itemsByLoc[item.LocLeftArm] = itemsByLoc[item.LocLeftArm][1:]
				}
			}
		}
	}

	return itemsByLoc
}

// equipBestItems equips the highest scoring items for each location
func equipBestItems(itemsByLoc map[item.LocationType][]data.Item, target item.LocationType) error {
	ctx := context.Get()

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

			// Skip if item is equipped by the other target (player/merc)
			if (itm.Location.LocationType == item.LocationMercenary && target == item.LocationEquipped) || (itm.Location.LocationType == item.LocationEquipped && target == item.LocationMercenary) {
				continue
			}

			if err := equip(itm, loc, target); err != nil {
				ctx.Logger.Error(fmt.Sprintf("Failed to equip %s: %v", itm.Name, err))
				continue
			}
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
		// We need to de-equip the item in the right ring slot first to prevent having to move cursor and click
		if bodyloc == item.LocRightRing {
			if itemFitsInventory(itm) {
				equippedRing := data.Position{ui.EquipRRinX, ui.EquipRRinY}
				if ctx.Data.LegacyGraphics {
					equippedRing = data.Position{ui.EquipRRinClassicX, ui.EquipRRinClassicY}
				}
				ctx.HID.ClickWithModifier(game.LeftButton, equippedRing.X, equippedRing.Y, game.ShiftKey)
				utils.Sleep(EquipDelayMS)
			}
		}
		ctx.HID.ClickWithModifier(game.LeftButton, itemCoords.X, itemCoords.Y, game.ShiftKey)
	}

	utils.Sleep(100)
	ctx.RefreshGameData()
	for _, inPlace := range ctx.Data.Inventory.AllItems {
		if itm.UnitID == inPlace.UnitID && inPlace.Location.LocationType != target {
			step.CloseAllMenus()
			ctx.Logger.Error(fmt.Sprintf("Failed to equip %s to %s using hotkeys, trying cursor", itm.Name, target))
			// Temporarily disabled using cursor to equip - it isn't needed for any case now and it will be removed after testing
			//return equipCursor(itm, bodyloc, target)
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

			return fmt.Errorf("Failed %s to %s equip to using cursor", itm.Name, target)
		}
	}

	step.CloseAllMenus()
	return nil
}
