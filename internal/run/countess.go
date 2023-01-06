package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/game/object"
)

type Countess struct {
	baseRun
}

func (c Countess) Name() string {
	return "Countess"
}

func (c Countess) BuildActions() (actions []action.Action) {
	// Moving to starting point (Black Marsh)
	actions = append(actions, c.builder.WayPoint(area.BlackMarsh))

	// Buff
	actions = append(actions, c.char.Buff())

	// Travel to boss level
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.ForgottenTower),
			step.MoveToLevel(area.TowerCellarLevel1),
			step.MoveToLevel(area.TowerCellarLevel2),
			step.MoveToLevel(area.TowerCellarLevel3),
			step.MoveToLevel(area.TowerCellarLevel4),
			step.MoveToLevel(area.TowerCellarLevel5),
		}
	}))

	// Try to move around Countess area
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		for _, o := range data.Objects {
			if o.Name == object.GoodChest {
				return []step.Step{step.MoveTo(o.Position.X, o.Position.Y, true)}
			}
		}

		// Try to teleport over Countess in case we are not able to find the chest position, a bit more risky
		if countess, found := data.Monsters.FindOne(npc.DarkStalker, game.MonsterTypeSuperUnique); found {
			return []step.Step{step.MoveTo(countess.Position.X, countess.Position.Y, true, step.StopAtDistance(15))}
		}

		return []step.Step{}
	}))

	// Kill Countess
	actions = append(actions, c.char.KillCountess())
	return
}
