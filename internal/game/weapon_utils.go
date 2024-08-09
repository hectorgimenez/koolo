package game

import (
	"github.com/hectorgimenez/d2go/pkg/data"
)

//@TODO Remove this file when the d2go project is updated to manage throwable max quantities
type weaponUtilsInterface interface {
	GetThrowableMaxQuantity(itemToCheck data.Item) int
	IsItemThrowable(item data.Item) bool
}

type weaponUtils struct {
	weaponUtilsInterface
}

var WeaponUtils = weaponUtils{}

func (t weaponUtils) GetThrowableMaxQuantity(itemToCheck data.Item) int {
	switch itemToCheck.Name {
	case "Spiculum":
		return 30
	case "ShortSpear", "Glaive", "Simbilan":
		return 40
	case "Pilum", "GreatPilum":
		return 50
	case "Javelin", "WarJavelin":
		return 60
	case "GhostGlaive":
		return 75
	case "ThrowingSpear", "Harpoon", "BalrogSpear", "WingedHarpoon", "MaidenJavelin", "CeremonialJavelin", "MatriarchalJavelin":
		return 80
	case "StygianPilum":
		return 90
	case "HyperionJavelin":
		return 100
	case "ThrowingAxe", "BalancedAxe", "Francisca", "Hurlbat":
		return 130
	case "ThrowingKnife", "BalancedKnife", "BattleDart", "WarDart":
		return 160
	case "FlyingAxe", "WingedAxe":
		return 180
	case "FlyingKnife", "WingedKnife":
		return 200
	default:
		return 50
	}
}

func (t weaponUtils) IsItemThrowable(item data.Item) bool {
	switch item.Name {
	case "Spiculum", "ShortSpear", "Glaive", "Simbilan", "Pilum", "GreatPilum", "Javelin", "WarJavelin", "GhostGlaive", "ThrowingSpear", "Harpoon", "BalrogSpear", "WingedHarpoon", "MaidenJavelin", "CeremonialJavelin", "MatriarchalJavelin", "StygianPilum", "HyperionJavelin", "ThrowingAxe", "BalancedAxe", "Francisca", "Hurlbat", "ThrowingKnife", "BalancedKnife", "BattleDart", "WarDart", "FlyingAxe", "WingedAxe", "FlyingKnife", "WingedKnife":
		return true
	default:
		return false
	}
}
