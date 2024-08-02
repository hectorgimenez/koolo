package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

type Endugu struct {
	baseRun
}

func (e Endugu) Name() string {
	return string(config.EnduguRun)
}

func (e Endugu) BuildActions() []action.Action {
	return []action.Action{
		e.builder.WayPoint(area.FlayerJungle),
		e.builder.MoveToArea(area.FlayerDungeonLevel1),
		e.builder.MoveToArea(area.FlayerDungeonLevel2),
		e.builder.MoveToArea(area.FlayerDungeonLevel3),
		e.moveToKhalimChest(),
		e.builder.ClearAreaAroundPlayer(15, data.MonsterEliteFilter()),
		action.NewStepChain(func(d game.Data) []step.Step {
			return []step.Step{step.Wait(3 * time.Second)}
		}),
		e.openKhalimChest(),
		e.builder.ItemPickup(false, 10),
	}
}

func (e Endugu) moveToKhalimChest() action.Action {
	return action.NewStepChain(func(d game.Data) []step.Step {
		for _, o := range d.Objects {
			if o.Name == object.KhalimChest2 {
				return []step.Step{step.MoveTo(o.Position, step.StopAtDistance(10))}
			}
		}
		return nil
	})
}

func (e Endugu) openKhalimChest() action.Action {
	return e.builder.InteractObject(object.KhalimChest2, func(d game.Data) bool {
		for _, obj := range d.Objects {
			if obj.Name == object.KhalimChest2 && !obj.Selectable {
				return true
			}
		}
		return false
	})
}
