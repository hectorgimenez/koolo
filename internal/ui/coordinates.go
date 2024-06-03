package ui

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
)

const (
	itemBoxSize        = 33
	itemBoxSizeClassic = 35

	inventoryTopLeftX        = 846
	inventoryTopLeftXClassic = 663

	inventoryTopLeftY        = 369
	inventoryTopLeftYClassic = 379

	topCornerVendorWindowX        = 109
	topCornerVendorWindowXClassic = 275

	topCornerVendorWindowY        = 147
	topCornerVendorWindowYClassic = 149

	MercAvatarPositionX        = 36
	MercAvatarPositionXClassic = 208

	MercAvatarPositionY        = 39
	MercAvatarPositionYClassic = 53

	CubeTransmuteBtnX        = 273
	CubeTransmuteBtnXClassic = 451

	CubeTransmuteBtnY        = 411
	CubeTransmuteBtnYClassic = 405

	CubeTakeItemX        = 306
	CubeTakeItemXClassic = 484

	CubeTakeItemY        = 365
	CubeTakeItemYClassic = 358

	WpTabStartX        = 131
	WpTabStartXClassic = 266

	WpTabStartY        = 148
	WpTabStartYClassic = 95

	WpTabSizeX        = 57
	WpTabSizeXClassic = 75

	WpListPositionX        = 200
	WpListPositionXClassic = 357

	WpListStartY        = 158
	WpListStartYClassic = 145

	WpAreaBtnHeight        = 41
	WpAreaBtnHeightClassic = 43

	AnvilCenterX = 272
	AnvilCenterY = 333
	AnvilBtnX    = 272
	AnvilBtnY    = 450

	MainSkillButtonX = 596
	MainSkillButtonY = 693

	SecondarySkillButtonX = 686
	SecondarySkillButtonY = 693

	GambleRefreshButtonX        = 390
	GambleRefreshButtonXClassic = 540

	GambleRefreshButtonY        = 515
	GambleRefreshButtonYClassic = 553

	SecondarySkillListFirstSkillX = 687
	MainSkillListFirstSkillX      = 592
	SkillListFirstSkillY          = 590
	SkillListSkillOffset          = 45

	FirstMercFromContractorListX = 175
	FirstMercFromContractorListY = 142

	StashGoldBtnX        = 966
	StashGoldBtnXClassic = 754

	StashGoldBtnY        = 526
	StashGoldBtnYClassic = 552

	StashGoldBtnConfirmX        = 547
	StashGoldBtnConfirmXClassic = 579

	StashGoldBtnConfirmY        = 388
	StashGoldBtnConfirmYClassic = 423

	SwitchStashTabBtnX        = 107
	SwitchStashTabBtnXClassic = 258

	SwitchStashTabBtnY        = 128
	SwitchStashTabBtnYClassic = 84

	SwitchStashTabBtnTabSize        = 82
	SwitchStashTabBtnTabSizeClassic = 96
)

func GetScreenCoordsForItem(itm data.Item) data.Position {
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

func GetScreenCoordsForItemClassic(itm data.Item) data.Position {
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
