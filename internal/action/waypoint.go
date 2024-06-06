package action

import (
	"log/slog"
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func (b *Builder) WayPoint(a area.ID) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		// We don't need to move, we are already at destination area
		if d.PlayerUnit.Area == a {
			return nil
		}

		return []Action{
			b.openWPAndSelectTab(a, d),
			b.useWP(a),
			b.Buff(),
		}
	})
}

func (b *Builder) openWPAndSelectTab(a area.ID, d game.Data) Action {
	wpCoords, found := area.WPAddresses[a]
	if !found {
		panic("Area destination is not mapped on WayPoint Action (waypoint.go)")
	}

	for _, o := range d.Objects {
		if o.IsWaypoint() {
			return b.InteractObject(o.Name, func(d game.Data) bool {
				return d.OpenMenus.Waypoint
			},
				step.SyncStep(func(d game.Data) error {

					if d.LegacyGraphics {
						actTabX := ui.WpTabStartXClassic + (wpCoords.Tab-1)*ui.WpTabSizeXClassic + (ui.WpTabSizeXClassic / 2)
						b.HID.Click(game.LeftButton, actTabX, ui.WpTabStartYClassic)
					} else {
						actTabX := ui.WpTabStartX + (wpCoords.Tab-1)*ui.WpTabSizeX + (ui.WpTabSizeX / 2)
						b.HID.Click(game.LeftButton, actTabX, ui.WpTabStartY)
					}
					helper.Sleep(200)

					return nil
				}),
			)
		}
	}

	return nil
}

func (b *Builder) useWP(a area.ID) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		finalDestination := a
		traverseAreas := make([]area.ID, 0)
		currentWP := area.WPAddresses[a]
		if !slices.Contains(d.PlayerUnit.AvailableWaypoints, a) {
			for {
				traverseAreas = append(currentWP.LinkedFrom, traverseAreas...)

				if currentWP.LinkedFrom != nil {
					a = currentWP.LinkedFrom[0]
				}

				currentWP = area.WPAddresses[currentWP.LinkedFrom[0]]

				if slices.Contains(d.PlayerUnit.AvailableWaypoints, a) {
					break
				}
			}
		}

		currentWP = area.WPAddresses[a]

		// First use the previous available waypoint that we have discovered
		actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
			return []step.Step{
				step.SyncStep(func(d game.Data) error {
					if d.LegacyGraphics {
						areaBtnY := ui.WpListStartYClassic + (currentWP.Row-1)*ui.WpAreaBtnHeightClassic + (ui.WpAreaBtnHeightClassic / 2)
						b.HID.Click(game.LeftButton, ui.WpListPositionXClassic, areaBtnY)
					} else {
						areaBtnY := ui.WpListStartY + (currentWP.Row-1)*ui.WpAreaBtnHeight + (ui.WpAreaBtnHeight / 2)
						b.HID.Click(game.LeftButton, ui.WpListPositionX, areaBtnY)
					}
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
