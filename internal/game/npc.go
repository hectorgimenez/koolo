package game

import (
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"strings"
)

type Resist string

type NPC struct {
	Name      string
	Positions []Position
}

type MonsterType string

type Monster struct {
	Name      npc.ID
	IsHovered bool
	Position  Position
	Stats     map[stat.Stat]int
	Type      MonsterType
}

type Monsters []Monster
type NPCs []NPC

const (
	ColdImmune Resist = "ColdImmune"
	FireImmune Resist = "FireImmune"
)

func (n NPCs) FindOne(npcid npc.ID) (NPC, bool) {
	for _, npc := range n {
		if strings.EqualFold(npc.Name, string(npcid)) {
			return npc, true
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

func (m Monster) IsImmune(resist Resist) bool {
	for st, value := range m.Stats {
		// We only want max resistance
		if value < 100 {
			continue
		}
		if resist == ColdImmune && st == stat.ColdResist {
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
