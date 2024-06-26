package run

import (
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Rushing struct {
	baseRun
}

func (a Rushing) Name() string {
	return string(config.RushingRun)
}

func (a Rushing) BuildActions() []action.Action {
	return []action.Action{
		a.rushAct1(),
		a.rushAct2(),
		a.rushAct3(),
		a.rushAct4(),
		a.rushAct5(),
	}
}

func (a Rushing) waitForParty(d game.Data) action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		for {
			data := a.Container.Reader.GetData(false)
			for _, c := range data.Roster {
				if c.Area == d.PlayerUnit.Area {
					return nil
				}
			}
			helper.Sleep(1000) 
		}
	})
}

//Original waitForPartyInCatacombsLevel4 function
//func (a Rushing) waitForPartyInCatacombsLevel4() action.Action {
//	return action.NewChain(func(d game.Data) []action.Action {
//		for {
//			data := a.Container.Reader.GetData(false)
//			for _, c := range data.Roster {
//				if c.Area == area.CatacombsLevel4 {
//					return nil
//				}
//			}
//			helper.Sleep(1000) 
//		}
//	})
//}