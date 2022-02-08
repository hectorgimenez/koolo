package data

type Status struct {
	Life    int
	MaxLife int
	Mana    int
	MaxMana int
	Merc    MercStatus
}

type MercStatus struct {
	Alive   bool
	Life    int
	MaxLife int
}

func (s Status) HPPercent() int {
	return int((float64(s.Life) / float64(s.MaxLife)) * 100)
}

func (s Status) MPPercent() int {
	return int((float64(s.Mana) / float64(s.MaxMana)) * 100)
}

func (s Status) MercHPPercent() int {
	return int((float64(s.Merc.Life) / float64(s.Merc.MaxLife)) * 100)
}
