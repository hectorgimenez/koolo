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

var (
	ErrItemTooFar  = errors.New("item is too far away")
	ErrNoLOSToItem = errors.New("no line of sight to item")
)

func PickupItem(it data.Item) error {
	ctx := context.Get()
	ctx.SetLastStep("PickupItem")

	if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, it.Position) {
		return ErrNoLOSToItem
	}

	// Check distance to item
	distance := ctx.PathFinder.DistanceFromMe(it.Position)
	if distance > 10 {
		return fmt.Errorf("%w (%d): %s", ErrItemTooFar, distance, it.Desc().Name)
	}

	ctx.Logger.Debug(fmt.Sprintf("Picking up: %s [%s]", it.Desc().Name, it.Quality.ToString()))

	waitingForInteraction := time.Time{}
	mouseOverAttempts := 0
	lastRun := time.Now()
	itemToPickup := it
	var currentX, currentY int

	for {
		ctx.PauseIfNotPriority()

		// Reset item to empty
		it = data.Item{}

		for _, i := range ctx.Data.Inventory.ByLocation(item.LocationGround) {
			if i.UnitID == itemToPickup.UnitID {
				it = i
			}
		}

		if it.UnitID != itemToPickup.UnitID {
			ctx.Logger.Info(fmt.Sprintf("Picked up: %s [%s]", itemToPickup.Desc().Name, itemToPickup.Quality.ToString()))
			return nil
		}

		if mouseOverAttempts > maxInteractions || (!waitingForInteraction.IsZero() && time.Since(waitingForInteraction) > time.Second*3) {
			return fmt.Errorf("item %s [%s] could not be picked up: mouseover attempts limit reached", it.Desc().Name, it.Quality.ToString())
		}
		// Rate limit attempts
		if time.Since(lastRun) < utils.RandomDurationMs(100, 250) {
			continue
		}

		lastRun = time.Now()
		objectX := it.Position.X - 1
		objectY := it.Position.Y - 1

		// Calculate screen coordinates for item
		mX, mY := ui.GameCoordsToScreenCords(objectX, objectY)

		if mouseOverAttempts == 0 {
			// First attempt - try direct center
			currentX, currentY = mX, mY
			ctx.HID.MovePointer(currentX, currentY)
		} else {
			// Subsequent attempts - use spiral pattern
			x, y := utils.ItemSpiral(mouseOverAttempts)
			currentX, currentY = mX+x, mY+y
			ctx.HID.MovePointer(currentX, currentY)
		}

		// Wait a bit and get fresh data to check hover state
		time.Sleep(50 * time.Millisecond)
		ctx.RefreshGameData()

		// Find our item again with fresh data
		var currentItem data.Item
		for _, i := range ctx.Data.Inventory.ByLocation(item.LocationGround) {
			if i.UnitID == itemToPickup.UnitID {
				currentItem = i
				break
			}
		}

		if currentItem.IsHovered {
			// Click at exact coordinates where mouse is currently positioned
			ctx.HID.Click(game.LeftButton, currentX, currentY)
			if waitingForInteraction.IsZero() {
				waitingForInteraction = time.Now()
			}
			continue
		}

		// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
		// on Andariel, so we open it
		if isChestHovered() {
			ctx.HID.Click(game.LeftButton, currentX, currentY)
		}

		mouseOverAttempts++
	}
}
func isChestHovered() bool {
	for _, o := range context.Get().Data.Objects {
		if o.IsChest() && o.IsHovered {
			return true
		}
	}
	return false
}
