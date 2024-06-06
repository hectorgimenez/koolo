package step

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const maxInteractions = 45

type PickupItemStep struct {
	basicStep
	item                  data.Item
	waitingForInteraction time.Time
	mouseOverAttempts     int
	logger                *slog.Logger
	startedAt             time.Time
	currentMouseCoords    data.Position
}

func PickupItem(logger *slog.Logger, item data.Item) *PickupItemStep {
	return &PickupItemStep{
		basicStep: newBasicStep(),
		item:      item,
		logger:    logger,
	}
}

func (p *PickupItemStep) Status(d game.Data, _ container.Container) Status {
	if p.status == StatusCompleted {
		return p.status
	}

	for _, i := range d.Inventory.ByLocation(item.LocationGround) {
		if i.UnitID == p.item.UnitID {
			return p.status
		}
	}

	p.logger.Info(fmt.Sprintf("Item picked up: %s [%s]", p.item.Desc().Name, p.item.Quality.ToString()))

	return p.tryTransitionStatus(StatusCompleted)
}

func (p *PickupItemStep) Run(d game.Data, container container.Container) error {
	for _, m := range d.Monsters.Enemies() {
		if dist := pather.DistanceFromMe(d, m.Position); dist < 7 && p.mouseOverAttempts > 1 {
			return fmt.Errorf("monster %d [%s] is too close to item %s [%s]", m.Name, m.Type, p.item.Desc().Name, p.item.Quality.ToString())
		}
	}

	if p.mouseOverAttempts > maxInteractions || !p.waitingForInteraction.IsZero() && time.Since(p.waitingForInteraction) > time.Second*3 {
		return fmt.Errorf("item %s [%s] could not be picked up", p.item.Desc().Name, p.item.Quality.ToString())
	}

	if p.status == StatusNotStarted {
		p.logger.Debug(fmt.Sprintf("Picking up: %s [%s]", p.item.Desc().Name, p.item.Quality.ToString()))
		p.startedAt = time.Now()
	}

	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) < helper.RandomDurationMs(120, 320) {
		return nil
	}

	if !p.waitingForInteraction.IsZero() && time.Since(p.lastRun) < time.Second {
		return nil
	}

	p.lastRun = time.Now()
	for _, i := range d.Inventory.ByLocation(item.LocationGround) {
		if i.UnitID == p.item.UnitID {
			objectX := i.Position.X - 1
			objectY := i.Position.Y - 1
			mX, mY := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX, objectY)

			if i.IsHovered {
				container.HID.Click(game.LeftButton, p.currentMouseCoords.X, p.currentMouseCoords.Y)
				if p.waitingForInteraction.IsZero() {
					p.waitingForInteraction = time.Now()
				}
				return nil
			} else {
				// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
				// on Andariel, so we open it
				if p.isChestHovered(d) {
					container.HID.Click(game.LeftButton, p.currentMouseCoords.X, p.currentMouseCoords.Y)
				}

				distance := pather.DistanceFromMe(d, i.Position)
				if distance > 7 {
					p.logger.Info("item is too far away", slog.String("item", p.item.Desc().Name))
					return fmt.Errorf("item is too far away: %s", p.item.Desc().Name)
				}

				x, y := helper.Spiral(p.mouseOverAttempts)
				p.currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
				container.HID.MovePointer(mX+x, mY+y)
				p.mouseOverAttempts++

				return nil
			}
		}
	}

	return fmt.Errorf("item %s not found", p.item.Desc().Name)
}

func (p *PickupItemStep) isChestHovered(d game.Data) bool {
	for _, o := range d.Objects {
		if o.IsChest() && o.IsHovered {
			return true
		}
	}

	return false
}
