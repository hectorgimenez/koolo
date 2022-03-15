package game

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"math/rand"
	"strings"
	"time"
)

const (
	ItemScrollOfTownPortal = "ScrollOfTownPortal"
	ItemScrollOfIdentify   = "ScrollOfIdentify"
	ItemTomeOfTownPortal   = "TomeOfTownPortal"
	ItemTomeOfIdentify     = "TomeOfIdentify"
	ItemSuperHealingPotion = "SuperHealingPotion"
	ItemSuperManaPotion    = "SuperManaPotion"
	ItemGrandCharm         = "GrandCharm"

	ItemQualityNormal   Quality = "NORMAL"
	ItemQualitySuperior Quality = "SUPERIOR"
	ItemQualityMagic    Quality = "MAGIC"
	ItemQualitySet      Quality = "SET"
	ItemQualityRare     Quality = "RARE"
	ItemQualityUnique   Quality = "UNIQUE"

	StatQuantity       Stat = "Quantity"
	StatGold           Stat = "Gold"
	StatLevel          Stat = "Level"
	StatStashGold      Stat = "StashGold"
	StatDurability     Stat = "Durability"
	StatMaxDurability  Stat = "MaxDurability"
	StatNumSockets     Stat = "NumSockets"
	StatFasterCastRate Stat = "FasterCastRate"
	StatEnhancedDamage Stat = "EnhancedDamage"
	StatDefense        Stat = "Defense"
)

type Stat string
type Quality string

type Items struct {
	Belt      Belt
	Inventory Inventory
	Shop      []Item
	Ground    []Item
}

type Inventory []Item

type Item struct {
	ID         int
	Name       string
	Quality    Quality
	Position   Position
	Ethereal   bool
	IsHovered  bool
	Stats      map[Stat]int
	Identified bool
}

func (i Item) PickupPass(checkStats bool) bool {
	for _, ip := range config.Pickit.Items {
		if !strings.EqualFold(i.Name, ip.Name) {
			continue
		}

		// Check item Quality
		if len(ip.Quality) > 0 {
			found := false
			for _, q := range ip.Quality {
				if strings.EqualFold(q, string(i.Quality)) {
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
				if s == i.Stats[StatNumSockets] {
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
		if !checkStats {
			return true
		}

		// Check for item stats, socket number skipped, already checked properly
		for stat, value := range i.Stats {
			for pickitStat, pickitValue := range ip.Stats {
				if pickitStat != string(StatNumSockets) && strings.EqualFold(string(stat), pickitStat) {
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
	return strings.Contains(strings.ToLower(i.Name), "healingpotion")
}

func (i Item) IsManaPotion() bool {
	return strings.Contains(strings.ToLower(i.Name), "manapotion")
}
func (i Item) IsRejuvPotion() bool {
	return strings.Contains(strings.ToLower(i.Name), "rejuvenationpotion")
}

func (i Inventory) ShouldBuyTPs() bool {
	for _, it := range i {
		if it.Name != ItemTomeOfTownPortal {
			continue
		}

		qty, found := it.Stats[StatQuantity]
		rand.Seed(time.Now().UnixNano())
		if qty <= rand.Intn(5-1)+1 || !found {
			return true
		}
	}
	return false
}

func (i Inventory) ShouldBuyIDs() bool {
	for _, it := range i {
		if it.Name != ItemTomeOfTownPortal {
			continue
		}

		qty, found := it.Stats[StatQuantity]
		rand.Seed(time.Now().UnixNano())
		if qty <= rand.Intn(7-3)+1 || !found {
			return true
		}
	}
	return false
}

func (i Inventory) NonLockedItems() (items []Item) {
	for _, item := range i {
		if config.Config.Inventory.InventoryLock[item.Position.Y][item.Position.X] == 1 {
			items = append(items, item)
		}
	}

	return
}
