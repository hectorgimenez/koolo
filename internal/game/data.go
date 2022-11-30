package game

import (
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/stat"
)

const (
	goldPerLevel = 10000

	// Classes
	ClassSorceress Class = "Sorceress"

	// Skills
	SkillBattleOrders Skill = "BattleOrders"

	// Monster Types
	MonsterTypeNone        MonsterType = "None"
	MonsterTypeChampion    MonsterType = "Champion"
	MonsterTypeMinion      MonsterType = "Minion"
	MonsterTypeUnique      MonsterType = "Unique"
	MonsterTypeSuperUnique MonsterType = "SuperUnique"
)

type Data struct {
	AreaOrigin       Position
	Corpse           Corpse
	Monsters         Monsters
	CollisionGrid    [][]bool
	PlayerUnit       PlayerUnit
	NPCs             NPCs
	Items            Items
	Objects          []Object
	AdjacentLevels   []Level
	PointsOfInterest []PointOfInterest
	OpenMenus        OpenMenus
}

func (d Data) MercHPPercent() int {
	for _, m := range d.Monsters {
		if m.IsMerc() {
			// Hacky thing to read merc life properly
			maxLife := m.Stats[stat.MaxLife] >> 8
			life := float64(m.Stats[stat.Life] >> 8)
			if m.Stats[stat.Life] <= 32768 {
				life = float64(m.Stats[stat.Life]) / 32768.0 * float64(maxLife)
			}

			return int(life / float64(maxLife) * 100)
		}
	}

	return 0
}

type Level struct {
	Area     area.Area
	Position Position
}

type Class string

type Corpse struct {
	Found     bool
	IsHovered bool
	Position  Position
}

type Position struct {
	X int
	Y int
}

type Skill string
type PlayerUnit struct {
	Name     string
	Area     area.Area
	Position Position
	Stats    map[stat.Stat]int
	Skills   map[Skill]int
}

func (pu PlayerUnit) MaxGold() int {
	return goldPerLevel * pu.Stats[stat.Level]
}

func (pu PlayerUnit) HPPercent() int {
	return int((float64(pu.Stats[stat.Life]) / float64(pu.Stats[stat.MaxLife])) * 100)
}

func (pu PlayerUnit) MPPercent() int {
	return int((float64(pu.Stats[stat.Mana]) / float64(pu.Stats[stat.MaxMana])) * 100)
}

type PointOfInterest struct {
	Name     string
	Position Position
}

type OpenMenus struct {
	Inventory     bool
	LoadingScreen bool
	NPCInteract   bool
	NPCShop       bool
	Stash         bool
	Waypoint      bool
}
