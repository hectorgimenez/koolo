package data

const (
	AreaRogueEncampment Area = "RogueEncampment"
	AreaLutGholein      Area = "LutGholein"
	AreaKurastDocks     Area = "KurastDocks"
	AreaPandemonium     Area = "ThePandemoniumFortress"
	AreaHarrogath       Area = "Harrogath"
)

type DataRepository interface {
	GameData() Data
}

type Area string
type Corpse struct {
	Found    bool
	Position Position
}
type Monster struct {
	Name     string
	Position Position
}

type Position struct {
	X int
	Y int
}

type PlayerUnit struct {
	Name     string
	Position Position
}

type NPC struct {
	Name      string
	Positions []Position
}
type Data struct {
	Area          Area
	AreaOrigin    Position
	Corpse        Corpse
	Inventory     Inventory
	Monsters      map[NPCID]Monster
	CollisionGrid [][]int
	PlayerUnit    PlayerUnit
	NPCs          map[NPCID]NPC
}
