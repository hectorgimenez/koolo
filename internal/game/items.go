package game

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/item"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"strings"
)

const (
	ItemScrollOfTownPortal = "ScrollOfTownPortal"
	ItemScrollOfIdentify   = "ScrollOfIdentify"
	ItemTomeOfTownPortal   = "TomeOfTownPortal"
	ItemTomeOfIdentify     = "TomeOfIdentify"
	ItemSuperHealingPotion = "SuperHealingPotion"
	ItemSuperManaPotion    = "SuperManaPotion"
	ItemGrandCharm         = "GrandCharm"
)

type Items struct {
	Belt      Belt
	Inventory Inventory
	Shop      []Item
	Ground    []Item
	Equipped  []Item
}

type Inventory []Item
type UnitID int

type Item struct {
	UnitID
	Name       item.Name
	Quality    item.Quality
	Position   Position
	Ethereal   bool
	IsHovered  bool
	Stats      map[stat.Stat]int
	Identified bool
	IsVendor   bool
}

func (i Item) PickupPass(checkStats bool) bool {
	for _, ip := range config.Pickit.Items {
		if !strings.EqualFold(string(i.Name), ip.Name) {
			continue
		}

		// Check item Quality
		if len(ip.Quality) > 0 {
			found := false
			for _, q := range ip.Quality {
				if q == i.Quality {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Check number of sockets
		if len(ip.Sockets) > 0 {
			found := false
			for _, s := range ip.Sockets {
				if s == i.Stats[stat.NumSockets] {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if ip.Ethereal != nil && i.Ethereal != *ip.Ethereal {
			continue
		}

		// Skip checking stats, for example when item is not identified we don't have them
		//if !checkStats {
		//	return true
		//}

		// Check for item stats, socket number skipped, already checked properly
		for s, value := range i.Stats {
			for pickitStat, pickitValue := range ip.Stats {
				if pickitStat != "sockets" && strings.EqualFold(s.String(), pickitStat) {
					if value < pickitValue {
						continue
					}
				}
			}
		}

		return true
	}

	return false
}

func (i Item) IsPotion() bool {
	return i.IsHealingPotion() || i.IsManaPotion() || i.IsRejuvPotion()
}

func (i Item) IsHealingPotion() bool {
	return strings.Contains(string(i.Name), string(HealingPotion))
}

func (i Item) IsManaPotion() bool {
	return strings.Contains(string(i.Name), string(ManaPotion))
}
func (i Item) IsRejuvPotion() bool {
	return strings.Contains(string(i.Name), string(RejuvenationPotion))
}

func (i Inventory) ShouldBuyTPs() bool {
	for _, it := range i {
		if it.Name != ItemTomeOfTownPortal {
			continue
		}

		qty, found := it.Stats[stat.Quantity]

		if qty <= helper.RandRng(1, 3) || !found {
			return true
		}
	}
	return false
}

func (i Inventory) ShouldBuyIDs() bool {
	for _, it := range i {
		if it.Name != ItemTomeOfIdentify {
			continue
		}

		qty, found := it.Stats[stat.Quantity]
		if qty <= helper.RandRng(3, 7) || !found {
			return true
		}
	}
	return false
}

func (i Inventory) NonLockedItems() (items []Item) {
	for _, itm := range i {
		if config.Config.Inventory.InventoryLock[itm.Position.Y][itm.Position.X] == 1 {
			items = append(items, itm)
		}
	}

	return
}
