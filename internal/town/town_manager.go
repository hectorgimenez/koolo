package town

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type Manager struct {
	towns      map[data.Area]Town
	pf         helper.PathFinder
	sm         ShopManager
	dr         data.DataRepository
	actionChan chan<- action.Action
}

func NewTownManager(repository data.DataRepository, pf helper.PathFinder, actionChan chan<- action.Action) Manager {
	return Manager{
		towns: map[data.Area]Town{
			data.AreaHarrogath: A5{},
		},
		pf:         pf,
		dr:         repository,
		actionChan: actionChan,
	}
}

func (tm Manager) BuyPotionsAndTPs(area data.Area) {
	t := tm.getTownByArea(area)
	tm.pf.InteractToNPC(t.RefillNPC())
	tm.openTradeMenu()
	tm.sm.buyPotsAndTPs()
}

func (tm Manager) Repair(area data.Area) {
	t := tm.getTownByArea(area)
	tm.pf.InteractToNPC(t.RepairNPC())
	tm.openTradeMenu()
	tm.actionChan <- action.NewAction(
		action.PriorityNormal,
		action.NewMouseDisplacement(time.Millisecond*850, int(float32(hid.GameAreaSizeX)/3.52), int(float32(hid.GameAreaSizeY)/1.37)),
		action.NewMouseClick(time.Millisecond*1300, hid.LeftButton),
	)
}

func (tm Manager) ReviveMerc(area data.Area) {
	t := tm.getTownByArea(area)
	tm.pf.InteractToNPC(t.MercContractorNPC())

}

func (tm Manager) openTradeMenu() {
	d := tm.dr.GameData()
	if d.OpenMenus.NPCInteract {
		tm.actionChan <- action.NewAction(
			action.PriorityNormal,
			action.NewKeyPress("down", time.Millisecond*760),
			action.NewKeyPress("enter", time.Second),
		)
	}

}

func (tm Manager) getTownByArea(area data.Area) Town {
	return tm.towns[area]
}
