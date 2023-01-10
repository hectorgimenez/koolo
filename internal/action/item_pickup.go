package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"strings"
	"time"
)

func (b Builder) ItemPickup(waitForDrop bool) *DynamicAction {
	firstCallTime := time.Time{}
	return BuildDynamic(func(data game.Data) ([]step.Step, bool) {
		if firstCallTime.IsZero() {
			firstCallTime = time.Now()
		}

		itemsToPickup := b.getItemsToPickup(data)
		if len(itemsToPickup) > 0 {
			i := itemsToPickup[0]
			b.logger.Debug(fmt.Sprintf(
				"Item Detected: %s [%d] at X:%d Y:%d",
				i.Name,
				i.Quality,
				i.Position.X,
				i.Position.Y,
			))

			return []step.Step{step.PickupItem(b.logger, i)}, true
		}

		// Add small delay, drop is not instant
		if waitForDrop && time.Since(firstCallTime) < time.Second {
			b.logger.Debug("Waiting a second to check again for drop")
			return []step.Step{
				step.SyncStep(func(data game.Data) error {
					helper.Sleep(int((time.Second - time.Since(firstCallTime)).Milliseconds()))
					return nil
				}),
			}, true
		}

		return nil, false
	}, CanBeSkipped())
}

func (b Builder) getItemsToPickup(data game.Data) []game.Item {
	missingHealingPotions := b.bm.GetMissingCount(data, game.HealingPotion)
	missingManaPotions := b.bm.GetMissingCount(data, game.ManaPotion)
	missingRejuvenationPotions := b.bm.GetMissingCount(data, game.RejuvenationPotion)
	var itemsToPickup []game.Item
	for _, itm := range data.Items.Ground {
		// Pickup potions only if they are required
		if itm.IsHealingPotion() {
			if missingHealingPotions == 0 {
				continue
			}
			if b.shouldBePickedUp(data, itm) {
				itemsToPickup = append(itemsToPickup, itm)
				missingHealingPotions--
			}
			continue
		}
		if itm.IsManaPotion() {
			if missingManaPotions == 0 {
				continue
			}
			if b.shouldBePickedUp(data, itm) {
				itemsToPickup = append(itemsToPickup, itm)
				missingManaPotions--
			}
			continue
		}
		if itm.IsRejuvPotion() {
			if missingRejuvenationPotions == 0 {
				continue
			}
			if b.shouldBePickedUp(data, itm) {
				itemsToPickup = append(itemsToPickup, itm)
				missingRejuvenationPotions--
			}
			continue
		}

		if b.shouldBePickedUp(data, itm) {
			itemsToPickup = append(itemsToPickup, itm)
			continue
		}
	}

	return itemsToPickup
}

func (b Builder) shouldBePickedUp(d game.Data, i game.Item) bool {
	// Check if we should pickup gold, based on amount
	if config.Pickit.PickupGold && strings.EqualFold(string(i.Name), "Gold") {
		if i.Stats[stat.Gold] >= config.Pickit.MinimumGoldToPickup && d.PlayerUnit.Stats[stat.Gold] < d.PlayerUnit.MaxGold() {
			return true
		}
	}

	return i.PickupPass(false)
}
