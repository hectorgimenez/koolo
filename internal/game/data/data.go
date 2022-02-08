package data

const (
	// Towns
	AreaRogueEncampment Area = "RogueEncampment"
	AreaLutGholein      Area = "LutGholein"
	AreaKurastDocks     Area = "KurastDocks"
	AreaPandemonium     Area = "ThePandemoniumFortress"
	AreaHarrogath       Area = "Harrogath"

	AreaNihlathaksTemple Area = "NihlathaksTemple"

	// Classes
	ClassSorceress Class = "Sorceress"
)

type DataRepository interface {
	GameData() Data
}

type Data struct {
	Status        Status
	Area          Area
	AreaOrigin    Position
	Corpse        Corpse
	Monsters      map[NPCID]Monster
	CollisionGrid [][]int
	PlayerUnit    PlayerUnit
	NPCs          map[NPCID]NPC
	Items         Items
	Objects       []Object
	OpenMenus     OpenMenus
}

type Area string

func (a Area) IsTown() bool {
	switch a {
	case AreaRogueEncampment, AreaLutGholein, AreaKurastDocks, AreaPandemonium, AreaHarrogath:
		return true
	}

	return false
}

type Class string
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
	Stats     map[string]int
	Class     Class
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
