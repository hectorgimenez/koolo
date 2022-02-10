package town

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

const (
	maxGoldPerStashTab = 2500000

	stashGoldBtnX = 1.2776
	stashGoldBtnY = 1.357

	inventoryTopLeftX = 1.494
	inventoryTopLeftY = 2.071
)

func (tm Manager) stashAllItems() {
	tm.stashGold()
	tm.stashInventory()
}

func (tm Manager) stashGold() {
	d := game.Status()
	if d.PlayerUnit.Stats[game.StatGold] == 0 {
		return
	}

	if d.PlayerUnit.Stats[game.StatStashGold] < maxGoldPerStashTab {
		stashGoldAction()
		d = game.Status()
		if d.PlayerUnit.Stats[game.StatGold] == 0 {
			return
		}
	}

	// We can not fetch shared stash status, so we don't know gold amount, let's try to stash on all of them
	for i := 0; i < 3; i++ {
		// TODO: Stash gold in other tabs
	}
}

func (tm Manager) stashInventory() {
	for _, i := range game.Status().Items.Inventory {
		if tm.cfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 {
			continue
		}
		stashItemAction(i)

		for tab := 0; tab < 3; tab++ {
			// TODO: Stash items in other tabs
		}
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

func stashItemAction(i game.Item) bool {
	spaceX := int(float32(hid.GameAreaSizeX)/inventoryTopLeftX) + i.Position.X*itemBoxSize + (itemBoxSize / 2)
	spaceY := int(float32(hid.GameAreaSizeY)/inventoryTopLeftY) + i.Position.Y*itemBoxSize + (itemBoxSize / 2)
	action.Run(
		action.NewMouseDisplacement(spaceX, spaceY, time.Millisecond*170),
		action.NewKeyDown("control", time.Millisecond*150),
		action.NewMouseClick(hid.LeftButton, time.Millisecond*200),
		action.NewKeyUp("control", time.Millisecond*150),
	)

	// TODO: Check if item has been stored correctly
	return true
}
