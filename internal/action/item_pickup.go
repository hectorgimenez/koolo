package action

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b *Builder) ItemPickup(waitForDrop bool, maxDistance int) *Chain {
	firstCallTime := time.Time{}
	var itemBeingPickedUp data.UnitID

	return NewChain(func(d game.Data) []Action {
		if firstCallTime.IsZero() {
			firstCallTime = time.Now()
		}

		itemsToPickup := b.getItemsToPickup(d, maxDistance)
		if len(itemsToPickup) > 0 {
			for _, m := range d.Monsters.Enemies() {
				if dist := pather.DistanceFromMe(d, m.Position); dist < 7 {
					b.Logger.Debug("Aborting item pickup, monster nearby", slog.Any("monster", m))
					itemBeingPickedUp = -1
					return []Action{b.ch.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
						return m.UnitID, true
					}, nil)}
				}
			}

			i := itemsToPickup[0]

			// Error picking up Item, go back to town, sell junk, stash and try again.
			if itemBeingPickedUp == i.UnitID {
				b.Logger.Debug("Item could not be picked up, going back to town to sell junk and stash")
				return []Action{NewChain(func(d game.Data) []Action {
					itemBeingPickedUp = -1
					return b.InRunReturnTownRoutine()
				})}
			}

			b.Logger.Debug(fmt.Sprintf(
				"Item Detected: %s [%d] at X:%d Y:%d",
				i.Name,
				i.Quality,
				i.Position.X,
				i.Position.Y,
			))

			itemBeingPickedUp = i.UnitID
			return []Action{
				b.MoveToCoords(i.Position),
				NewStepChain(func(d game.Data) []step.Step {
					return []step.Step{step.PickupItem(b.Logger, i)}
				}, IgnoreErrors()),
			}
		}

		// Add small delay, drop is not instant
		if waitForDrop && time.Since(firstCallTime) < time.Second {
			msToWait := time.Second - time.Since(firstCallTime)
			b.Logger.Debug("No items detected, waiting a bit and will try again", slog.Int("waitMs", int(msToWait.Milliseconds())))

			return []Action{b.Wait(msToWait)}
		}

		return nil
	}, RepeatUntilNoSteps())
}

func (b *Builder) getItemsToPickup(d game.Data, maxDistance int) []data.Item {
	missingHealingPotions := b.bm.GetMissingCount(d, data.HealingPotion)
	missingManaPotions := b.bm.GetMissingCount(d, data.ManaPotion)
	missingRejuvenationPotions := b.bm.GetMissingCount(d, data.RejuvenationPotion)
	var itemsToPickup []data.Item
	_, isLevelingChar := b.ch.(LevelingCharacter)
	for _, itm := range d.Inventory.ByLocation(item.LocationGround) {
		// Skip itempickup on party leveling Maggot Lair, is too narrow and causes characters to get stuck
		if d.CharacterCfg.Companion.Enabled && isLevelingChar && !itm.IsFromQuest() && (d.PlayerUnit.Area == area.MaggotLairLevel1 || d.PlayerUnit.Area == area.MaggotLairLevel2 || d.PlayerUnit.Area == area.MaggotLairLevel3 || d.PlayerUnit.Area == area.ArcaneSanctuary) {
			continue
		}

		// Skip items that are outside pickup radius, this is useful when clearing big areas to prevent
		// character going back to pickup potions all the time after using them
		itemDistance := pather.DistanceFromMe(d, itm.Position)
		if maxDistance > 0 && itemDistance > maxDistance {
			continue
		}

		if !b.shouldBePickedUp(d, itm) {
			continue
		}

		// Pickup potions only if they are required
		if itm.IsHealingPotion() && missingHealingPotions > 0 {
			itemsToPickup = append(itemsToPickup, itm)
			missingHealingPotions--
			continue
		}
		if itm.IsManaPotion() && missingManaPotions > 0 {
			itemsToPickup = append(itemsToPickup, itm)
			missingManaPotions--
			continue
		}
		if itm.IsRejuvPotion() && missingRejuvenationPotions > 0 {
			itemsToPickup = append(itemsToPickup, itm)
			missingRejuvenationPotions--
			continue
		}

		if !itm.IsPotion() {
			itemsToPickup = append(itemsToPickup, itm)
		}
	}

	return itemsToPickup
}

func (b *Builder) shouldBePickedUp(d game.Data, i data.Item) bool {
	// Skip picking up gold if we can not carry more
	gold, _ := d.PlayerUnit.FindStat(stat.Gold, 0)
	if gold.Value >= d.PlayerUnit.MaxGold() {
		b.Logger.Debug("Skipping gold pickup, inventory full")
		return false
	}

	// Always pickup WirtsLeg!
	if i.Name == "WirtsLeg" {
		return true
	}

	// Only during leveling if gold amount is low pickup items to sell as junk
	_, isLevelingChar := b.ch.(LevelingCharacter)

	// Skip picking up gold, usually early game there are small amounts of gold in many places full of enemies, better
	// stay away of that
	if isLevelingChar && d.PlayerUnit.TotalPlayerGold() < 50000 && i.Name != "Gold" {
		return true
	}

	minGoldPickupThreshold := b.Container.CharacterCfg.Game.MinGoldPickupThreshold
	// Pickup all magic or superior items if total gold is low, filter will not pass and items will be sold to vendor
	if d.PlayerUnit.TotalPlayerGold() < minGoldPickupThreshold && i.Quality >= item.QualityMagic {
		return true
	}

	matchedRule, result := d.CharacterCfg.Runtime.Rules.EvaluateAll(i)
	if result == nip.RuleResultNoMatch {
		return false
	}

	if result == nip.RuleResultPartial {
		return true
	}

	exceedQuantity := b.doesExceedQuantity(i, matchedRule, d)

	return !exceedQuantity
}
