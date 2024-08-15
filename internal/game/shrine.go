package game

type Shrine struct {
	ID         int
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
		Refill:          Shrine{ID: 1, DesiredFor: []string{"hp", "mp"}},
		Health:          Shrine{ID: 2, DesiredFor: []string{"hp"}},
		Mana:            Shrine{ID: 3, DesiredFor: []string{"mp"}},
		HPXChange:       Shrine{ID: 4, Desired: false},
		ManaXChange:     Shrine{ID: 5, Desired: false},
		Armor:           Shrine{ID: 6},
		Combat:          Shrine{ID: 7, DesiredFor: []string{"combat"}},
		ResistFire:      Shrine{ID: 8},
		ResistCold:      Shrine{ID: 9},
		ResistLightning: Shrine{ID: 10},
		ResistPoison:    Shrine{ID: 11},
		Skill:           Shrine{ID: 12},
		ManaRegen:       Shrine{ID: 13, DesiredFor: []string{"mp"}},
		Stamina:         Shrine{ID: 14, DesiredFor: []string{"stamina"}},
		Experience:      Shrine{ID: 15},
		UnknownShrine:   Shrine{ID: 16},
		Portal:          Shrine{ID: 17, Desired: false},
		Gem:             Shrine{ID: 18, DesiredFor: []string{"gems"}},
		Fire:            Shrine{ID: 19, Desired: false},
		Monster:         Shrine{ID: 20, DesiredFor: []string{"elites"}},
		Explosive:       Shrine{ID: 21, Desired: false},
		Poison:          Shrine{ID: 22, Desired: false},
	}
}
