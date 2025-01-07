package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const maxInteractions = 45

var ErrItemTooFar = errors.New("item is too far away")

func PickupItem(it data.Item) error {
	ctx := context.Get()
	ctx.SetLastStep("PickupItem")

	ctx.Logger.Debug(fmt.Sprintf("Picking up: %s [%s]", it.Desc().Name, it.Quality.ToString()))

	mouseOverAttempts := 0
	lastRun := time.Now()

	for mouseOverAttempts < maxInteractions {
		ctx.PauseIfNotPriority()

		// Check if item was picked up
		found := false
		for _, i := range ctx.Data.Inventory.ByLocation(item.LocationGround) {
			if i.UnitID == it.UnitID {
				it = i
				found = true
				break
			}
		}
		if !found {
			ctx.Logger.Info(fmt.Sprintf("Picked up: %s [%s]", it.Desc().Name, it.Quality.ToString()))
			return nil
		}

		// Rate limit attempts
		if time.Since(lastRun) < utils.RandomDurationMs(120, 320) {
			continue
		}
		lastRun = time.Now()

		// Check distance to item
		distance := ctx.PathFinder.DistanceFromMe(it.Position)
		if distance > 10 {
			return fmt.Errorf("%w (%d): %s", ErrItemTooFar, distance, it.Desc().Name)
		}

		// Calculate screen coordinates for item
		mX, mY := ui.GameCoordsToScreenCords(it.Position.X-1, it.Position.Y-1)
		x, y := utils.ItemSpiral(mouseOverAttempts)
		ctx.HID.MovePointer(mX+x, mY+y)

		time.Sleep(50 * time.Millisecond)

		if it.IsHovered {
			ctx.HID.Click(game.LeftButton, mX+x, mY+y)
			// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
			// on Andariel, so we open it
		} else if isChestHovered() {
			ctx.HID.Click(game.LeftButton, mX+x, mY+y)
		}

		mouseOverAttempts++
	}

	return fmt.Errorf("item %s [%s] could not be picked up: mouseover attempts limit reached", it.Desc().Name, it.Quality.ToString())
}
func isChestHovered() bool {
	for _, o := range context.Get().Data.Objects {
		if o.IsChest() && o.IsHovered {
			return true
		}
	}

	return false
}
