package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
)

const (
	wpTabStartX     = 130
	wpTabStartY     = 148
	wpTabSizeX      = 57
	wpListPositionX = 200
	wpListStartY    = 158
	wpAreaBtnHeight = 41
)

func (b Builder) WayPoint(a area.Area) *Factory {
	usedWP := false
	isChild := false
	return NewFactory(func(d data.Data) Action {
		if d.PlayerUnit.Area != a && !usedWP {
			usedWP = true
			return b.useWP(a)
		}

		if d.PlayerUnit.Area != a {
			dstWP := area.WPAddresses[a]
			if isChild {
				return b.traverseNextWP(a, dstWP.LinkedFrom)
			} else {
				b.logger.Info("Waypoint not found (or error occurred) try to autodiscover it", zap.Any("area", a))

				_, found := area.WPAddresses[dstWP.LinkedFrom[0]]
				if found {
					isChild = true
					return b.WayPoint(dstWP.LinkedFrom[0])
				}
			}
		}

		return nil
	})
}

func (b Builder) getLinkedWP(a area.Area) {

}

func (b Builder) traverseNextWP(dst area.Area, areas []area.Area) Action {
	return NewChain(func(d data.Data) (actions []Action) {
		areas = append(areas, dst)
		logAreas := make([]int, len(areas))
		b.logger.Debug("Traversing to next WP...")
		for _, a := range areas {
			logAreas = append(logAreas, int(a))
			switch a {
			case area.MonasteryGate:
				b.logger.Debug("Monastery Gate detected, moving to static coords")
				actions = append(actions,
					b.MoveToCoords(data.Position{X: 15139, Y: 5098}),
					b.MoveToCoords(data.Position{X: 15148, Y: 4978}),
				)
			case area.ArcaneSanctuary:
				b.logger.Debug("Arcane Sanctuary detected, finding the Portal")
				actions = append(actions, BuildStatic(func(d data.Data) []step.Step {
					return []step.Step{
						step.InteractObject(object.ArcaneSanctuaryPortal, func(d data.Data) bool {
							return d.PlayerUnit.Area == area.ArcaneSanctuary
						}),
					}
				}))
			default:
				actions = append(actions,
					b.ch.Buff(),
					b.MoveToArea(a),
				)
			}
		}

		actions = append(actions,
			b.DiscoverWaypoint(),
		)

		b.logger.Debug("Linked areas to traverse", zap.Ints("areas", logAreas))
		return
	})
}

func (b Builder) useWP(a area.Area) Action {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		// We don't need to move
		if d.PlayerUnit.Area == a {
			return
		}

		wpCoords, found := area.WPAddresses[a]
		if !found {
			panic("Area destination is not mapped on WayPoint Action (waypoint.go)")
		}

		for _, o := range d.Objects {
			if o.IsWaypoint() {
				steps = append(steps,
					step.InteractObject(o.Name, func(d data.Data) bool {
						return d.OpenMenus.Waypoint
					}),
					step.SyncStep(func(d data.Data) error {
						actTabX := wpTabStartX + (wpCoords.Tab-1)*wpTabSizeX + (wpTabSizeX / 2)

						areaBtnY := wpListStartY + (wpCoords.Row-1)*wpAreaBtnHeight + (wpAreaBtnHeight / 2)
						hid.MovePointer(actTabX, wpTabStartY)
						helper.Sleep(200)
						hid.Click(hid.LeftButton)
						helper.Sleep(200)
						hid.MovePointer(wpListPositionX, areaBtnY)
						helper.Sleep(200)
						hid.Click(hid.LeftButton)
						helper.Sleep(1000)

						return nil
					}),
				)
			}
		}

		return
	}, IgnoreErrors())
}
