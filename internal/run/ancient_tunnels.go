package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/object"
)

type AncientTunnels struct {
	baseRun
}

func (a AncientTunnels) Name() string {
	return "AncientTunnels"
}

func (a AncientTunnels) BuildActions() (actions []action.Action) {
	// Moving to starting point (Lost City)
	actions = append(actions, a.builder.WayPoint(area.LostCity))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to ancient tunnels
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.AncientTunnels),
		}
	}))

	// Open the chest
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		for _, o := range data.Objects {
			if o.Name == object.SparklyChest && o.IsChest {
				return []step.Step{
					step.MoveTo(o.Position.X, o.Position.Y, true),
					step.InteractObject(object.SparklyChest, func(data game.Data) bool {
						for _, o := range data.Objects {
							if o.Name == object.SparklyChest && o.IsChest {
								return false
							}
						}

						return true
					}),
				}
			}
		}

		return []step.Step{}
	}, action.CanBeSkipped()))

	// Clear Ancient Tunnels
	actions = append(actions, a.char.ClearAncientTunnels())
	return
}
