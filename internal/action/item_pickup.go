package action

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"

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

	for {
		itemsToPickup := GetItemsToPickup(maxDistance)
		if len(itemsToPickup) == 0 {
			return nil
		}

		itemToPickup := data.Item{}
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

		// Clear enemy monsters near the item
		ClearAreaAroundPosition(itemToPickup.Position, 4, data.MonsterAnyFilter())

		ctx.Logger.Debug(fmt.Sprintf(
			"Item Detected: %s [%d] at X:%d Y:%d",
			itemToPickup.Name,
			itemToPickup.Quality,
			itemToPickup.Position.X,
			itemToPickup.Position.Y,
		))

		// First try pickup if we're already in range
		currentDistance := ctx.PathFinder.DistanceFromMe(itemToPickup.Position)
		if currentDistance <= 4 {
			err := step.PickupItem(itemToPickup)
			if err == nil {
				continue // Item picked up successfully, move to next item
			}

			// If we get a line of sight error even at close range, try moving beyond
			if errors.Is(err, step.ErrNoLOSToItem) {
				beyondTarget := moveBeyondItem(itemToPickup.Position, 2)
				if err := MoveToCoords(beyondTarget); err != nil {
					ctx.Logger.Warn("Failed moving beyond item, continuing...")
					continue
				}
				err = step.PickupItem(itemToPickup)
				if err == nil {
					continue
				}
			}
		} else {
			// Not in range, need to move closer first
			moveTarget := itemToPickup.Position
			if err := step.MoveTo(moveTarget, step.WithDistanceToFinish(3)); err != nil {
				ctx.Logger.Warn("Failed moving closer to item, trying to pickup anyway")
			}

			err := step.PickupItem(itemToPickup)
			if err == nil {
				continue // Item picked up successfully, move to next item
			}

			if errors.Is(err, step.ErrItemTooFar) {
				ctx.Logger.Debug("Item is too far away, retrying...")
				continue
			}

			if errors.Is(err, step.ErrNoLOSToItem) {
				ctx.Logger.Debug("No line of sight to item, moving further...")
				beyondTarget := moveBeyondItem(itemToPickup.Position, 2)
				if err := MoveToCoords(beyondTarget); err != nil {
					ctx.Logger.Warn("Failed moving beyond item, continuing...")
					continue
				}
				err = step.PickupItem(itemToPickup)
				if err == nil {
					continue
				}
			}
		}
		// One final attempt before blacklisting - maybe there was undetected monsters
		ClearAreaAroundPosition(itemToPickup.Position, 3, data.MonsterAnyFilter())
		err := step.PickupItem(itemToPickup)
		if err == nil {
			continue
		}
		// If we still can't pick up the item after all attempts, blacklist it
		ctx.CurrentGame.BlacklistedItems = append(ctx.CurrentGame.BlacklistedItems, itemToPickup)
		ctx.Logger.Warn(
			"Failed picking up item, blacklisting it",
			slog.String("itemName", itemToPickup.Desc().Name),
			slog.Int("unitID", int(itemToPickup.UnitID)),
		)
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
