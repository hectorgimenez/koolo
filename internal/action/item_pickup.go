package action

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
)

func itemFitsInventory(i data.Item) bool {
	invMatrix := context.Get().Data.Inventory.Matrix()

	for y := 0; y <= len(invMatrix)-i.Desc().InventoryHeight; y++ {
		for x := 0; x <= len(invMatrix[0])-i.Desc().InventoryWidth; x++ {
			freeSpace := true
			for dy := 0; dy < i.Desc().InventoryHeight; dy++ {
				for dx := 0; dx < i.Desc().InventoryWidth; dx++ {
					if invMatrix[y+dy][x+dx] {
						freeSpace = false
						break
					}
				}
				if !freeSpace {
					break
				}
			}

			if freeSpace {
				return true
			}
		}
	}

	return false
}

func ItemPickup(maxDistance int) error {
	ctx := context.Get()
	ctx.SetLastAction("ItemPickup")

	const maxRetries = 5
	const maxItemTooFarAttempts = 5

	for {
		ctx.PauseIfNotPriority()

		itemsToPickup := GetItemsToPickup(maxDistance)
		if len(itemsToPickup) == 0 {
			return nil
		}

		// Find first item that fits in inventory
		var itemToPickup data.Item
		for _, i := range itemsToPickup {
			if itemFitsInventory(i) {
				itemToPickup = i
				break
			}
		}

		if itemToPickup.UnitID == 0 {
			ctx.Logger.Debug("Inventory is full, returning to town to sell junk and stash items")
			InRunReturnTownRoutine()
			continue
		}

		ctx.Logger.Debug(fmt.Sprintf(
			"Item Detected: %s [%d] at X:%d Y:%d",
			itemToPickup.Name,
			itemToPickup.Quality,
			itemToPickup.Position.X,
			itemToPickup.Position.Y,
		))

		// Try to pick up the item with retries
		var lastError error
		attempt := 1
		attemptItemTooFar := 1
		for attempt <= maxRetries {
			// Clear monsters on each attempt
			ClearAreaAroundPosition(itemToPickup.Position, 4, data.MonsterAnyFilter())

			// Calculate position to move to based on attempt number
			// on 2nd and 3rd attempt try position left/right of item
			// on 4th and 5th attempt try position further away
			pickupPosition := itemToPickup.Position
			moveDistance := 3
			if attempt > 1 {
				switch attempt {
				case 2:
					pickupPosition = data.Position{
						X: itemToPickup.Position.X + moveDistance,
						Y: itemToPickup.Position.Y - 1,
					}
				case 3:
					pickupPosition = data.Position{
						X: itemToPickup.Position.X - moveDistance,
						Y: itemToPickup.Position.Y + 1,
					}
				case 4:
					pickupPosition = data.Position{
						X: itemToPickup.Position.X + moveDistance + 2,
						Y: itemToPickup.Position.Y - 3,
					}
				case 5:
					MoveToCoords(ctx.PathFinder.BeyondPosition(ctx.Data.PlayerUnit.Position, itemToPickup.Position, 4))
				}
			}

			distance := ctx.PathFinder.DistanceFromMe(itemToPickup.Position)
			if distance >= 7 || attempt > 1 {
				distanceToFinish := 3
				if attempt > 1 {
					distanceToFinish = 2
				}
				if err := step.MoveTo(pickupPosition, step.WithDistanceToFinish(distanceToFinish)); err != nil {
					ctx.Logger.Debug(fmt.Sprintf("Failed moving to item on attempt %d: %v", attempt, err))
					lastError = err
					attempt++
					continue
				}
			}

			// Try to pick up the item
			err := step.PickupItem(itemToPickup, attempt)
			if err == nil {
				break // Success!
			}

			lastError = err
			// Skip logging when casting moving error and don't count these specific errors as retry attempts
			if errors.Is(err, step.ErrCastingMoving) {
				continue
			}
			ctx.Logger.Debug(fmt.Sprintf("Pickup attempt %d failed: %v", attempt, err))

			// Don't count these specific errors as retry attempts
			if errors.Is(err, step.ErrMonsterAroundItem) {
				continue
			}

			// Item too far retry logic
			if errors.Is(err, step.ErrItemTooFar) {
				// Use default retries first, if we hit last attempt retry add random movement and continue until maxItemTooFarAttempts
				if attempt >= maxRetries && attemptItemTooFar <= maxItemTooFarAttempts {
					ctx.Logger.Debug(fmt.Sprintf("Item too far pickup attempt %d", attemptItemTooFar))
					attemptItemTooFar++
					ctx.PathFinder.RandomMovement()
					continue
				}
			}

			if errors.Is(err, step.ErrNoLOSToItem) {
				ctx.Logger.Debug("No line of sight to item, moving closer",
					slog.String("item", itemToPickup.Desc().Name))

				// Try moving beyond the item for better line of sight
				beyondPos := ctx.PathFinder.BeyondPosition(ctx.Data.PlayerUnit.Position, itemToPickup.Position, 2+attempt)
				if mvErr := MoveToCoords(beyondPos); mvErr == nil {
					err = step.PickupItem(itemToPickup, attempt)
					if err == nil {
						break
					}
					lastError = err
				} else {
					lastError = mvErr
				}
			}

			attempt++

		}

		// If all attempts failed, blacklist the item
		if attempt > maxRetries && lastError != nil {
			ctx.CurrentGame.BlacklistedItems = append(ctx.CurrentGame.BlacklistedItems, itemToPickup)

			// Screenshot with show items on
			ctx.HID.KeyDown(ctx.Data.KeyBindings.ShowItems)
			screenshot := ctx.GameReader.Screenshot()
			event.Send(event.ItemBlackListed(event.WithScreenshot(ctx.Name, fmt.Sprintf("Item %s [%s] BlackListed in Area:%s", itemToPickup.Name, itemToPickup.Quality.ToString(), ctx.Data.PlayerUnit.Area.Area().Name), screenshot), data.Drop{Item: itemToPickup}))
			ctx.HID.KeyUp(ctx.Data.KeyBindings.ShowItems)

			ctx.Logger.Warn(
				"Failed picking up item after all attempts, blacklisting it",
				slog.String("itemName", itemToPickup.Desc().Name),
				slog.Int("unitID", int(itemToPickup.UnitID)),
				slog.String("lastError", lastError.Error()),
			)
		}
	}
}
func GetItemsToPickup(maxDistance int) []data.Item {
	ctx := context.Get()
	ctx.SetLastAction("GetItemsToPickup")

	missingHealingPotions := ctx.BeltManager.GetMissingCount(data.HealingPotion)
	missingManaPotions := ctx.BeltManager.GetMissingCount(data.ManaPotion)
	missingRejuvenationPotions := ctx.BeltManager.GetMissingCount(data.RejuvenationPotion)

	var itemsToPickup []data.Item
	_, isLevelingChar := ctx.Char.(context.LevelingCharacter)

	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationGround) {
		// Skip itempickup on party leveling Maggot Lair, is too narrow and causes characters to get stuck
		if isLevelingChar && itm.Name != "StaffOfKings" && (ctx.Data.PlayerUnit.Area == area.MaggotLairLevel1 ||
			ctx.Data.PlayerUnit.Area == area.MaggotLairLevel2 ||
			ctx.Data.PlayerUnit.Area == area.MaggotLairLevel3 ||
			ctx.Data.PlayerUnit.Area == area.ArcaneSanctuary) {
			continue
		}

		// Skip potion pickup for Berserker Barb in Travincal if configured
		if ctx.CharacterCfg.Character.Class == "berserker" &&
			ctx.CharacterCfg.Character.BerserkerBarb.SkipPotionPickupInTravincal &&
			ctx.Data.PlayerUnit.Area == area.Travincal &&
			itm.IsPotion() {
			continue
		}

		// Skip items that are outside pickup radius, this is useful when clearing big areas to prevent
		// character going back to pickup potions all the time after using them
		itemDistance := ctx.PathFinder.DistanceFromMe(itm.Position)
		if maxDistance > 0 && itemDistance > maxDistance && itm.IsPotion() {
			continue
		}

		if itm.IsPotion() {
			if (itm.IsHealingPotion() && missingHealingPotions > 0) ||
				(itm.IsManaPotion() && missingManaPotions > 0) ||
				(itm.IsRejuvPotion() && missingRejuvenationPotions > 0) {
				if shouldBePickedUp(itm) {
					itemsToPickup = append(itemsToPickup, itm)
					switch {
					case itm.IsHealingPotion():
						missingHealingPotions--
					case itm.IsManaPotion():
						missingManaPotions--
					case itm.IsRejuvPotion():
						missingRejuvenationPotions--
					}
				}
			}
		} else if shouldBePickedUp(itm) {
			itemsToPickup = append(itemsToPickup, itm)
		}
	}

	// Remove blacklisted items from the list, we don't want to pick them up
	filteredItems := make([]data.Item, 0, len(itemsToPickup))
	for _, itm := range itemsToPickup {
		isBlacklisted := false
		for _, blacklistedItem := range ctx.CurrentGame.BlacklistedItems {
			if itm.UnitID == blacklistedItem.UnitID {
				isBlacklisted = true
				break
			}
		}
		if !isBlacklisted {
			filteredItems = append(filteredItems, itm)
		}
	}

	return filteredItems
}

func shouldBePickedUp(i data.Item) bool {
	ctx := context.Get()
	ctx.SetLastAction("shouldBePickedUp")

	// Always pickup Runewords and Wirt's Leg
	if i.IsRuneword || i.Name == "WirtsLeg" {
		return true
	}

	// Pick up quest items if we're in leveling or questing run
	specialRuns := slices.Contains(ctx.CharacterCfg.Game.Runs, "quests") || slices.Contains(ctx.CharacterCfg.Game.Runs, "leveling")
	if specialRuns {
		switch i.Name {
		case "ScrollOfInifuss", "LamEsensTome", "HoradricCube", "AmuletoftheViper", "StaffofKings", "HoradricStaff", "AJadeFigurine", "KhalimsEye", "KhalimsBrain", "KhalimsHeart", "KhalimsFlail":
			return true
		}
	}
	if i.ID == 552 { // Book of Skill doesnt work by name, so we find it by ID
		return true
	}
	// Skip picking up gold if we can not carry more
	gold, _ := ctx.Data.PlayerUnit.FindStat(stat.Gold, 0)
	if gold.Value >= ctx.Data.PlayerUnit.MaxGold() && i.Name == "Gold" {
		ctx.Logger.Debug("Skipping gold pickup, inventory full")
		return false
	}

	// Skip picking up gold, usually early game there are small amounts of gold in many places full of enemies, better
	// stay away of that
	_, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if isLevelingChar && ctx.Data.PlayerUnit.TotalPlayerGold() < 50000 && i.Name != "Gold" {
		return true
	}

	// Pickup all magic or superior items if total gold is low, filter will not pass and items will be sold to vendor
	minGoldPickupThreshold := ctx.CharacterCfg.Game.MinGoldPickupThreshold
	if ctx.Data.PlayerUnit.TotalPlayerGold() < minGoldPickupThreshold && i.Quality >= item.QualityMagic {
		return true
	}

	// Evaluate item based on NIP rules
	matchedRule, result := ctx.Data.CharacterCfg.Runtime.Rules.EvaluateAll(i)
	if result == nip.RuleResultNoMatch {
		return false
	}
	if result == nip.RuleResultPartial {
		return true
	}
	return !doesExceedQuantity(matchedRule)
}
