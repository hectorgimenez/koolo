package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
	"strings"
)

const (
	maxGoldPerStashTab = 2500000
	stashGoldBtnX      = 1.2776
	stashGoldBtnY      = 1.357
)

func (b Builder) Stash(forceStash bool) *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if !b.isStashingRequired(data) && !forceStash {
			return
		}

		b.logger.Info("Stashing items...")
		steps = append(steps,
			step.InteractObject("Bank", func(data game.Data) bool {
				return data.OpenMenus.Stash
			}),
			step.SyncStep(func(data game.Data) error {
				stashGold(data)
				b.orderInventoryPotions(data)
				b.stashInventory(data)
				hid.PressKey("esc")
				return nil
			}),
		)

		return
	}, CanBeSkipped())
}

func (b Builder) orderInventoryPotions(data game.Data) {
	for _, i := range data.Items.Inventory {
		if i.IsPotion() {
			if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 {
				continue
			}
			x := int(float32(hid.GameAreaSizeX)/town.InventoryTopLeftX) + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
			y := int(float32(hid.GameAreaSizeY)/town.InventoryTopLeftY) + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
			hid.MovePointer(x, y)
			helper.Sleep(100)
			hid.Click(hid.RightButton)
			helper.Sleep(200)
		}
	}
}

func (b Builder) isStashingRequired(data game.Data) bool {
	for _, i := range data.Items.Inventory {
		if b.shouldStashIt(i) {
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
	}
}

func (b Builder) stashInventory(data game.Data) {
	for _, i := range data.Items.Inventory {
		if !b.shouldStashIt(i) {
			continue
		}
		stashItemAction(i)
		b.logger.Debug(fmt.Sprintf("Item %s [%s] stashed", i.Name, i.Quality))

		for tab := 0; tab < 3; tab++ {
			// TODO: Stash items in other tabs
		}
	}
}

func (b Builder) shouldStashIt(i game.Item) bool {
	if config.Config.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 || i.IsPotion() {
		return false
	}

	for _, pi := range config.Pickit.Items {
		if strings.EqualFold(i.Name, pi.Name) {
			if pi.Quality != "" && !strings.EqualFold(string(i.Quality), pi.Quality) {
				continue
			}

			if pi.Ethereal != nil && i.Ethereal != *pi.Ethereal {
				continue
			}

			stash := true
			for stat, value := range i.Stats {
				for pickitStat, pickitValue := range pi.Stats {
					if strings.EqualFold(string(stat), pickitStat) {
						if value < pickitValue {
							stash = false
							break
						}
					}
				}
			}

			if stash {
				return true
			}
		}
	}

	return false
}

func stashItemAction(i game.Item) bool {
	x := int(float32(hid.GameAreaSizeX)/town.InventoryTopLeftX) + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	y := int(float32(hid.GameAreaSizeY)/town.InventoryTopLeftY) + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
	hid.MovePointer(x, y)
	helper.Sleep(170)
	hid.KeyDown("control")
	helper.Sleep(150)
	hid.Click(hid.LeftButton)
	helper.Sleep(200)
	hid.KeyUp("control")
	helper.Sleep(150)

	// TODO: Check if item has been stored correctly
	return true
}

func clickStashGoldBtn() {
	btnX := int(float32(hid.GameAreaSizeX) / stashGoldBtnX)
	btnY := int(float32(hid.GameAreaSizeY) / stashGoldBtnY)

	hid.MovePointer(btnX, btnY)
	helper.Sleep(170)
	hid.Click(hid.LeftButton)
	helper.Sleep(200)
	hid.PressKey("enter")
	helper.Sleep(500)
}
