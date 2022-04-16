package game

import "strings"

const (
	// A1 Town NPCs
	AkaraNPC  NPCID = "Akara"
	KashyaNPC NPCID = "Kashya"
	CharsiNPC NPCID = "Charsi"

	// A2 Town NPCs
	DrognanNPC NPCID = "Drognan"
	FaraNPC    NPCID = "Fara"
	GreizNPC   NPCID = "Greiz"

	// A3 Town NPCs
	OrmusNPC   NPCID = "Ormus"
	AshearaNPC NPCID = "Asheara"
	HratliNPC  NPCID = "HratliNPC"

	// A5 Town NPCs
	MalahNPC    NPCID = "Malah"
	LarzukNPC   NPCID = "Larzuk"
	QualKehkNPC NPCID = "Qual-Kehk"
	CainNPC     NPCID = "DeckardCain"

	// Monsters
	Countess   NPCID = "The Countess"
	Andariel   NPCID = "Andariel"
	Pindleskin NPCID = "Pindleskin"
	Mephisto   NPCID = "Mephisto"
	Summoner   NPCID = "Summoner"
	Nihlathak  NPCID = "Nihlathak"
)

type NPCID string
type Resist string

type Monsters []Monster
type NPCs []NPC

func (n NPCs) FindOne(npcid NPCID) (NPC, bool) {
	for _, npc := range n {
		if strings.EqualFold(npc.Name, string(npcid)) {
			return npc, true
		}
	}

	return NPC{}, false
}

func (m Monsters) FindOne(npcid NPCID) (Monster, bool) {
	for _, monster := range m {
		if strings.EqualFold(monster.Name, string(npcid)) {
			return monster, true
		}
	}

	return Monster{}, false
}

func (m Monster) IsImmune(resist Resist) bool {
	for _, i := range m.Immunities {
		if strings.EqualFold(string(i), string(resist)) {
			return true
		}
	}

	return false
}
