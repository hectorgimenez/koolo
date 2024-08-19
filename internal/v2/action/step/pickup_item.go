package step

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/ui"
	"github.com/hectorgimenez/koolo/internal/v2/utils"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
)

const maxInteractions = 45

func PickupItem(it data.Item) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "PickupItem"

	ctx.Logger.Debug(fmt.Sprintf("Picking up: %s [%s]", it.Desc().Name, it.Quality.ToString()))

	waitingForInteraction := time.Time{}
	mouseOverAttempts := 0
	currentMouseCoords := data.Position{}
	lastRun := time.Now()

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		for _, i := range ctx.Data.Inventory.AllItems {
			if i.UnitID == it.UnitID {
				it = i
			}
		}

		if it.Location.LocationType != item.LocationGround {
			return nil
		}

		if mouseOverAttempts > maxInteractions || !waitingForInteraction.IsZero() && time.Since(waitingForInteraction) > time.Second*3 {
			return fmt.Errorf("item %s [%s] could not be picked up", it.Desc().Name, it.Quality.ToString())
		}

		if time.Since(lastRun) < utils.RandomDurationMs(120, 320) {
			continue
		}

		if !waitingForInteraction.IsZero() && time.Since(lastRun) < time.Second {
			continue
		}

		lastRun = time.Now()
		objectX := it.Position.X - 1
		objectY := it.Position.Y - 1
		mX, mY := ui.GameCoordsToScreenCords(objectX, objectY)

		if it.IsHovered {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			if waitingForInteraction.IsZero() {
				waitingForInteraction = time.Now()
			}
			continue
		} else {
			// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
			// on Andariel, so we open it
			if isChestHovered() {
				ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			}

			distance := ctx.PathFinder.DistanceFromMe(it.Position)
			if distance > 10 {
				ctx.Logger.Info("item is too far away", slog.String("item", it.Desc().Name))
				return fmt.Errorf("item is too far away: %s", it.Desc().Name)
			}

			x, y := utils.Spiral(mouseOverAttempts)
			currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
			ctx.HID.MovePointer(mX+x, mY+y)
			mouseOverAttempts++
		}
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
