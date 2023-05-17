package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b Builder) DiscoverWaypoint() *Factory {
	interacted := false

	return NewFactory(func(d data.Data) Action {
		if interacted {
			return nil
		}

		for _, o := range d.Objects {
			if o.IsWaypoint() {
				if pather.DistanceFromMe(d, o.Position) < 15 {
					return BuildStatic(func(d data.Data) []step.Step {
						return []step.Step{
							step.MoveTo(o.Position),
							step.InteractObject(o.Name, func(d data.Data) bool {
								return d.OpenMenus.Waypoint
							}),
							step.SyncStep(func(d data.Data) error {
								helper.Sleep(1000)
								hid.PressKey("esc")
								interacted = true
								return nil
							}),
						}
					})
				}

				return b.MoveAndKill(func(d data.Data) (data.Position, bool) {
					for _, o := range d.Objects {
						if o.IsWaypoint() {
							return o.Position, true
						}
					}

					return data.Position{}, false
				})
			}
		}

		return nil
	})
}
