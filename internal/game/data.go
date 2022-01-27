package game

type DataRepository interface {
	GameData() Data
}

type Area string
type Corpse struct {
	Found bool
	X     int
	Y     int
}
type Data struct {
	Area   Area
	Corpse Corpse
}

func (b Bot) data() Data {
	return b.dataRepository.GameData()
}
