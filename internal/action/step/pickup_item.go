package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

const maxInteractions = 45

type PickupItemStep struct {
	basicStep
	item                  data.Item
	waitingForInteraction time.Time
	mouseOverAttempts     int
	logger                *zap.Logger
	startedAt             time.Time
}

func PickupItem(logger *zap.Logger, item data.Item) *PickupItemStep {
	return &PickupItemStep{
		basicStep: newBasicStep(),
		item:      item,
		logger:    logger,
	}
}

func (p *PickupItemStep) Status(d data.Data) Status {
	if p.status == StatusCompleted {
		return p.status
	}

	for _, i := range d.Items.ByLocation(item.LocationGround) {
		if i.UnitID == p.item.UnitID {
			return p.status
		}
	}

	p.logger.Info(fmt.Sprintf("Item picked up: %s [%s]", p.item.Name, p.item.Quality.ToString()))

	return p.tryTransitionStatus(StatusCompleted)
}

func (p *PickupItemStep) Run(d data.Data) error {
	for _, m := range d.Monsters.Enemies() {
		if dist := pather.DistanceFromMe(d, m.Position); dist < 7 && p.mouseOverAttempts > 1 {
			return fmt.Errorf("monster %d [%s] is too close to item %s [%s]", m.Name, m.Type, p.item.Name, p.item.Quality.ToString())
		}
	}

	if p.mouseOverAttempts > maxInteractions || !p.waitingForInteraction.IsZero() && time.Since(p.waitingForInteraction) > time.Second*3 {
		return fmt.Errorf("item %s [%s] could not be picked up", p.item.Name, p.item.Quality.ToString())
	}

	if p.status == StatusNotStarted {
		p.logger.Debug(fmt.Sprintf("Picking up: %s [%s]", p.item.Name, p.item.Quality.ToString()))
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
	for _, i := range d.Items.ByLocation(item.LocationGround) {
		if i.UnitID == p.item.UnitID {
			if i.IsHovered {
				hid.Click(hid.LeftButton)
				if p.waitingForInteraction.IsZero() {
					p.waitingForInteraction = time.Now()
				}
				return nil
			} else {
				// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
				// on Andariel, so we open it
				if p.isChestHovered(d) {
					hid.Click(hid.LeftButton)
				}

				distance := pather.DistanceFromMe(d, i.Position)
				if distance > 7 {
					p.logger.Info("item is too far away", zap.String("item", string(p.item.Name)))
					return fmt.Errorf("item is too far away: %s", p.item.Name)
				}

				objectX := i.Position.X - 1
				objectY := i.Position.Y - 1
				mX, mY := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX, objectY)
				x, y := helper.Spiral(p.mouseOverAttempts)
				hid.MovePointer(mX+x, mY+y)
				p.mouseOverAttempts++

				return nil
			}
		}
	}

	return fmt.Errorf("item %s not found", p.item.Name)
}

func (p *PickupItemStep) isChestHovered(d data.Data) bool {
	for _, o := range d.Objects {
		if o.IsChest() && o.IsHovered {
			return true
		}
	}

	return false
}
