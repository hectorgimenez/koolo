package town

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

const (
	wpTabStartX     = 21.719
	wpTabStartY     = 7.928
	wpListStartX    = 5.37
	wpListStartY    = 6.852
	wpTabSize       = 69
	wpAreaBtnHeight = 49
)

type Manager struct {
	towns       map[game.Area]Town
	cfg         config.Config
	pf          helper.PathFinder
	shopManager ShopManager
}

func NewTownManager(cfg config.Config, pf helper.PathFinder, shopManager ShopManager) Manager {
	return Manager{
		towns: map[game.Area]Town{
			game.AreaRogueEncampment: A1{},
			game.AreaKurastDocks:     A3{},
			game.AreaHarrogath:       A5{},
		},
		cfg:         cfg,
		pf:          pf,
		shopManager: shopManager,
	}
}

func (tm Manager) BuyPotionsAndTPs(area game.Area, buyTPs bool) {
	t := tm.getTownByArea(area)
	tm.pf.InteractToNPC(t.RefillNPC())
	tm.openTradeMenu()
	tm.shopManager.buyPotsAndTPs(buyTPs)

}

func (tm Manager) Heal(area game.Area) {
	t := tm.getTownByArea(area)
	tm.pf.InteractToNPC(t.RefillNPC())
}

func (tm Manager) Repair(area game.Area) {
	t := tm.getTownByArea(area)
	tm.pf.InteractToNPC(t.RepairNPC())
	tm.openTradeMenu()
	action.Run(
		action.NewMouseDisplacement(int(float32(hid.GameAreaSizeX)/3.52), int(float32(hid.GameAreaSizeY)/1.37), time.Millisecond*850),
		action.NewMouseClick(hid.LeftButton, time.Millisecond*1300),
	)
	tm.closeTradeMenu()
}

func (tm Manager) ReviveMerc(area game.Area) {
	t := tm.getTownByArea(area)
	tm.pf.InteractToNPC(t.MercContractorNPC())
	tm.openTradeMenu()
	tm.closeTradeMenu()
}

func (tm Manager) WPTo(act int, area int) error {
	for _, o := range game.Status().Objects {
		if o.IsWaypoint() {
			err := tm.pf.InteractToObject(o, func(data game.Data) bool {
				return data.OpenMenus.Waypoint
			})
			if err != nil {
				return err
			}

			currentArea := game.Status().Area
			actTabX := int(float32(hid.GameAreaSizeX)/wpTabStartX) + (act-1)*wpTabSize + (wpTabSize / 2)
			actTabY := int(float32(hid.GameAreaSizeY) / wpTabStartY)

			areaBtnX := int(float32(hid.GameAreaSizeX) / wpListStartX)
			areaBtnY := int(float32(hid.GameAreaSizeY)/wpListStartY) + (area-1)*wpAreaBtnHeight + (wpAreaBtnHeight / 2)
			action.Run(
				action.NewMouseDisplacement(actTabX, actTabY, time.Millisecond*200),
				action.NewMouseClick(hid.LeftButton, time.Millisecond*200),
				action.NewMouseDisplacement(areaBtnX, areaBtnY, time.Millisecond*200),
				action.NewMouseClick(hid.LeftButton, time.Second*1),
			)

			for i := 0; i < 10; i++ {
				if game.Status().Area != currentArea {
					// Give some time to load the area
					time.Sleep(time.Second * 4)
					return nil
				}
				time.Sleep(time.Second * 1)
			}
		}
	}

	return errors.New("error changing area zone")
}

func (tm Manager) openTradeMenu() {
	if game.Status().OpenMenus.NPCInteract {
		action.Run(action.NewKeyPress("down", time.Millisecond*150), action.NewKeyPress("enter", time.Millisecond*500))
	}
}

func (tm Manager) closeTradeMenu() {
	if game.Status().OpenMenus.NPCInteract {
		action.Run(action.NewKeyPress("esc", time.Millisecond*150))
	}
}

func (tm Manager) getTownByArea(area game.Area) Town {
	return tm.towns[area]
}
