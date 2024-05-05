package action

import (
	"fmt"
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) doesExceedQuantity(i data.Item, rule nip.Rule, stashItems []data.Item) bool {
	maxQuantity := rule.MaxQuantity()
	if maxQuantity == 0 {
		return false
	}

	// For now, use this only for gems, runes, tokens, ubers. Add more items after testing
	allowedTypeGroups := []string{"runes", "ubers", "tokens", "chippedgems", "flawedgems", "gems", "flawlessgems", "perfectgems"}
	if !slices.Contains(allowedTypeGroups, i.TypeAsString()) {
		b.Logger.Debug(fmt.Sprintf("Skipping max quantity check for %s item", i.Name))
		return false
	}

	if maxQuantity == 0 {
		b.Logger.Debug(fmt.Sprintf("Max quantity for %s item is 0, skipping further logic", i.Name))
		return false
	}

	matchedItemsInStash := 0

	for _, stashItem := range stashItems {
		res, _ := rule.Evaluate(stashItem)
		if res == nip.RuleResultFullMatch {
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
