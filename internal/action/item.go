package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) doesExceedQuantity(i data.Item, rule nip.Rule, d game.Data) bool {
	stashItems := d.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash)

	maxQuantity := rule.MaxQuantity()
	if maxQuantity == 0 {
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
