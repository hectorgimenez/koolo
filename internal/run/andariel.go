package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	andarielStartingPositionX = 22561
	andarielStartingPositionY = 9553
)

type Andariel struct {
	baseRun
}

func (a Andariel) Name() string {
	return "Andariel"
}

func (a Andariel) BuildActions() (actions []action.Action) {
	// Moving to starting point (Catacombs Level 2)
	actions = append(actions, a.builder.WayPoint(game.AreaCatacombsLevel2))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(game.AreaCatacombsLevel3),
			step.MoveToLevel(game.AreaCatacombsLevel4),
			step.MoveTo(andarielStartingPositionX, andarielStartingPositionY, true),
		}
	}))

	// Kill Andariel
	actions = append(actions, a.char.KillAndariel())
	return
}
