package game

import (
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/skill"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/game/state"
)

// since stat.MaxLife is returning max life without stats, we are setting the max life value that we read from the
// game memory, overwriting this value each time it increases. It's not a good solution but it will provide
// more accurate values for the life %. This value is checked for each memory iteration.
var maxLife = 0
var maxLifeBO = 0

const (
	goldPerLevel = 10000

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

type PlayerUnit struct {
	Name     string
	Area     area.Area
	Position Position
	Stats    map[stat.Stat]int
	Skills   map[skill.Skill]int
	States   state.States
}

func (pu PlayerUnit) MaxGold() int {
	return goldPerLevel * pu.Stats[stat.Level]
}

func (pu PlayerUnit) HPPercent() int {
	if maxLifeBO == 0 && maxLife == 0 {
		maxLife = pu.Stats[stat.MaxLife]
		maxLifeBO = pu.Stats[stat.MaxLife]
	}

	if pu.States.HasState(state.STATE_BATTLEORDERS) {
		if maxLifeBO < pu.Stats[stat.Life] {
			maxLifeBO = pu.Stats[stat.Life]
		}
		return int((float64(pu.Stats[stat.Life]) / float64(maxLifeBO)) * 100)
	}

	if maxLife < pu.Stats[stat.Life] {
		maxLife = pu.Stats[stat.Life]
	}

	return int((float64(pu.Stats[stat.Life]) / float64(maxLife)) * 100)
}

func (pu PlayerUnit) MPPercent() int {
	return int((float64(pu.Stats[stat.Mana]) / float64(pu.Stats[stat.MaxMana])) * 100)
}

func (pu PlayerUnit) HasDebuff() bool {
	debuffs := []state.State{
		state.STATE_AMPLIFYDAMAGE,
		state.STATE_ATTRACT,
		state.STATE_CONFUSE,
		state.STATE_CONVERSION,
		state.STATE_DECREPIFY,
		state.STATE_DIMVISION,
		state.STATE_IRONMAIDEN,
		state.STATE_LIFETAP,
		state.STATE_LOWERRESIST,
		state.STATE_TERROR,
		state.STATE_WEAKEN,
		state.STATE_CONVICTED,
		state.STATE_CONVICTION,
		state.STATE_POISON,
		state.STATE_COLD,
		state.STATE_SLOWED,
		state.STATE_BLOOD_MANA,
		state.STATE_DEFENSE_CURSE,
	}

	for _, s := range pu.States {
		for _, d := range debuffs {
			if s == d {
				return true
			}
		}
	}

	return false
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
	MapShown      bool
}
