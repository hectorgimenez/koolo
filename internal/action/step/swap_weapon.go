package step

import (
	"time"

	"github.com/hectorgimenez/koolo/internal/context"
)

func SwapToMainWeapon() error {
	return swapWeapon(false)
}

func SwapToSecondary() error {
	return swapWeapon(true)
}

func swapWeapon(toSecondary bool) error {
	lastRun := time.Time{}

	ctx := context.Get()
	ctx.SetLastStep("swapWeapon")

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		if time.Since(lastRun) < time.Millisecond*500 {
			continue
		}

		onPrimary := ctx.Data.ActiveWeaponSlot == 0
		needsSwap := (onPrimary && toSecondary) || (!onPrimary && !toSecondary)

		if !needsSwap {
			return nil
		}

		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)

		lastRun = time.Now()
	}
}
