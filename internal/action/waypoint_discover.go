package action

import (
	"log/slog"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/lxn/win"
)

func (b *Builder) DiscoverWaypoint() *Chain {
	return NewChain(func(d game.Data) []Action {
		b.Logger.Info("Trying to autodiscover Waypoint for current area", slog.String("area", d.PlayerUnit.Area.Area().Name))
		for _, o := range d.Objects {
			if o.IsWaypoint() {
				return []Action{
					b.MoveToCoordsWithMinDistance(o.Position, 2),
					b.InteractObject(o.Name,
						func(d game.Data) bool {
							return d.OpenMenus.Waypoint
						},
						step.SyncStep(func(d game.Data) error {
							b.Logger.Info("Waypoint discovered", slog.String("area", d.PlayerUnit.Area.Area().Name))
							helper.Sleep(500)
							b.HID.PressKey(win.VK_ESCAPE)
							return nil
						}),
					)}
			}
		}

		b.Logger.Info("Waypoint not found :(", slog.String("area", d.PlayerUnit.Area.Area().Name))
		return nil
	})
}
