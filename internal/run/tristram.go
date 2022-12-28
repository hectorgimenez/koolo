package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type Tristram struct {
	baseRun
}

func (a Tristram) Name() string {
	return "Tristram"
}

func (a Tristram) BuildActions() (actions []action.Action) {
	// Moving to starting point (Stony Field)
	actions = append(actions, a.builder.WayPoint(area.StonyField))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to Tristram portal
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		for _, o := range data.Objects {
			if o.Name == object.CairnStoneAlpha {
				return []step.Step{
					step.MoveTo(o.Position.X, o.Position.Y, true),
				}
			}
		}

		return nil
	}))

	// Clear monsters around the portal
	if config.Config.Game.Tristram.ClearPortal {
		actions = append(actions, action.BuildDynamic(func(data game.Data) ([]step.Step, bool) {
			var closeMonsters []game.Monster
			for _, m := range data.Monsters.Enemies() {
				if pather.DistanceFromMe(data, m.Position.X, m.Position.Y) < 2 {
					closeMonsters = append(closeMonsters, m)
				}
			}
			if len(closeMonsters) == 0 {
				return nil, false
			}

			return a.char.KillMonsterSequence(data, closeMonsters[0].UnitID), true
		}))
	}

	// Enter Tristram portal
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.InteractObject(object.PermanentTownPortal, func(data game.Data) bool {
				return data.PlayerUnit.Area == area.Tristram
			}),
			step.SyncStep(func(data game.Data) error {
				// Add small delay to fetch the monsters
				helper.Sleep(2000)
				return nil
			}),
		}
	}))

	// Clear Tristram
	actions = append(actions, action.BuildDynamic(func(data game.Data) ([]step.Step, bool) {
		monsters := data.Monsters.Enemies()

		// Clear only elite monsters
		if config.Config.Game.Tristram.FocusOnElitePacks {
			monsters = data.Monsters.Enemies(game.MonsterEliteFilter())
		}

		if len(monsters) == 0 {
			return nil, false
		}

		return a.char.KillMonsterSequence(data, monsters[0].UnitID), true

	}, action.CanBeSkipped()))
	return
}
