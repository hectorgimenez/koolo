package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxInteractions = 24 // 25 attempts since we start at 0
	clickDelay      = 25 * time.Millisecond
	pickupTimeout   = 3 * time.Second
)

var (
	ErrItemTooFar        = errors.New("item is too far away")
	ErrNoLOSToItem       = errors.New("no line of sight to item")
	ErrMonsterAroundItem = errors.New("monsters detected around item")
	ErrCastingMoving     = errors.New("char casting or moving")
)

func PickupItem(it data.Item, itemPickupAttempt int) error {
	ctx := context.Get()
	ctx.SetLastStep("PickupItem")

	// Casting skill/moving return back
	for ctx.Data.PlayerUnit.Mode == mode.CastingSkill || ctx.Data.PlayerUnit.Mode == mode.Running || ctx.Data.PlayerUnit.Mode == mode.Walking || ctx.Data.PlayerUnit.Mode == mode.WalkingInTown {
		time.Sleep(25 * time.Millisecond)
		return ErrCastingMoving
	}

	// Calculate base screen position for item
	baseX := it.Position.X - 1
	baseY := it.Position.Y - 1
	baseScreenX, baseScreenY := ctx.PathFinder.GameCoordsToScreenCords(baseX, baseY)

	// Check for monsters first
	if hasHostileMonstersNearby(it.Position) {
		return ErrMonsterAroundItem
	}

	// Validate line of sight
	if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, it.Position) {
		return ErrNoLOSToItem
	}

	// Check distance
	distance := ctx.PathFinder.DistanceFromMe(it.Position)
	if distance >= 7 {
		return fmt.Errorf("%w (%d): %s", ErrItemTooFar, distance, it.Desc().Name)
	}

	ctx.Logger.Debug(fmt.Sprintf("Picking up: %s [%s]", it.Desc().Name, it.Quality.ToString()))

	// Track interaction state
	waitingForInteraction := time.Time{}
	spiralAttempt := 0
	targetItem := it
	lastMonsterCheck := time.Now()
	const monsterCheckInterval = 150 * time.Millisecond

	startTime := time.Now()

	for {
		ctx.PauseIfNotPriority()
		ctx.RefreshGameData()

		// Periodic monster check
		if time.Since(lastMonsterCheck) > monsterCheckInterval {
			if hasHostileMonstersNearby(it.Position) {
				return ErrMonsterAroundItem
			}
			lastMonsterCheck = time.Now()
		}

		// Check if item still exists
		currentItem, exists := findItemOnGround(targetItem.UnitID)
		if !exists {
			ctx.Logger.Info(fmt.Sprintf("Picked up: %s [%s] | Item Pickup Attempt:%d | Spiral Attempt:%d", targetItem.Desc().Name, targetItem.Quality.ToString(), itemPickupAttempt, spiralAttempt))
			return nil // Success!
		}

		// Check timeout conditions
		if spiralAttempt > maxInteractions ||
			(!waitingForInteraction.IsZero() && time.Since(waitingForInteraction) > pickupTimeout) ||
			time.Since(startTime) > pickupTimeout {
			return fmt.Errorf("failed to pick up %s after %d attempts", it.Desc().Name, spiralAttempt)
		}

		offsetX, offsetY := utils.ItemSpiral(spiralAttempt)
		cursorX := baseScreenX + offsetX
		cursorY := baseScreenY + offsetY

		// Move cursor directly to target position
		ctx.HID.MovePointer(cursorX, cursorY)

		// Click on item if mouse is hovering over
		if currentItem.UnitID == ctx.GameReader.GameReader.GetData().HoverData.UnitID {
			ctx.HID.Click(game.LeftButton, cursorX, cursorY)
			time.Sleep(clickDelay)

			if waitingForInteraction.IsZero() {
				waitingForInteraction = time.Now()
			}
			continue
		}

		// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
		// on Andariel, so we open it
		if isChestHovered() {
			ctx.HID.Click(game.LeftButton, cursorX, cursorY)
			time.Sleep(50 * time.Millisecond)
		}

		spiralAttempt++
	}
}
func hasHostileMonstersNearby(pos data.Position) bool {
	ctx := context.Get()

	for _, monster := range ctx.Data.Monsters.Enemies() {
		if monster.Stats[stat.Life] > 0 && pather.DistanceFromPoint(pos, monster.Position) <= 4 {
			return true
		}
	}
	return false
}

func findItemOnGround(targetID data.UnitID) (data.Item, bool) {
	ctx := context.Get()

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationGround) {
		if i.UnitID == targetID {
			return i, true
		}
	}
	return data.Item{}, false
}

func isChestHovered() bool {
	ctx := context.Get()

	for _, o := range ctx.Data.Objects {
		if o.IsChest() && o.IsHovered {
			return true
		}
	}
	return false
}
