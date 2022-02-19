package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
	"time"
)

const (
	maxGoldPerStashTab = 2500000
	stashGoldBtnX      = 1.2776
	stashGoldBtnY      = 1.357

	inventoryTopLeftX = 1.494
	inventoryTopLeftY = 2.071
)

func (b Builder) Stash() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if !b.isStashingRequired(data) {
			return
		}

		steps = append(steps,
			step.NewInteractObject("Bank", func(data game.Data) bool {
				return data.OpenMenus.Stash
			}),
			step.NewSyncAction(func(data game.Data) error {
				stashGold(data)
				b.stashInventory()
				hid.PressKey("esc")
				return nil
			}),
		)

		return
	})
}

func (b Builder) isStashingRequired(data game.Data) bool {
	for _, i := range data.Items.Inventory {
		if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 1 {
			return true
		}
	}

	return false
}

func stashGold(d game.Data) {
	// TODO: Handle multiple tabs
	if d.PlayerUnit.Stats[game.StatGold] == 0 {
		return
	}

	if d.PlayerUnit.Stats[game.StatStashGold] < maxGoldPerStashTab {
		clickStashGoldBtn()
		d = game.Status()
		if d.PlayerUnit.Stats[game.StatGold] == 0 {
			return
		}
	}
}

func (b Builder) stashInventory() {
	for _, i := range game.Status().Items.Inventory {
		if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 {
			continue
		}
		stashItemAction(i)

		for tab := 0; tab < 3; tab++ {
			// TODO: Stash items in other tabs
		}
	}
}

func stashItemAction(i game.Item) bool {
	x := int(float32(hid.GameAreaSizeX)/inventoryTopLeftX) + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	y := int(float32(hid.GameAreaSizeY)/inventoryTopLeftY) + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
	hid.MovePointer(x, y)
	time.Sleep(time.Millisecond * 170)
	hid.KeyDown("control")
	time.Sleep(time.Millisecond * 150)
	hid.Click(hid.LeftButton)
	time.Sleep(time.Millisecond * 200)
	hid.KeyUp("control")
	time.Sleep(time.Millisecond * 150)

	// TODO: Check if item has been stored correctly
	return true
}

func clickStashGoldBtn() {
	btnX := int(float32(hid.GameAreaSizeX) / stashGoldBtnX)
	btnY := int(float32(hid.GameAreaSizeY) / stashGoldBtnY)

	hid.MovePointer(btnX, btnY)
	time.Sleep(time.Millisecond * 170)
	hid.Click(hid.LeftButton)
	time.Sleep(time.Millisecond * 200)
	hid.PressKey("enter")
	time.Sleep(time.Millisecond * 500)
}
