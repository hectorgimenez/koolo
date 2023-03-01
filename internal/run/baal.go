package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	baalThronePositionX = 15095
	baalThronePositionY = 5042
)

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
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.TheWorldStoneKeepLevel3),
			step.MoveToLevel(area.ThroneOfDestruction),
		}
	}))

	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{step.MoveTo(baalThronePositionX, baalThronePositionY, true)}
	}))

	// Kill monsters inside Baal throne
	actions = append(actions, s.builder.ClearAreaAroundPlayer(50))

	// Let's move to a safe area and open the portal in companion mode
	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
			return []step.Step{step.MoveTo(15116, 5071, true), step.OpenPortal()}
		}))
	}

	// Come back to previous position
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{step.MoveTo(baalThronePositionX, baalThronePositionY, true)}
	}))

	lastWave := false
	actions = append(actions, action.NewFactory(func(data game.Data) action.Action {
		if !lastWave {
			if _, found := data.Monsters.FindOne(npc.BaalsMinion, game.MonsterTypeMinion); found {
				lastWave = true
			}

			enemies := false
			for _, e := range data.Monsters.Enemies() {
				d := pather.DistanceFromPoint(game.Position{
					X: baalThronePositionX,
					Y: baalThronePositionY,
				}, e.Position)
				if d > 50 {
					enemies = true
				}
			}
			if !enemies {
				d := pather.DistanceFromMe(data, game.Position{
					X: baalThronePositionX,
					Y: baalThronePositionY,
				})
				if d > 5 {
					return action.BuildStatic(func(data game.Data) []step.Step {
						return []step.Step{step.MoveTo(baalThronePositionX, baalThronePositionY, true)}
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
