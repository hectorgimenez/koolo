package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (a Rushing) rushAct1() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.RogueEncampment {
			return nil
		}

		running = true
		
		if a.CharacterCfg.Game.Rushing.GiveWPs {
			return []action.Action{
				a.builder.VendorRefill(true, false),
				a.GiveAct1WPs(),
				a.killAandarielQuest(),
			}
		}

		return []action.Action{
			a.builder.VendorRefill(true, false),
			a.killAandarielQuest(),
		}
	})
}

func (a Rushing) GiveAct1WPs() action.Action {
	areas := []area.ID{
		area.StonyField,
		area.DarkWood,
		area.BlackMarsh,
		area.InnerCloister,
		area.OuterCloister,
		area.CatacombsLevel2,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		actions := []action.Action{}

		for _, areaID := range areas {
			actions = append(actions,
				a.builder.WayPoint(areaID),
				a.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
				a.builder.OpenTP(),
				a.builder.Wait(time.Second*5),
			)
		}

		return actions
	})
}

func (a Rushing) killAandarielQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.CatacombsLevel2),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.CatacombsLevel3),
			a.builder.MoveToArea(area.CatacombsLevel4),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
//			a.waitForParty(d),
			a.builder.MoveToCoords(andarielStartingPosition),
			a.char.KillAndariel(),
			a.builder.ReturnTown(),
			a.builder.WayPoint(area.LutGholein),			
		}
	})
}

