package health

type Repository interface {
	CurrentStatus() (Status, error)
}

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
	return (s.Life / s.MaxLife) * 100
}

func (s Status) MPPercent() int {
	return (s.Mana / s.MaxMana) * 100
}

func (s Status) MercHPPercent() int {
	return (s.Merc.Life / s.Merc.MaxLife) * 100
}
