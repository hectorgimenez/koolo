package step

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
	"time"
)

const maxInteractions = 45

type PickupItemStep struct {
	basicStep
	item                  data.Item
	waitingForInteraction bool
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

	for _, i := range d.Items.Ground {
		if i.UnitID == p.item.UnitID {
			return p.status
		}
	}

	p.logger.Info(fmt.Sprintf("Item picked up: %s [%s]", p.item.Name, p.item.Quality.ToString()))

	return p.tryTransitionStatus(StatusCompleted)
}

func (p *PickupItemStep) Run(d data.Data) error {
	if p.mouseOverAttempts > maxInteractions {
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

	if p.waitingForInteraction && time.Since(p.lastRun) < time.Second*2 {
		return nil
	}

	// Set teleport for first time
	if p.lastRun.IsZero() {
		hid.PressKey(config.Config.Bindings.Teleport)
	}

	p.lastRun = time.Now()
	for _, i := range d.Items.Ground {
		if i.UnitID == p.item.UnitID {
			if i.IsHovered {
				hid.Click(hid.LeftButton)
				p.waitingForInteraction = true
				return nil
			} else {
				// Sometimes we got stuck because mouse is hovering a chest and item is in behind, it usually happens a lot
				// on Andariel
				if p.isChestHovered(d) {
					hid.Click(hid.LeftButton)
				}

				path, distance, _ := pather.GetPath(d, i.Position)
				if distance > 6 {
					pather.MoveThroughPath(path, 15, true)
					return nil
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
