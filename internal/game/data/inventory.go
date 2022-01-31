package data

type InventoryRepository interface {
	Inventory() Inventory
}

type Inventory struct {
	Belt Belt
}

func (i Inventory) ShouldBuyTPs() bool {
	return true
}
