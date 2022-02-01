package data

type Items struct {
	Belt      Belt
	Inventory Inventory
}

type Inventory struct {
	Items []BaseItem
}

type BaseItem struct {
	Name     string
	Position Position
}

func (i Inventory) ShouldBuyTPs() bool {
	return true
}
