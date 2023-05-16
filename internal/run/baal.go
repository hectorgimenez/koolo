package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/pather"
)

var baalThronePosition = data.Position{
	X: 15095,
	Y: 5042,
}

type Baal struct {
	baseRun
}

func (s Baal) Name() string {
	return "Baal"
}

func (s Baal) BuildActions() (actions []action.Action) {
	// Moving to starting point (The World StoneKeep Level 2)
	actions = append(actions, s.builder.WayPoint(area.TheWorldStoneKeepLevel2))

	// Buff
	actions = append(actions, s.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.TheWorldStoneKeepLevel3),
			step.MoveToLevel(area.ThroneOfDestruction),
		}
	}))

	actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
		return []step.Step{step.MoveTo(baalThronePosition)}
	}))

	// Kill monsters inside Baal throne
	actions = append(actions, s.builder.ClearAreaAroundPlayer(50))

	// Let's move to a safe area and open the portal in companion mode
	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
			return []step.Step{step.MoveTo(data.Position{
				X: 15116,
				Y: 5071,
			}), step.OpenPortal()}
		}))
	}

	// Come back to previous position
	actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
		return []step.Step{step.MoveTo(baalThronePosition)}
	}))

	lastWave := false
	actions = append(actions, action.NewFactory(func(d data.Data) action.Action {
		if !lastWave {
			if _, found := d.Monsters.FindOne(npc.BaalsMinion, data.MonsterTypeMinion); found {
				lastWave = true
			}

			enemies := false
			for _, e := range d.Monsters.Enemies() {
				dist := pather.DistanceFromPoint(baalThronePosition, e.Position)
				if dist > 50 {
					enemies = true
				}
			}
			if !enemies {
				dist := pather.DistanceFromMe(d, baalThronePosition)
				if dist > 5 {
					return action.BuildStatic(func(d data.Data) []step.Step {
						return []step.Step{step.MoveTo(baalThronePosition)}
					})
				}
			}

			return s.builder.ClearAreaAroundPlayer(50)
		}

		return nil
	}))

	actions = append(actions, s.builder.ItemPickup(false, 30))

	return
}
