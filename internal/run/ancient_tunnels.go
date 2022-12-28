package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
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

	// Clear Ancient Tunnels
	actions = append(actions, action.BuildStatic(func(data game.Data) (steps []step.Step) {
		var eliteMonsters []game.Monster
		for _, m := range data.Monsters {
			if m.Type == game.MonsterTypeMinion || m.Type == game.MonsterTypeUnique || m.Type == game.MonsterTypeChampion {
				eliteMonsters = append(eliteMonsters, m)
			}
		}

		sort.Slice(eliteMonsters, func(i, j int) bool {
			distanceI := pather.DistanceFromMe(data, eliteMonsters[i].Position.X, eliteMonsters[i].Position.Y)
			distanceJ := pather.DistanceFromMe(data, eliteMonsters[j].Position.X, eliteMonsters[j].Position.Y)

			return distanceI > distanceJ
		})

		for _, m := range eliteMonsters {
			return a.char.KillMonsterSequence(data, m.UnitID)
		}
		return
	}, action.CanBeSkipped()))

	// Open the chest
	//actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
	//	for _, o := range data.Objects {
	//		if o.Name == object.SparklyChest && o.Selectable {
	//			return []step.Step{
	//				step.MoveTo(o.Position.X, o.Position.Y, true),
	//				step.InteractObject(object.SparklyChest, func(data game.Data) bool {
	//					for _, o := range data.Objects {
	//						if o.Name == object.SparklyChest && o.Selectable {
	//							return false
	//						}
	//					}
	//
	//					return true
	//				}),
	//			}
	//		}
	//	}
	//
	//	return []step.Step{}
	//}, action.CanBeSkipped()))

	return
}
