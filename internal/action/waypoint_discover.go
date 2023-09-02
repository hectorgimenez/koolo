package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
)

func (b *Builder) DiscoverWaypoint() *Factory {
	interacted := false

	return NewFactory(func(d data.Data) Action {
		b.logger.Info("Trying to autodiscover Waypoint for current area", zap.Any("area", d.PlayerUnit.Area))
		if interacted {
			return nil
		}

		for _, o := range d.Objects {
			if o.IsWaypoint() {
				return b.InteractObject(o.Name,
					func(d data.Data) bool {
						return d.OpenMenus.Waypoint
					},
					step.SyncStep(func(d data.Data) error {
						helper.Sleep(1000)
						hid.PressKey("esc")
						interacted = true
						return nil
					}),
				)
			}
		}

		b.logger.Info("Waypoint not found :(", zap.Any("area", d.PlayerUnit.Area))
		return nil
	})
}
