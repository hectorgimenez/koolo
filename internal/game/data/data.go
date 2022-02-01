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

type Data struct {
	Area          Area
	AreaOrigin    Position
	Corpse        Corpse
	Inventory     Inventory
	Monsters      map[NPCID]Monster
	CollisionGrid [][]int
	PlayerUnit    PlayerUnit
	NPCs          map[NPCID]NPC
	Items         Items
	OpenMenus     OpenMenus
}

type Area string
type Corpse struct {
	Found     bool
	IsHovered bool
	Position  Position
}
type Monster struct {
	Name      string
	IsHovered bool
	Position  Position
}

type Position struct {
	X int
	Y int
}

type PlayerUnit struct {
	Name      string
	IsHovered bool
	Position  Position
}

type NPC struct {
	Name      string
	Positions []Position
}

type OpenMenus struct {
	Inventory   bool
	NPCInteract bool
	NPCShop     bool
	Stash       bool
	Waypoint    bool
}
