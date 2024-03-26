package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	hid2 "github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"log/slog"
	"slices"
)

const (
	wpTabStartX     = 131
	wpTabStartY     = 148
	wpTabSizeX      = 57
	wpListPositionX = 200
	wpListStartY    = 158
	wpAreaBtnHeight = 41
)

func (b *Builder) WayPoint(a area.Area) *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		// We don't need to move, we are already at destination area
		if d.PlayerUnit.Area == a {
			return nil
		}

		return []Action{
			b.openWPAndSelectTab(a, d),
			b.useWP(a),
		}
	})
}

func (b *Builder) openWPAndSelectTab(a area.Area, d data.Data) Action {
	wpCoords, found := area.WPAddresses[a]
	if !found {
		panic("Area destination is not mapped on WayPoint Action (waypoint.go)")
	}

	for _, o := range d.Objects {
		if o.IsWaypoint() {
			return b.InteractObject(o.Name, func(d data.Data) bool {
				return d.OpenMenus.Waypoint
			},
				step.SyncStep(func(d data.Data) error {
					actTabX := wpTabStartX + (wpCoords.Tab-1)*wpTabSizeX + (wpTabSizeX / 2)

					b.HID.Click(hid2.LeftButton, actTabX, wpTabStartY)
					helper.Sleep(200)

					return nil
				}),
			)
		}
	}

	return nil
}

func (b *Builder) useWP(a area.Area) *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		finalDestination := a
		traverseAreas := make([]area.Area, 0)
		currentWP := area.WPAddresses[a]
		if !slices.Contains(d.PlayerUnit.AvailableWaypoints, a) {
			for {
				traverseAreas = append(currentWP.LinkedFrom, traverseAreas...)

				if slices.Contains(d.PlayerUnit.AvailableWaypoints, a) {
					break
				}

				currentWP = area.WPAddresses[currentWP.LinkedFrom[0]]

				if currentWP.LinkedFrom != nil {
					a = currentWP.LinkedFrom[0]
				}
			}
		}

		currentWP = area.WPAddresses[a]

		// First use the previous available waypoint that we have discovered
		actions = append(actions, NewStepChain(func(d data.Data) []step.Step {
			return []step.Step{
				step.SyncStep(func(d data.Data) error {
					areaBtnY := wpListStartY + (currentWP.Row-1)*wpAreaBtnHeight + (wpAreaBtnHeight / 2)
					b.HID.Click(hid2.LeftButton, wpListPositionX, areaBtnY)
					helper.Sleep(1000)

					return nil
				}),
			}
		}))

		// We have the WP discovered, just use it
		if len(traverseAreas) == 0 {
			return actions
		}

		traverseAreas = append(traverseAreas, finalDestination)

		// Next keep traversing all the areas from the previous available waypoint until we reach the destination, trying to discover WPs during the way
		b.Logger.Info("Traversing areas to reach destination", slog.Any("areas", traverseAreas))

		for i, dst := range traverseAreas {
			if !dst.IsTown() {
				actions = append(actions,
					b.Buff(),
				)
			}

			if i > 0 {
				actions = append(actions,
					b.MoveToArea(dst),
					b.DiscoverWaypoint(),
				)
			}
		}

		return actions
	})
}
