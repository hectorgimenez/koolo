package game

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
	CainNPC     NPCID = "Deckard Cain"

	// Monsters
	Countess    NPCID = "The Countess"
	Andariel    NPCID = "Andariel"
	Pindleskin  NPCID = "Pindleskin"
	Mephisto    NPCID = "Mephisto"
	TheSummoner NPCID = "The Summoner"
	Nihlathak   NPCID = "Nihlathak"
)

type NPCID string
type Resist string

type Monsters []Monster

func (m Monsters) FindOne(npcid NPCID) (Monster, bool) {
	for _, monster := range m {
		if monster.Name == string(npcid) {
			return monster, true
		}
	}

	return Monster{}, false
}

func (m Monster) IsImmune(resist Resist) bool {
	for _, i := range m.Immunities {
		if i == resist {
			return true
		}
	}

	return false
}
