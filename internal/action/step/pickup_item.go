package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

var (
	ErrItemTooFar        = errors.New("item is too far away")
	ErrNoLOSToItem       = errors.New("no line of sight to item")
	ErrMonsterAroundItem = errors.New("monsters detected around item")
)

func PickupItem(it data.Item) error {
	ctx := context.Get()
	ctx.SetLastStep("PickupItem")

	// Double check for monsters around the item
	for _, monster := range ctx.Data.Monsters.Enemies() {
		if monster.Stats[stat.Life] > 0 && pather.DistanceFromPoint(it.Position, monster.Position) <= 4 {
			return ErrMonsterAroundItem
		}
	}
	// Check if we have line of sight to item, maybe is behind a wall.
	if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, it.Position) {
		return ErrNoLOSToItem
	}

	// Check distance to item
	distance := ctx.PathFinder.DistanceFromMe(it.Position)
	if distance >= 6 {
		return fmt.Errorf("%w (%d): %s", ErrItemTooFar, distance, it.Desc().Name)
	}

	ctx.Logger.Debug(fmt.Sprintf("Picking up: %s [%s]", it.Desc().Name, it.Quality.ToString()))

	waitingForInteraction := time.Time{}
	mouseOverAttempts := 0
	lastRun := time.Now()
	itemToPickup := it

	// Base click position  -1 offset (Resurrected item drop spacing)
	baseX := it.Position.X - 1
	baseY := it.Position.Y - 1
	mX, mY := ui.GameCoordsToScreenCords(baseX, baseY)

	for {
		ctx.PauseIfNotPriority()

		if time.Since(lastRun) < 30*time.Millisecond {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		ctx.RefreshGameData()
		lastRun = time.Now()

		// Verify item still exists
		var currentItem data.Item
		for _, i := range ctx.Data.Inventory.ByLocation(item.LocationGround) {
			if i.UnitID == itemToPickup.UnitID {
				currentItem = i
				break
			}
		}

		if currentItem.UnitID != itemToPickup.UnitID {
			ctx.Logger.Info(fmt.Sprintf("Picked up: %s [%s]", itemToPickup.Desc().Name, itemToPickup.Quality.ToString()))
			return nil // Item picked up
		}

		if mouseOverAttempts > 45 || (!waitingForInteraction.IsZero() && time.Since(waitingForInteraction) > time.Second*3) {
			return fmt.Errorf("couldn't hover item %s after %d attempts", it.Desc().Name, mouseOverAttempts)
		}

		// Calculate click position
		offsetX, offsetY := utils.ItemSpiral(mouseOverAttempts)
		currentX := mX + offsetX
		currentY := mY + offsetY

		// Move and check hover
		ctx.HID.MovePointer(currentX, currentY)
		time.Sleep(75 * time.Millisecond)

		ctx.RefreshGameData()

		if currentItem.IsHovered {
			ctx.HID.Click(game.LeftButton, currentX, currentY)
			time.Sleep(25 * time.Millisecond)

			if waitingForInteraction.IsZero() {
				waitingForInteraction = time.Now()
			}
			continue
		}

		// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
		// on Andariel, so we open it
		if isChestHovered() {
			ctx.HID.Click(game.LeftButton, currentX, currentY)
			time.Sleep(50 * time.Millisecond)
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
