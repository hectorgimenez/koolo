package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Duriel struct {
	baseRun
}

func (a Duriel) Name() string {
	return string(config.DurielRun)
}

func (a Duriel) BuildActions() (actions []action.Action) {

	// Traveling to the real Tomb
	actions = append(actions, a.travelToTomb())

	//Entering lair and killing Duriel
	actions = append(actions, a.killDuriel())

	return actions
}

func (a Duriel) travelToTomb() action.Action {
	a.logger.Info("Traveling to real tomb")
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		var realTomb area.ID
		for _, tomb := range talRashaTombs {
			_, _, objects, _ := a.Reader.GetCachedMapData(false).NPCsExitsAndObjects(data.Position{}, tomb)
			for _, obj := range objects {
				if obj.Name == object.HoradricOrifice {
					realTomb = tomb
					break
				}
			}
		}

		if realTomb == 0 {
			a.logger.Info("Could not find the real tomb")
			return nil
		}

		return []action.Action{
			a.builder.WayPoint(area.CanyonOfTheMagi),
			a.builder.Buff(),
			a.builder.MoveToArea(realTomb),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				orifice, _ := d.Objects.FindOne(object.HoradricOrifice)
				return orifice.Position, true
			}),
			a.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
		}
	})
}

func (a Duriel) killDuriel() action.Action {
	a.logger.Info("Entering lair and killing Duriel")
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		return []action.Action{
			action.NewChain(func(d game.Data) []action.Action {
				_, found := d.Objects.FindOne(object.DurielsLairPortal)
				if found {
					return []action.Action{a.builder.InteractObject(object.DurielsLairPortal, func(d game.Data) bool {
						return d.PlayerUnit.Area == area.DurielsLair
					})}
				}
				return nil
			}),
			a.char.KillDuriel(),
			a.builder.ItemPickup(true, 30),
		}
	})
}
