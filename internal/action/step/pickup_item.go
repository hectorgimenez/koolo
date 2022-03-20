package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
	"time"
)

const maxInteractions = 45

type PickupItemStep struct {
	basicStep
	item                  game.Item
	waitingForInteraction bool
	mouseOverAttempts     int
	logger                *zap.Logger
}

func PickupItem(logger *zap.Logger, item game.Item) *PickupItemStep {
	return &PickupItemStep{
		basicStep: newBasicStep(),
		item:      item,
		logger:    logger,
	}
}

func (p *PickupItemStep) Status(data game.Data) Status {
	if p.status == StatusCompleted {
		return p.status
	}

	for _, i := range data.Items.Ground {
		if i.ID == p.item.ID {
			return p.status
		}
	}

	return p.tryTransitionStatus(StatusCompleted)
}

func (p *PickupItemStep) Run(data game.Data) error {
	if p.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("item %s [%s] could not be picked up", p.item.Name, p.item.Quality)
	}

	if p.status == StatusNotStarted {
		p.logger.Info(fmt.Sprintf("Picking up: %s [%s]", p.item.Name, p.item.Quality))
	}

	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) < time.Millisecond*200 {
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
	for _, i := range data.Items.Ground {
		if i.ID == p.item.ID {
			if i.IsHovered {
				hid.Click(hid.LeftButton)
				p.waitingForInteraction = true
				return nil
			} else {
				path, distance, _ := pather.GetPathToDestination(data, i.Position.X, i.Position.Y)
				if distance > 5 {
					pather.MoveThroughPath(path, 15, true)
					return nil
				}

				objectX := i.Position.X - 1
				objectY := i.Position.Y - 1
				mX, mY := pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, objectX, objectY)
				x, y := helper.Spiral(p.mouseOverAttempts)
				hid.MovePointer(mX+x, mY+y)
				p.mouseOverAttempts++

				return nil
			}
		}
	}

	return fmt.Errorf("item %s not found", p.item.Name)
}
