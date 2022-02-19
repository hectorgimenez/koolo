package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type PickupItem struct {
	basicStep
	item                  game.Item
	waitingForInteraction bool
}

func NewPickupItem(item game.Item) *PickupItem {
	return &PickupItem{
		basicStep: newBasicStep(),
		item:      item,
	}
}

func (p PickupItem) Status(data game.Data) Status {
	for _, i := range data.Items.Ground {
		if i.ID == p.item.ID {
			return p.status
		}
	}

	return p.tryTransitionStatus(StatusCompleted)
}

func (p PickupItem) Run(data game.Data) error {
	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) < time.Millisecond*500 {
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
				path, distance, _ := helper.GetPathToDestination(data, i.Position.X-2, i.Position.Y-2)
				if distance > 15 {
					helper.MoveThroughPath(path, 15, false)
					return nil
				}
				helper.MoveThroughPath(path, 0, false)

				return nil
			}
		}
	}

	return fmt.Errorf("item %s not found", p.item.Name)
}
