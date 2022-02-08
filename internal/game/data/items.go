package data

const (
	ItemScrollTownPortal   = "Scroll of Town Portal"
	ItemSuperHealingPotion = "Super Healing Potion"
	ItemSuperManaPotion    = "Super Mana Potion"

	ItemQualityNormal   Quality = "NORMAL"
	ItemQualitySuperior Quality = "SUPERIOR"
	ItemQualityMagic    Quality = "MAGIC"
	ItemQualitySet      Quality = "SET"
	ItemQualityRare     Quality = "RARE"
	ItemQualityUNIQUE   Quality = "UNIQUE"
)

type Quality string

type Items struct {
	Belt      Belt
	Inventory Inventory
	Shop      []Item
	Ground    []Item
}

type Inventory struct {
	Items []Item
}

type Item struct {
	Name     string
	Quality  Quality
	Position Position
	Ethereal bool
}

func (i Inventory) ShouldBuyTPs() bool {
	return true
}
