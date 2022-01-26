package inventory

const (
	HealingPotion      PotionType = "HealingPotion"
	ManaPotion         PotionType = "ManaPotion"
	RejuvenationPotion PotionType = "RejuvenationPotion"
)

type Belt struct {
	Potions []Potion
}

func (b Belt) GetFirstPotion(potionType PotionType) (Potion, bool) {
	for _, p := range b.Potions {
		if p.Type == potionType && p.Row == 0 {
			return p, true
		}
	}

	return Potion{}, false
}

type PotionType string
type Potion struct {
	Row    int
	Column int
	Type   PotionType
}
