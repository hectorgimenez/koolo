package inventory

type InventoryRepository interface {
	Inventory() Inventory
}

type Inventory struct {
	Belt Belt
}
