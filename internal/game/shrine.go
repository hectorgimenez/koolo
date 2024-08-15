package game

import (
	"github.com/hectorgimenez/d2go/pkg/data"
)

type Shrine struct {
	ID         int
	TypeID     int
	ShrineType string
	Position   data.Position
	DesiredFor []string
	Desired    bool
}

type ShrineType struct {
	Refill          Shrine
	Health          Shrine
	Mana            Shrine
	HPXChange       Shrine
	ManaXChange     Shrine
	Armor           Shrine
	Combat          Shrine
	ResistFire      Shrine
	ResistCold      Shrine
	ResistLightning Shrine
	ResistPoison    Shrine
	Skill           Shrine
	ManaRegen       Shrine
	Stamina         Shrine
	Experience      Shrine
	UnknownShrine   Shrine
	Portal          Shrine
	Gem             Shrine
	Fire            Shrine
	Monster         Shrine
	Explosive       Shrine
	Poison          Shrine
}

func Shrines() ShrineType {
	return ShrineType{
		Refill:          Shrine{ID: 1, ShrineType: "Refill Shrine", DesiredFor: []string{"hp", "mp"}},
		Health:          Shrine{ID: 2, ShrineType: "Health Shrine", DesiredFor: []string{"hp"}},
		Mana:            Shrine{ID: 3, ShrineType: "Mana Shrine", DesiredFor: []string{"mp"}},
		HPXChange:       Shrine{ID: 4, ShrineType: "HPXChange Shrine", Desired: false},
		ManaXChange:     Shrine{ID: 5, ShrineType: "ManaXChange Shrine", Desired: false},
		Armor:           Shrine{ID: 6, ShrineType: "Armor Shrine"},
		Combat:          Shrine{ID: 7, ShrineType: "Combat Shrine", DesiredFor: []string{"combat"}},
		ResistFire:      Shrine{ID: 8, ShrineType: "Resist Fire Shrine"},
		ResistCold:      Shrine{ID: 9, ShrineType: "Resist Cold Shrine"},
		ResistLightning: Shrine{ID: 10, ShrineType: "Resist Lightning  Shrine"},
		ResistPoison:    Shrine{ID: 11, ShrineType: "Resist Poison Shrine"},
		Skill:           Shrine{ID: 12, ShrineType: "Skill Shrine"},
		ManaRegen:       Shrine{ID: 13, ShrineType: "Mana Regeneration Shrine", DesiredFor: []string{"mp"}},
		Stamina:         Shrine{ID: 14, ShrineType: "Stamina Shrine", DesiredFor: []string{"stamina"}},
		Experience:      Shrine{ID: 15, ShrineType: "Experience Shrine"},
		UnknownShrine:   Shrine{ID: 16, ShrineType: "Unknown Shrine"},
		Portal:          Shrine{ID: 17, ShrineType: "Portal Shrine", Desired: false},
		Gem:             Shrine{ID: 18, ShrineType: "Gem Shrine", DesiredFor: []string{"gems"}},
		Fire:            Shrine{ID: 19, ShrineType: "Fire Shrine", Desired: false},
		Monster:         Shrine{ID: 20, ShrineType: "Monster Shrine", DesiredFor: []string{"elites"}},
		Explosive:       Shrine{ID: 21, ShrineType: "Explosive Shrine", Desired: false},
		Poison:          Shrine{ID: 22, ShrineType: "Poison Shrine", Desired: false},
	}
}

func GetShrines(objects []data.Object, areaOrigin data.Position) []Shrine {
	shrinesList := []Shrine{}
	shrineTypes := Shrines()

	// Create a map for quick lookup of DesiredFor by ID
	shrineMap := map[int]Shrine{
		shrineTypes.Refill.ID:          shrineTypes.Refill,
		shrineTypes.Health.ID:          shrineTypes.Health,
		shrineTypes.Mana.ID:            shrineTypes.Mana,
		shrineTypes.HPXChange.ID:       shrineTypes.HPXChange,
		shrineTypes.ManaXChange.ID:     shrineTypes.ManaXChange,
		shrineTypes.Armor.ID:           shrineTypes.Armor,
		shrineTypes.Combat.ID:          shrineTypes.Combat,
		shrineTypes.ResistFire.ID:      shrineTypes.ResistFire,
		shrineTypes.ResistCold.ID:      shrineTypes.ResistCold,
		shrineTypes.ResistLightning.ID: shrineTypes.ResistLightning,
		shrineTypes.ResistPoison.ID:    shrineTypes.ResistPoison,
		shrineTypes.Skill.ID:           shrineTypes.Skill,
		shrineTypes.ManaRegen.ID:       shrineTypes.ManaRegen,
		shrineTypes.Stamina.ID:         shrineTypes.Stamina,
		shrineTypes.Experience.ID:      shrineTypes.Experience,
		shrineTypes.UnknownShrine.ID:   shrineTypes.UnknownShrine,
		shrineTypes.Portal.ID:          shrineTypes.Portal,
		shrineTypes.Gem.ID:             shrineTypes.Gem,
		shrineTypes.Fire.ID:            shrineTypes.Fire,
		shrineTypes.Monster.ID:         shrineTypes.Monster,
		shrineTypes.Explosive.ID:       shrineTypes.Explosive,
		shrineTypes.Poison.ID:          shrineTypes.Poison,
	}

	for _, obj := range objects {
		if obj.IsShrine() {
			shrine, exists := shrineMap[int(obj.Name)]
			if exists {
				thisShrine := Shrine{
					ID:         int(obj.ID),
					TypeID:     int(obj.Name),
					ShrineType: shrine.ShrineType,
					Position:   obj.Position,
					DesiredFor: shrine.DesiredFor,
					Desired:    shrine.Desired,
				}
				shrinesList = append(shrinesList, thisShrine)
			}
		}
	}
	return shrinesList
}
