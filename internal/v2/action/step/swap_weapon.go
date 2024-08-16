package step

import (
	"time"

	"github.com/hectorgimenez/koolo/internal/v2/context"

	"github.com/hectorgimenez/d2go/pkg/data/skill"
)

func SwapToMainWeapon() error {
	return swapWeapon(false)
}

func SwapToCTA() error {
	return swapWeapon(true)
}

func swapWeapon(toCTA bool) error {
	lastRun := time.Time{}

	ctx := context.Get()
	for {
		// Pause the execution if the priority is not the same as the execution priority
		if ctx.ExecutionPriority != ctx.Priority {
			continue
		}

		if time.Since(lastRun) < time.Second {
			continue
		}

		_, found := ctx.Data.PlayerUnit.Skills[skill.BattleOrders]
		if (toCTA && found) || (!toCTA && !found) {
			return nil
		}

		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)

		lastRun = time.Now()
	}
}