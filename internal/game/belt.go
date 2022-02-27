package game

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
		// Ensure potion is in row 0 and one of the four columns
		if p.Type == potionType && p.Position.Y == 0 && (p.Position.X == 0 || p.Position.X == 1 || p.Position.X == 2 || p.Position.X == 3) {
			return p, true
		}
	}

	return Potion{}, false
}

type PotionType string
type Potion struct {
	Item
	Type PotionType
}
