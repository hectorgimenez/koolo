package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	stat2 "github.com/hectorgimenez/koolo/internal/event/stat"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

const (
	maxGoldPerStashTab = 2500000
	stashGoldBtnX      = 966
	stashGoldBtnY      = 526
)

func (b Builder) Stash(forceStash bool) *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		if !b.isStashingRequired(data, forceStash) {
			return
		}

		b.logger.Info("Stashing items...")

		switch data.PlayerUnit.Area {
		case area.KurastDocks:
			steps = append(steps, step.MoveTo(5146, 5067, false))
		case area.LutGholein:
			steps = append(steps, step.MoveTo(5130, 5086, false))
		}

		steps = append(steps,
			step.InteractObject(object.Bank, func(data game.Data) bool {
				return data.OpenMenus.Stash
			}),
			step.SyncStep(func(data game.Data) error {
				b.stashGold(data)
				b.orderInventoryPotions(data)
				b.stashInventory(data, forceStash)
				hid.PressKey("esc")
				return nil
			}),
		)

		return
	}, Resettable(), CanBeSkipped())
}

func (b Builder) orderInventoryPotions(data game.Data) {
	for _, i := range data.Items.Inventory {
		if i.IsPotion() {
			if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 {
				continue
			}
			x := town.InventoryTopLeftX + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
			y := town.InventoryTopLeftY + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
			hid.MovePointer(x, y)
			helper.Sleep(100)
			hid.Click(hid.RightButton)
			helper.Sleep(200)
		}
	}
}

func (b Builder) isStashingRequired(data game.Data, forceStash bool) bool {
	for _, i := range data.Items.Inventory {
		if b.shouldStashIt(i, forceStash) {
			return true
		}
	}

	return false
}

func (b Builder) stashGold(d game.Data) {
	gold, found := d.PlayerUnit.Stats[stat.Gold]
	if !found || gold == 0 {
		return
	}

	if d.PlayerUnit.Stats[stat.StashGold] < maxGoldPerStashTab {
		switchTab(1)
		clickStashGoldBtn()
		helper.Sleep(200)
	}

	for i := 2; i < 5; i++ {
		data := b.gr.GetData(false)
		gold, found = data.PlayerUnit.Stats[stat.Gold]
		if !found || gold == 0 {
			return
		}

		switchTab(i)
		clickStashGoldBtn()
	}
	b.logger.Info("All stash tabs are full of gold :D")
}

func (b Builder) stashInventory(data game.Data, forceStash bool) {
	currentTab := 1
	switchTab(currentTab)

	for _, i := range data.Items.Inventory {
		if !b.shouldStashIt(i, forceStash) {
			continue
		}
		for currentTab < 5 {
			if b.stashItemAction(i, forceStash) {
				b.logger.Debug(fmt.Sprintf("Item %s [%s] stashed", i.Name, i.Quality))
				break
			}
			if currentTab == 5 {
				// TODO: Stop the bot, stash is full
			}
			b.logger.Debug(fmt.Sprintf("Tab %d is full, switching to next one", currentTab))
			currentTab++
			switchTab(currentTab)
		}
	}
}

func (b Builder) shouldStashIt(i game.Item, forceStash bool) bool {
	if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 || i.IsPotion() {
		return false
	}

	return forceStash || i.PickupPass(true)
}

func (b Builder) stashItemAction(i game.Item, forceStash bool) bool {
	x := town.InventoryTopLeftX + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	y := town.InventoryTopLeftY + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
	hid.MovePointer(x, y)
	helper.Sleep(170)
	screenshot := helper.Screenshot()
	hid.KeyDown("control")
	helper.Sleep(300)
	hid.Click(hid.LeftButton)
	helper.Sleep(200)
	hid.KeyUp("control")
	helper.Sleep(300)

	data := b.gr.GetData(false)
	for _, it := range data.Items.Inventory {
		if it.UnitID == i.UnitID {
			return false
		}
	}

	// Don't log items that we already have in inventory during first run
	if !forceStash {
		stat2.ItemStashed(i, screenshot)
	}
	return true
}

func clickStashGoldBtn() {
	hid.MovePointer(stashGoldBtnX, stashGoldBtnY)
	helper.Sleep(170)
	hid.Click(hid.LeftButton)
	helper.Sleep(200)
	hid.PressKey("enter")
	helper.Sleep(500)
}

func switchTab(tab int) {
	x := 107
	y := 128
	tabSize := 82
	x = x + tabSize*tab - tabSize/2

	hid.MovePointer(x, y)
	helper.Sleep(400)
	hid.Click(hid.LeftButton)
}
