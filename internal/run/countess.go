package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Countess struct {
	baseRun
}

func (c Countess) Name() string {
	return "Countess"
}

func (c Countess) BuildActions() (actions []action.Action) {
	// Moving to starting point (Black Marsh)
	actions = append(actions, c.builder.WayPoint(game.AreaBlackMarsh))

	// Buff
	actions = append(actions, c.char.Buff())

	// Travel to boss level
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(game.AreaForgottenTower),
			step.MoveToLevel(game.AreaTowerCellarLevel1),
			step.MoveToLevel(game.AreaTowerCellarLevel2),
			step.MoveToLevel(game.AreaTowerCellarLevel3),
			step.MoveToLevel(game.AreaTowerCellarLevel4),
			step.MoveToLevel(game.AreaTowerCellarLevel5),
		}
	}))

	// Try to move around Countess area
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		for _, o := range data.Objects {
			if o.Name == "GoodChest" {
				steps = append(steps, step.MoveTo(o.Position.X, o.Position.Y, true))
			}
		}
		return
	}))

	// Let's teleport over Countess
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		countess, found := data.Monsters.FindOne(game.Countess)
		if !found {
			return
		}

		steps = append(steps, step.MoveTo(countess.Position.X, countess.Position.Y, true))
		return
	}))

	// Kill Countess
	actions = append(actions, c.char.KillCountess())
	return
}
