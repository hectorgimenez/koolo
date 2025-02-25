package ui

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/context"
)

var (
	ClassicCoords = map[item.LocationType]data.Position{
		item.LocHead:      {X: EquipHeadClassicX, Y: EquipHeadClassicY},
		item.LocNeck:      {X: EquipNeckClassicX, Y: EquipNeckClassicY},
		item.LocLeftArm:   {X: EquipLArmClassicX, Y: EquipLArmClassicY},
		item.LocRightArm:  {X: EquipRArmClassicX, Y: EquipRArmClassicY},
		item.LocTorso:     {X: EquipTorsClassicX, Y: EquipTorsClassicY},
		item.LocBelt:      {X: EquipBeltClassicX, Y: EquipBeltClassicY},
		item.LocGloves:    {X: EquipGlovClassicX, Y: EquipGlovClassicY},
		item.LocFeet:      {X: EquipFeetClassicX, Y: EquipFeetClassicY},
		item.LocLeftRing:  {X: EquipLRinClassicX, Y: EquipLRinClassicY},
		item.LocRightRing: {X: EquipRRinClassicX, Y: EquipRRinClassicY},
	}

	ResurrectedCoords = map[item.LocationType]data.Position{
		item.LocHead:      {X: EquipHeadX, Y: EquipHeadY},
		item.LocNeck:      {X: EquipNeckX, Y: EquipNeckY},
		item.LocLeftArm:   {X: EquipLArmX, Y: EquipLArmY},
		item.LocRightArm:  {X: EquipRArmX, Y: EquipRArmY},
		item.LocTorso:     {X: EquipTorsX, Y: EquipTorsY},
		item.LocBelt:      {X: EquipBeltX, Y: EquipBeltY},
		item.LocGloves:    {X: EquipGlovX, Y: EquipGlovY},
		item.LocFeet:      {X: EquipFeetX, Y: EquipFeetY},
		item.LocLeftRing:  {X: EquipLRinX, Y: EquipLRinY},
		item.LocRightRing: {X: EquipRRinX, Y: EquipRRinY},
	}

	ClassicMercCoords = map[item.LocationType]data.Position{
		item.LocHead:    {X: EquipMercHeadClassicX, Y: EquipMercHeadClassicY},
		item.LocLeftArm: {X: EquipMercLArmClassicX, Y: EquipMercLArmClassicY},
		item.LocTorso:   {X: EquipMercTorsClassicX, Y: EquipMercTorsClassicY},
	}

	ResurrectedMercCoords = map[item.LocationType]data.Position{
		item.LocHead:    {X: EquipMercHeadX, Y: EquipMercHeadY},
		item.LocLeftArm: {X: EquipMercLArmX, Y: EquipMercLArmY},
		item.LocTorso:   {X: EquipMercTorsX, Y: EquipMercTorsY},
	}
)

func GetScreenCoordsForItem(itm data.Item) data.Position {
	ctx := context.Get()
	if ctx.GameReader.LegacyGraphics() {
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
	case item.LocationCube:
		x := topCornerCubeWindowX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
		y := topCornerCubeWindowY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

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
	case item.LocationCube:
		x := topCornerCubeWindowXClassic + itm.Position.X*itemBoxSizeClassic + (itemBoxSizeClassic / 2)
		y := topCornerCubeWindowYClassic + itm.Position.Y*itemBoxSizeClassic + (itemBoxSizeClassic / 2)

		return data.Position{X: x, Y: y}
	}

	x := inventoryTopLeftXClassic + itm.Position.X*itemBoxSizeClassic + (itemBoxSizeClassic / 2)
	y := inventoryTopLeftYClassic + itm.Position.Y*itemBoxSizeClassic + (itemBoxSizeClassic / 2)

	return data.Position{X: x, Y: y}
}

func GetEquipCoords(bodyLoc item.LocationType, target item.LocationType) data.Position {
	ctx := context.Get()
	if ctx.Data.LegacyGraphics {
		if target == item.LocationEquipped {
			coord := ClassicCoords[bodyLoc]
			return coord
		} else if target == item.LocationMercenary {
			coord := ClassicMercCoords[bodyLoc]
			return coord
		}
	} else {
		if target == item.LocationEquipped {
			coord := ResurrectedCoords[bodyLoc]
			return coord
		}
		if target == item.LocationMercenary {
			coord := ResurrectedMercCoords[bodyLoc]
			return coord
		}
	}
	return data.Position{}
}
