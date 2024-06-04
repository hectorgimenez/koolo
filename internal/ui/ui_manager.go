package ui

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Manager struct {
	gr *game.MemoryReader
}

func NewManager(gr *game.MemoryReader) *Manager {
	return &Manager{
		gr: gr,
	}
}

func (ui *Manager) GetScreenCoordsForItem(itm data.Item) data.Position {
	if ui.gr.LegacyGraphics() {
		return getScreenCoordsForItemClassic(itm)
	}

	return getScreenCoordsForItem(itm)
}

func getScreenCoordsForItem(itm data.Item) data.Position {
	switch itm.Location.LocationType {
	case item.LocationVendor, item.LocationStash, item.LocationSharedStash:
		x := topCornerVendorWindowX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
		y := topCornerVendorWindowY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

		return data.Position{X: x, Y: y}
	}

	x := inventoryTopLeftX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
	y := inventoryTopLeftY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

	return data.Position{X: x, Y: y}
}

func getScreenCoordsForItemClassic(itm data.Item) data.Position {
	switch itm.Location.LocationType {
	case item.LocationVendor, item.LocationStash, item.LocationSharedStash:
		x := topCornerVendorWindowXClassic + itm.Position.X*itemBoxSizeClassic + (itemBoxSizeClassic / 2)
		y := topCornerVendorWindowYClassic + itm.Position.Y*itemBoxSizeClassic + (itemBoxSizeClassic / 2)

		return data.Position{X: x, Y: y}
	}

	x := inventoryTopLeftXClassic + itm.Position.X*itemBoxSizeClassic + (itemBoxSizeClassic / 2)
	y := inventoryTopLeftYClassic + itm.Position.Y*itemBoxSizeClassic + (itemBoxSizeClassic / 2)

	return data.Position{X: x, Y: y}
}
