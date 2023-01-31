package game

import (
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/game/stat"
)

type NPC struct {
	ID        npc.ID
	Name      string
	Positions []Position
}

type MonsterType string

type Monster struct {
	UnitID
	Name      npc.ID
	IsHovered bool
	Position  Position
	Stats     map[stat.Stat]int
	Type      MonsterType
}

type Monsters []Monster
type NPCs []NPC

func (n NPCs) FindOne(npcid npc.ID) (NPC, bool) {
	for _, np := range n {
		if np.ID == npcid {
			return np, true
		}
	}

	return NPC{}, false
}

func (m Monsters) FindOne(id npc.ID, t MonsterType) (Monster, bool) {
	for _, monster := range m {
		if monster.Name == id {
			if t == MonsterTypeNone || t == monster.Type {
				return monster, true
			}
		}
	}

	return Monster{}, false
}

func (m Monsters) Enemies(filters ...MonsterFilter) []Monster {
	monsters := make([]Monster, 0)
	for _, mo := range m {
		if !mo.IsMerc() {
			monsters = append(monsters, mo)
		}
	}

	for _, f := range filters {
		monsters = f(m)
	}

	return monsters
}

type MonsterFilter func(m Monsters) []Monster

func MonsterEliteFilter() MonsterFilter {
	return func(m Monsters) []Monster {
		var filteredMonsters []Monster
		for _, mo := range m {
			if mo.Type == MonsterTypeMinion || mo.Type == MonsterTypeUnique || mo.Type == MonsterTypeChampion || mo.Type == MonsterTypeSuperUnique {
				filteredMonsters = append(filteredMonsters, mo)
			}
		}

		return filteredMonsters
	}
}

func MonsterAnyFilter() MonsterFilter {
	return func(m Monsters) []Monster {
		return m
	}
}

func (m Monsters) FindByID(id UnitID) (Monster, bool) {
	for _, monster := range m {
		if monster.UnitID == id {
			return monster, true
		}
	}

	return Monster{}, false
}

func (m Monster) IsImmune(resist stat.Resist) bool {
	for st, value := range m.Stats {
		// We only want max resistance
		if value < 100 {
			continue
		}
		if resist == stat.ColdImmune && st == stat.ColdResist {
			return true
		}
	}
	return false
}

func (m Monster) IsMerc() bool {
	if m.Name == npc.Guard {
		return true
	}

	return false
}
