package data

const (
	HealingPotion      PotionType = "healing"
	ManaPotion         PotionType = "mana"
	RejuvenationPotion PotionType = "rejuvenation"
)

type Belt struct {
	Potions []Potion
}

func (b Belt) GetFirstPotion(potionType PotionType) (Potion, bool) {
	for _, p := range b.Potions {
		if p.Type == potionType && p.Position.Y == 0 {
			return p, true
		}
	}

	return Potion{}, false
}

type PotionType string
type Potion struct {
	BaseItem
	Type PotionType
}
