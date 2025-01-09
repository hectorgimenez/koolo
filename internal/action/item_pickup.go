package action

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
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

	const maxRetries = 3

	for {
		// Get all items that should be picked up
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

		// Track item pickup attempts
		successfulPickup := false
		attempt := 1

		// Keep trying until we hit max retries
		for attempt <= maxRetries {
			// Calculate target position based on attempt number
			// Slightly to the Right second attempt,  Left last attempt
			pickupPosition := itemToPickup.Position
			if attempt > 1 {
				switch attempt {
				case 2:
					pickupPosition.X += 3
					pickupPosition.Y -= 1
				case 3:
					pickupPosition.X -= 3
					pickupPosition.Y += 1
				}

				ctx.Logger.Debug(fmt.Sprintf(
					"Using different offset position for attempt %d: X:%d Y:%d",
					attempt,
					pickupPosition.X,
					pickupPosition.Y,
				))
			}

			// Try to move to the calculated position
			if err := step.MoveTo(pickupPosition, step.WithDistanceToFinish(2)); err != nil {
				ctx.Logger.Warn(fmt.Sprintf(
					"Failed moving to position: %v",
					err,
				))
				// If we can't move, count this as an attempt
				attempt++
				continue
			}

			// Check distance before making an attempt
			currentDistance := ctx.PathFinder.DistanceFromMe(itemToPickup.Position)
			if currentDistance > 4 {
				ctx.Logger.Debug(fmt.Sprintf(
					"Item too far (distance: %d), adjusting position",
					currentDistance,
				))
				// Don't count being too far as an attempt, just try a new position
				continue
			}

			ctx.Logger.Debug(fmt.Sprintf(
				"Attempting to pickup item (try %d/%d): %s [%d] at X:%d Y:%d",
				attempt,
				maxRetries,
				itemToPickup.Name,
				itemToPickup.Quality,
				itemToPickup.Position.X,
				itemToPickup.Position.Y,
			))

			// Try the pickup
			err := step.PickupItem(itemToPickup)
			if err == nil {
				successfulPickup = true
				break
			}

			ctx.Logger.Debug(fmt.Sprintf(
				"Pickup attempt %d failed: %v",
				attempt,
				err,
			))

			// Handle different error cases
			if errors.Is(err, step.ErrMonsterAroundItem) {
				// Clear monsters and don't count as an attempt
				ClearAreaAroundPosition(itemToPickup.Position, 4, data.MonsterAnyFilter())
				continue
			}

			// Handle line of sight issues
			if errors.Is(err, step.ErrNoLOSToItem) {
				beyondTarget := moveBeyondItem(itemToPickup.Position, 2+(attempt-1))
				if err := MoveToCoords(beyondTarget); err != nil {
					ctx.Logger.Warn(fmt.Sprintf("Failed moving to get line of sight on attempt  %d: %v", attempt, err))
					attempt++
					continue
				}

				err = step.PickupItem(itemToPickup)
				if err == nil {
					successfulPickup = true
					break
				}

				if errors.Is(err, step.ErrMonsterAroundItem) {
					ClearAreaAroundPosition(itemToPickup.Position, 4, data.MonsterAnyFilter())
					continue
				}

				ctx.Logger.Debug(fmt.Sprintf(
					"Beyond position pickup attempt %d failed: %v",
					attempt,
					err,
				))
			}

			// This was a real attempt, increment counter and add delay
			attempt++
			if attempt <= maxRetries {
				delay := 150 * time.Duration(attempt-1) * time.Millisecond
				time.Sleep(delay)
			}
		}

		// If all retries failed, blacklist the item
		if !successfulPickup {
			ctx.CurrentGame.BlacklistedItems = append(ctx.CurrentGame.BlacklistedItems, itemToPickup)
			ctx.Logger.Warn(
				"Failed picking up item after all attempts, blacklisting it",
				slog.String("itemName", itemToPickup.Desc().Name),
				slog.Int("unitID", int(itemToPickup.UnitID)),
				slog.Int("position_x", itemToPickup.Position.X),
				slog.Int("position_y", itemToPickup.Position.Y),
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
		if isLevelingChar && !itm.IsFromQuest() && (ctx.Data.PlayerUnit.Area == area.MaggotLairLevel1 ||
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
		case "Scrollofinifuss", "LamEsensTome", "HoradricCube", "AmuletoftheViper", "StaffofKings", "HoradricStaff", "AJadeFigurine", "KhalimsEye", "KhalimsBrain", "KhalimsHeart", "KhalimsFlail":
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

// TODO refactor this since its similar to the one in attack.go(ensureenemyisinrange) and put in move package.
func moveBeyondItem(itemPos data.Position, distance int) data.Position {
	ctx := context.Get()
	playerPos := ctx.Data.PlayerUnit.Position

	// Calculate direction vector
	dx := float64(itemPos.X - playerPos.X)
	dy := float64(itemPos.Y - playerPos.Y)

	// Normalize
	length := math.Sqrt(dx*dx + dy*dy)
	if length == 0 {
		return itemPos
	}

	dx = dx / length
	dy = dy / length

	// Extend beyond item position
	return data.Position{
		X: itemPos.X + int(dx*float64(distance)),
		Y: itemPos.Y + int(dy*float64(distance)),
	}
}
