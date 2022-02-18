package item

//import (
//	"fmt"
//	"github.com/hectorgimenez/koolo/internal/config"
//	"github.com/hectorgimenez/koolo/internal/game"
//	"github.com/hectorgimenez/koolo/internal/health"
//	"github.com/hectorgimenez/koolo/internal/helper"
//	"go.uber.org/zap"
//	"strings"
//)
//
//type Pickup struct {
//	logger    *zap.Logger
//	bm        health.BeltManager
//	pf        helper.PathFinder
//	pickitCfg config.Pickit
//}
//
//func NewPickup(logger *zap.Logger, bm health.BeltManager, pf helper.PathFinder, pickitCfg config.Pickit) Pickup {
//	return Pickup{
//		logger:    logger,
//		bm:        bm,
//		pf:        pf,
//		pickitCfg: pickitCfg,
//	}
//}
//
//func (f Pickup) Pickup() int {
//	itemsToPickup := f.getItemsToPickup()
//	itemsPickedUp := 0
//	for _, item := range itemsToPickup {
//		f.logger.Debug(fmt.Sprintf("Picking %s [%s] at X: %d Y: %d", item.Name, item.Quality, item.Position.X, item.Position.Y))
//		if err := f.pf.PickupItem(item); err != nil {
//			f.logger.Error(fmt.Sprintf("Error picking up %s item! %s", item.Name, err.Error()))
//		} else {
//			itemsPickedUp++
//		}
//	}
//
//	return itemsPickedUp
//}
//
//func (f Pickup) getItemsToPickup() []game.Item {
//	groundItems := game.Status().Items.Ground
//
//	missingHealingPotions := f.bm.GetMissingCount(game.HealingPotion)
//	missingManaPotions := f.bm.GetMissingCount(game.ManaPotion)
//	missingRejuvenationPotions := f.bm.GetMissingCount(game.RejuvenationPotion)
//	var itemsToPickup []game.Item
//	for _, item := range groundItems {
//		for _, pickitItem := range f.pickitCfg.Items {
//			if strings.EqualFold(item.Name, pickitItem.Name) {
//				// Pickup potions only if they are required
//				if strings.Contains(strings.ToLower(item.Name), "healingpotion") {
//					if missingHealingPotions == 0 {
//						break
//					}
//					itemsToPickup = append(itemsToPickup, item)
//					missingHealingPotions--
//					break
//				}
//				if strings.Contains(strings.ToLower(item.Name), "manapotion") {
//					if missingManaPotions == 0 {
//						break
//					}
//					itemsToPickup = append(itemsToPickup, item)
//					missingManaPotions--
//					break
//				}
//				if strings.Contains(strings.ToLower(item.Name), "rejuvenationpotion") {
//					if missingRejuvenationPotions == 0 {
//						break
//					}
//					itemsToPickup = append(itemsToPickup, item)
//					missingRejuvenationPotions--
//					break
//				}
//
//				if pickitItem.Quality == "" || strings.EqualFold(string(item.Quality), pickitItem.Quality) {
//					itemsToPickup = append(itemsToPickup, item)
//					break
//				}
//			}
//
//			// Check if we should pickup gold, based on amount
//			if f.pickitCfg.PickupGold && strings.EqualFold(item.Name, "Gold") {
//				if item.Stats[game.StatGold] >= f.pickitCfg.MinimumGoldToPickup {
//					itemsToPickup = append(itemsToPickup, item)
//					break
//				}
//			}
//		}
//	}
//
//	return itemsToPickup
//}
