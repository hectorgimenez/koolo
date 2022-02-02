package data

const (
	ItemScrollTownPortal   = "Scroll of Town Portal"
	ItemSuperHealingPotion = "Super Healing Potion"
	ItemSuperManaPotion    = "Super Mana Potion"
)

type Items struct {
	Belt      Belt
	Inventory Inventory
	Shop      []Item
}

type Inventory struct {
	Items []Item
}

type Item struct {
	Name     string
	Position Position
}

func (i Inventory) ShouldBuyTPs() bool {
	return true
}
