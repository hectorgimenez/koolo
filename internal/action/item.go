package action

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/itemfilter"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/game"
	"slices"
)

func (b *Builder) doesExceedQuantity(i data.Item, rule nip.Rule, stashItems []data.Item) bool {
	if len(rule.MaxQuantity) == 0 {
		return false
	}

	// For now, use this only for gems, runes, tokens, ubers. Add more items after testing
	allowedTypeGroups := []string{"runes", "ubers", "tokens", "chippedgems", "flawedgems", "gems", "flawlessgems", "perfectgems"}
	if !slices.Contains(allowedTypeGroups, i.Type()) {
		b.Logger.Debug(fmt.Sprintf("Skipping max quantity check for %s item", i.Name))
		return false
	}

	maxQuantity := 0

	for _, maxQuantityGroup := range rule.MaxQuantity {
		for _, maxQComparable := range maxQuantityGroup.Comparable {
			if maxQComparable.Keyword == "maxquantity" && maxQComparable.ValueInt > 0 {
				maxQuantity = maxQComparable.ValueInt
				break
			}
		}
	}

	if maxQuantity == 0 {
		b.Logger.Debug(fmt.Sprintf("Max quantity for %s item is 0, skipping further logic", i.Name))
		return false
	}

	matchedItemsInStash := 0

	for _, stashItem := range stashItems {
		_, found := itemfilter.Evaluate(stashItem, []nip.Rule{rule})
		if found {
			matchedItemsInStash += 1
		}
	}

	b.Logger.Debug(fmt.Sprintf("For item %s found %d max quantity from pickit rule, number of items in the stash tabs %d", i.Name, maxQuantity, matchedItemsInStash))

	return matchedItemsInStash >= maxQuantity
}

func (b *Builder) allStashItems(d game.Data) (S []data.Item) {
	return slices.Concat(
		d.Items.ByLocation(item.LocationStash),
		d.Items.ByLocation(item.LocationVendor),       // When stash is open, this returns all items in the three shared stash tabs
		d.Items.ByLocation(item.LocationSharedStash1), // Broken, always returns nil
		d.Items.ByLocation(item.LocationSharedStash2), // Broken, always returns nil
		d.Items.ByLocation(item.LocationSharedStash3), // Broken, always returns nil
	)
}
