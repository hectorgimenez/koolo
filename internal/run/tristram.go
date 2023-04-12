package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
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
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		for _, o := range d.Objects {
			if o.Name == object.CairnStoneAlpha {
				return []step.Step{
					step.MoveTo(o.Position.X, o.Position.Y, true),
					step.SyncStep(func(d data.Data) error {
						helper.Sleep(1000)
						return nil
					}),
				}
			}
		}

		return nil
	}))

	// Clear monsters around the portal
	if config.Config.Game.Tristram.ClearPortal {
		actions = append(actions, a.builder.ClearAreaAroundPlayer(5))
	}

	// Enter Tristram portal
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
				return d.PlayerUnit.Area == area.Tristram
			}),
			step.SyncStep(func(d data.Data) error {
				// Add small delay to fetch the monsters
				helper.Sleep(2000)
				return nil
			}),
		}
	}))

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear Tristram
	actions = append(actions, a.builder.ClearArea(false, data.MonsterAnyFilter()))
	//actions = append(actions, a.char.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
	//	monsters := d.Monsters.Enemies()
	//
	//	// Clear only elite monsters
	//	if config.Config.Game.Tristram.FocusOnElitePacks {
	//		monsters = d.Monsters.Enemies(data.MonsterEliteFilter())
	//	}
	//
	//	if len(monsters) == 0 {
	//		return 0, false
	//	}
	//
	//	return monsters[0].UnitID, true
	//}, nil))

	return
}
