package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"strings"
)

func (b Builder) ItemPickup() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		itemsToPickup := b.getItemsToPickup(data)
		for _, item := range itemsToPickup {
			b.logger.Debug(fmt.Sprintf("Item Detected: %s [%s] at X:%d Y:%d", item.Name, item.Quality, item.Position.X, item.Position.Y))
			steps = append(steps, step.PickupItem(b.logger, item))
		}

		return
	}, CanBeSkipped())
}

func (b Builder) getItemsToPickup(data game.Data) []game.Item {
	missingHealingPotions := b.bm.GetMissingCount(data, game.HealingPotion)
	missingManaPotions := b.bm.GetMissingCount(data, game.ManaPotion)
	missingRejuvenationPotions := b.bm.GetMissingCount(data, game.RejuvenationPotion)
	var itemsToPickup []game.Item
	for _, item := range data.Items.Ground {
		for _, pickitItem := range config.Pickit.Items {
			if strings.EqualFold(item.Name, pickitItem.Name) {
				// Pickup potions only if they are required
				if item.IsHealingPotion() {
					if missingHealingPotions == 0 {
						break
					}
					itemsToPickup = append(itemsToPickup, item)
					missingHealingPotions--
					break
				}
				if item.IsManaPotion() {
					if missingManaPotions == 0 {
						break
					}
					itemsToPickup = append(itemsToPickup, item)
					missingManaPotions--
					break
				}
				if item.IsRejuvPotion() {
					if missingRejuvenationPotions == 0 {
						break
					}
					itemsToPickup = append(itemsToPickup, item)
					missingRejuvenationPotions--
					break
				}

				if b.shouldBePickedUp(data, item, pickitItem) {
					itemsToPickup = append(itemsToPickup, item)
					break
				}
			}

			// Check if we should pickup gold, based on amount
			if config.Pickit.PickupGold && strings.EqualFold(item.Name, "Gold") {
				if item.Stats[game.StatGold] >= config.Pickit.MinimumGoldToPickup && data.PlayerUnit.Stats[game.StatGold] < data.PlayerUnit.MaxGold() {
					itemsToPickup = append(itemsToPickup, item)
					break
				}
			}
		}
	}

	return itemsToPickup
}

func (b Builder) shouldBePickedUp(d game.Data, i game.Item, pi config.ItemPickit) bool {
	if pi.Quality != "" && !strings.EqualFold(string(i.Quality), pi.Quality) {
		return false
	}

	if pi.Ethereal != nil && i.Ethereal != *pi.Ethereal {
		return false
	}

	// Exclude Gheed if we are already have one
	if i.Name == game.ItemGrandCharm && i.Quality == game.ItemQualityUnique {
		for _, invItem := range d.Items.Inventory {
			if invItem.Name == game.ItemGrandCharm && invItem.Quality == game.ItemQualityUnique {
				b.logger.Warn("Gheed's Fortune dropped, but you already have one in the inventory, skipping.")
				return false
			}
		}
	}

	return true
}
