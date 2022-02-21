package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type PickupItemStep struct {
	basicStep
	item                  game.Item
	waitingForInteraction bool
}

func PickupItem(item game.Item) *PickupItemStep {
	return &PickupItemStep{
		basicStep: newBasicStep(),
		item:      item,
	}
}

func (p *PickupItemStep) Status(data game.Data) Status {
	for _, i := range data.Items.Ground {
		if i.ID == p.item.ID {
			return p.status
		}
	}

	return p.tryTransitionStatus(StatusCompleted)
}

func (p *PickupItemStep) Run(data game.Data) error {
	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) < time.Second {
		return nil
	}

	if p.waitingForInteraction && time.Since(p.lastRun) < time.Second*2 {
		return nil
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
					pather.MoveThroughPath(path, 15, false)
					return nil
				}
				x, y := pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, i.Position.X-2, i.Position.Y-2)
				hid.MovePointer(x, y)

				return nil
			}
		}
	}

	return fmt.Errorf("item %s not found", p.item.Name)
}
