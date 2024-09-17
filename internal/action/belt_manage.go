package action

import (
	"log/slog"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/context"
)

func ManageBelt() error {

	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ManageBelt"

	// Check for misplaced potions
	misplacedPotions := checkMisplacedPotions()
	haveMisplacedPotions := len(misplacedPotions) > 0

	// Consume misplaced potions
	for haveMisplacedPotions {

		for _, potion := range misplacedPotions {
			slog.Info("Consuming misplaced potion", "potion", potion.Name, "position", potion.Position)
			ctx.HID.PressKey(ctx.Data.KeyBindings.UseBelt[potion.Position.X].Key1[0])
			time.Sleep(500 * time.Millisecond)
		}

		misplacedPotions = checkMisplacedPotions()
		haveMisplacedPotions = len(misplacedPotions) > 0

		if !haveMisplacedPotions {
			break
		}
	}

	return nil
}

func checkMisplacedPotions() []data.Item {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "CheckMisplacedPotions"

	// Get list of potions in the first row
	potions := []data.Item{}

	for _, pot := range ctx.Data.Inventory.Belt.Items {
		if pot.Position.Y == 0 && (pot.Position.X == 0 || pot.Position.X == 1 || pot.Position.X == 2 || pot.Position.X == 3) {
			potions = append(potions, pot)
		}
	}

	expectedPotions := ctx.Data.CharacterCfg.Inventory.BeltColumns

	// Count occurrences of each potion type
	expectedCounts := make(map[string]int)
	for _, potion := range expectedPotions {
		expectedCounts[potion]++
	}

	// Helper function to match potion types
	matchPotionType := func(actualPotion, expectedType string) bool {
		return strings.Contains(strings.ToLower(actualPotion), expectedType)
	}

	// Check for misplaced potions
	misplacedPotions := []data.Item{}

	for _, potion := range potions {
		matched := false
		for expectedType, count := range expectedCounts {
			if matchPotionType(string(potion.Name), expectedType) {
				if count > 0 {
					expectedCounts[expectedType]--
				} else {
					misplacedPotions = append(misplacedPotions, potion)
				}
				matched = true
				break
			}
		}
		if !matched {
			misplacedPotions = append(misplacedPotions, potion)
		}
	}

	return misplacedPotions
}
