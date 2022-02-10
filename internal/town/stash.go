package town

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

const (
	maxGoldPerStashTab = 2500000
	stashGoldBtnX      = 1.2776
	stashGoldBtnY      = 1.357
)

func (tm Manager) stashAllItems() {
	tm.stashGold()
}

func (tm Manager) stashGold() {
	d := data.Status
	if d.PlayerUnit.Stats[data.StatGold] == 0 {
		return
	}

	if d.PlayerUnit.Stats[data.StatStashGold] < maxGoldPerStashTab {
		stashGoldAction()
		if d.PlayerUnit.Stats[data.StatGold] == 0 {
			return
		}
	}

	// We can not fetch shared stash status, so we don't know gold amount, let's try to stash on all of them
	for i := 0; i < 3; i++ {
		// TODO: Stash gold in other tabs
	}
}

func stashGoldAction() {
	btnX := int(float32(hid.GameAreaSizeX) / stashGoldBtnX)
	btnY := int(float32(hid.GameAreaSizeY) / stashGoldBtnY)
	action.Run(
		action.NewMouseDisplacement(btnX, btnY, time.Millisecond*170),
		action.NewMouseClick(hid.LeftButton, time.Millisecond*200),
		action.NewKeyPress("enter", time.Millisecond*500),
	)
}
