package game

import "strings"

const (
	HealingPotion      PotionType = "HealingPotion"
	ManaPotion         PotionType = "ManaPotion"
	RejuvenationPotion PotionType = "RejuvenationPotion"
)

type Belt []Item

func (b Belt) GetFirstPotion(potionType PotionType) (Position, bool) {
	for _, i := range b {
		// Ensure potion is in row 0 and one of the four columns
		if strings.Contains(string(i.Name), string(potionType)) && i.Position.Y == 0 && (i.Position.X == 0 || i.Position.X == 1 || i.Position.X == 2 || i.Position.X == 3) {
			return i.Position, true
		}
	}

	return Position{}, false
}

type PotionType string
