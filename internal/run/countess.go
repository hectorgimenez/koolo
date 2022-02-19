package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	countessStartingPositionX = 12554
	countessStartingPositionY = 11014
)

type Countess struct {
	BaseRun
}

func NewCountess(run BaseRun) Countess {
	return Countess{
		BaseRun: run,
	}
}

func (c Countess) Name() string {
	return "Countess"
}

func (c Countess) BuildActions(data game.Data) (actions []action.Action) {
	// Moving to starting point (Black Marsh)
	if data.Area != game.AreaBlackMarsh {
		actions = append(actions, c.builder.WayPoint(game.AreaBlackMarsh))
	}

	// Buff
	actions = append(actions, c.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(game.AreaForgottenTower),
			step.MoveToLevel(game.AreaTowerCellarLevel1),
			step.MoveToLevel(game.AreaTowerCellarLevel2),
			step.MoveToLevel(game.AreaTowerCellarLevel3),
			step.MoveToLevel(game.AreaTowerCellarLevel4),
			step.MoveToLevel(game.AreaTowerCellarLevel5),
			step.MoveTo(countessStartingPositionX, countessStartingPositionY, true),
		}
	}))

	// Kill Countess
	actions = append(actions, c.char.KillCountess())
	return
}
