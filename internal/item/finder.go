package item

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"go.uber.org/zap"
	"strings"
)

type Pickup struct {
	logger    *zap.Logger
	bm        health.BeltManager
	pf        helper.PathFinder
	pickitCfg config.Pickit
}

func NewPickup(logger *zap.Logger, bm health.BeltManager, pf helper.PathFinder, pickitCfg config.Pickit) Pickup {
	return Pickup{
		logger:    logger,
		bm:        bm,
		pf:        pf,
		pickitCfg: pickitCfg,
	}
}

func (f Pickup) Pickup() {
	itemsToPickup := f.getItemsToPickup()
	for _, item := range itemsToPickup {
		f.logger.Debug(fmt.Sprintf("Picking %s [%s] at X: %d Y: %d", item.Name, item.Quality, item.Position.X, item.Position.Y))
		if err := f.pf.PickupItem(item); err != nil {
			f.logger.Error(fmt.Sprintf("Error picking up %s item! %s", item.Name, err.Error()))
		}
	}
}

func (f Pickup) getItemsToPickup() []data.Item {
	groundItems := data.Status.Items.Ground

	missingHealingPotions := f.bm.GetMissingCount(data.HealingPotion)
	missingManaPotions := f.bm.GetMissingCount(data.ManaPotion)
	missingRejuvenationPotions := f.bm.GetMissingCount(data.RejuvenationPotion)
	var itemsToPickup []data.Item
	for _, item := range groundItems {
		for _, pickitItem := range f.pickitCfg.Items {
			// Pickup potions only if they are required
			if strings.Contains(strings.ToLower(item.Name), "healingpotion") {
				if missingHealingPotions == 0 {
					break
				}
				itemsToPickup = append(itemsToPickup, item)
				missingHealingPotions--
				break
			}
			if strings.EqualFold(strings.ToLower(item.Name), "manapotion") {
				if missingManaPotions == 0 {
					break
				}
				itemsToPickup = append(itemsToPickup, item)
				missingManaPotions--
				break
			}
			if strings.EqualFold(strings.ToLower(item.Name), "rejuvenationpotion") {
				if missingRejuvenationPotions == 0 {
					break
				}
				itemsToPickup = append(itemsToPickup, item)
				missingRejuvenationPotions--
				break
			}

			if strings.EqualFold(item.Name, pickitItem.Name) {
				if pickitItem.Quality == "" || strings.EqualFold(string(item.Quality), pickitItem.Quality) {
					itemsToPickup = append(itemsToPickup, item)
					break
				}
			}

			// Check if we should pickup gold, based on amount
			if f.pickitCfg.PickupGold && strings.EqualFold(item.Name, "Gold") {
				if item.Stats[data.StatGold] >= f.pickitCfg.MinimumGoldToPickup {
					itemsToPickup = append(itemsToPickup, item)
					break
				}
			}
		}
	}

	return itemsToPickup
}
