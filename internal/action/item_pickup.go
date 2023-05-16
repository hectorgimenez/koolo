package action

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/itemfilter"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) ItemPickup(waitForDrop bool, maxDistance int) Action {
	firstCallTime := time.Time{}
	var itemBeingPickedUp data.UnitID
	var townCycle = map[string]bool{
		"town":   false,
		"id":     false,
		"vendor": false,
	}
	sellingJunk := false

	return NewFactory(func(d data.Data) Action {
		if sellingJunk {
			// No comments... only bad practices here.
			if !townCycle["id"] {
				townCycle["id"] = true
				return b.IdentifyAll(false)
			}
			if !townCycle["vendor"] {
				townCycle["vendor"] = true
				return b.VendorRefill()
			}

			return BuildStatic(func(d data.Data) []step.Step {
				for _, o := range d.Objects {
					if o.IsPortal() {
						return []step.Step{
							step.InteractObject(object.TownPortal, func(d data.Data) bool {
								if !d.PlayerUnit.Area.IsTown() {
									sellingJunk = false
									itemBeingPickedUp = -1
									return true
								}

								return false
							}),
						}
					}
				}

				return nil
			})
		}

		if firstCallTime.IsZero() {
			firstCallTime = time.Now()
		}

		itemsToPickup := b.getItemsToPickup(d, maxDistance)
		if len(itemsToPickup) > 0 {
			i := itemsToPickup[0]

			// Error picking up item, go back to town, sell junk, stash and try again.
			if itemBeingPickedUp == i.UnitID {
				b.logger.Debug("Item could not be picked up, going back to town to sell junk and stash")
				sellingJunk = true

				townCycle["town"] = true
				return b.ReturnTown()
			}

			b.logger.Debug(fmt.Sprintf(
				"Item Detected: %s [%d] at X:%d Y:%d",
				i.Name,
				i.Quality,
				i.Position.X,
				i.Position.Y,
			))

			return BuildStatic(func(d data.Data) []step.Step {
				itemBeingPickedUp = i.UnitID
				return []step.Step{step.PickupItem(b.logger, i)}
			}, IgnoreErrors())
		}

		// Add small delay, drop is not instant
		if waitForDrop && time.Since(firstCallTime) < time.Second {
			msToWait := int((time.Second - time.Since(firstCallTime)).Milliseconds())
			b.logger.Debug("No items detected, waiting a bit and will try again", zap.Int("waitMs", msToWait))

			return BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{step.SyncStep(func(d data.Data) error {
					helper.Sleep(msToWait)
					return nil
				})}
			})
		}

		return nil
	})
}

func (b Builder) getItemsToPickup(d data.Data, maxDistance int) []data.Item {
	missingHealingPotions := b.bm.GetMissingCount(d, data.HealingPotion)
	missingManaPotions := b.bm.GetMissingCount(d, data.ManaPotion)
	missingRejuvenationPotions := b.bm.GetMissingCount(d, data.RejuvenationPotion)
	var itemsToPickup []data.Item
	for _, itm := range d.Items.Ground {
		// Skip items that are outside pickup radius, this is useful when clearing big areas to prevent
		// character going back to pickup potions all the time after using them
		itemDistance := pather.DistanceFromMe(d, itm.Position)
		if maxDistance > 0 && itemDistance > maxDistance {
			continue
		}

		// Pickup potions only if they are required
		if itm.IsHealingPotion() {
			if missingHealingPotions == 0 {
				continue
			}
			if b.shouldBePickedUp(d, itm) {
				itemsToPickup = append(itemsToPickup, itm)
				missingHealingPotions--
			}
			continue
		}
		if itm.IsManaPotion() {
			if missingManaPotions == 0 {
				continue
			}
			if b.shouldBePickedUp(d, itm) {
				itemsToPickup = append(itemsToPickup, itm)
				missingManaPotions--
			}
			continue
		}
		if itm.IsRejuvPotion() {
			if missingRejuvenationPotions == 0 {
				continue
			}
			if b.shouldBePickedUp(d, itm) {
				itemsToPickup = append(itemsToPickup, itm)
				missingRejuvenationPotions--
			}
			continue
		}

		if b.shouldBePickedUp(d, itm) {
			itemsToPickup = append(itemsToPickup, itm)
			continue
		}
	}

	return itemsToPickup
}

func (b Builder) shouldBePickedUp(d data.Data, i data.Item) bool {
	// Only during leveling if gold amount is low pickup items to sell as junk
	_, isLevelingChar := b.ch.(LevelingCharacter)
	if isLevelingChar && d.PlayerUnit.TotalGold() < 10000 {
		return true
	}

	return itemfilter.Evaluate(i, config.Config.Runtime.Rules)
}
