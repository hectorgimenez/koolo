package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
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

					hid.MovePointer(actTabX, wpTabStartY)
					helper.Sleep(200)
					hid.Click(hid.LeftButton)
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
		sc := helper.Screenshot()

		nextAvailableWP := area.WPAddresses[a]
		traverseAreas := make([]area.Area, 0)
		for {
			//tm := b.tf.FindInArea("ui_discovered_wp", sc, wpTabStartX, wpListStartY+(wpAreaBtnHeight*(nextAvailableWP.Row-1)), wpTabStartX+60, wpListStartY+(wpAreaBtnHeight*nextAvailableWP.Row))
			tm := b.tf.MatchColorInArea(sc, gocv.NewScalar(3, 230, 230, 0), gocv.NewScalar(140, 255, 255, 0), wpTabStartX, wpListStartY+(wpAreaBtnHeight*(nextAvailableWP.Row-1)), wpTabStartX+9, wpListStartY+(wpAreaBtnHeight*nextAvailableWP.Row))
			if !tm.Found && nextAvailableWP.Row != 1 {
				traverseAreas = append(nextAvailableWP.LinkedFrom, traverseAreas...)
				nextAvailableWP = area.WPAddresses[nextAvailableWP.LinkedFrom[0]]
				continue
			}
			break
		}

		// First use the previous available waypoint that we have discovered
		actions = append(actions, NewStepChain(func(d data.Data) []step.Step {
			return []step.Step{
				step.SyncStep(func(d data.Data) error {
					areaBtnY := wpListStartY + (nextAvailableWP.Row-1)*wpAreaBtnHeight + (wpAreaBtnHeight / 2)
					hid.MovePointer(wpListPositionX, areaBtnY)
					helper.Sleep(200)
					hid.Click(hid.LeftButton)
					helper.Sleep(1000)

					return nil
				}),
			}
		}))

		traverseAreas = append(traverseAreas, a)

		// Next keep traversing all the areas from the previous available waypoint until we reach the destination, trying to discover WPs during the way
		b.logger.Info("Traversing areas to reach destination", zap.Any("areas", traverseAreas))

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
