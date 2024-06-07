package action

import (
	"fmt"
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
)

const (
	maxGoldPerStashTab = 2500000
)

func (b *Builder) Stash(forceStash bool) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		b.Logger.Debug("Checking for items to stash...")
		if !b.isStashingRequired(d, forceStash) {
			return
		}

		b.Logger.Info("Stashing items...")

		switch d.PlayerUnit.Area {
		case area.KurastDocks:
			actions = append(actions, b.MoveToCoords(data.Position{X: 5146, Y: 5067}))
		case area.LutGholein:
			actions = append(actions, b.MoveToCoords(data.Position{X: 5130, Y: 5086}))
		}

		return append(actions,
			b.InteractObject(object.Bank,
				func(d game.Data) bool {
					return d.OpenMenus.Stash
				},
				step.SyncStep(func(d game.Data) error {
					b.stashGold(d)
					b.orderInventoryPotions(d)
					b.stashInventory(d, forceStash)
					b.HID.PressKey(win.VK_ESCAPE)
					return nil
				}),
			),
		)
	})
}

func (b *Builder) orderInventoryPotions(d game.Data) {
	for _, i := range d.Inventory.ByLocation(item.LocationInventory) {
		if i.IsPotion() {
			if d.CharacterCfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 {
				continue
			}

			screenPos := b.UIManager.GetScreenCoordsForItem(i)
			helper.Sleep(100)
			b.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
			helper.Sleep(200)
		}
	}
}

func (b *Builder) isStashingRequired(d game.Data, firstRun bool) bool {
	for _, i := range d.Inventory.ByLocation(item.LocationInventory) {
		if b.shouldStashIt(d, i, firstRun) {
			return true
		}
	}

	isStashFull := true
	for _, goldInStash := range d.Inventory.StashedGold {
		if goldInStash < maxGoldPerStashTab {
			isStashFull = false
		}
	}

	if d.Inventory.Gold > d.PlayerUnit.MaxGold()/3 && !isStashFull {
		return true
	}

	return false
}

func (b *Builder) stashGold(d game.Data) {
	if d.Inventory.Gold == 0 {
		return
	}

	b.Logger.Info("Stashing gold...", slog.Int("gold", d.Inventory.Gold))

	for tab, goldInStash := range d.Inventory.StashedGold {
		d = b.Reader.GetData(false)
		if d.Inventory.Gold == 0 {
			return
		}

		if goldInStash < maxGoldPerStashTab {
			b.switchTab(tab + 1)
			b.clickStashGoldBtn()
			helper.Sleep(500)
		}
	}

	b.Logger.Info("All stash tabs are full of gold :D")
}

func (b *Builder) stashInventory(d game.Data, firstRun bool) {
	currentTab := 1
	if b.CharacterCfg.Character.StashToShared {
		currentTab = 2
	}
	b.switchTab(currentTab)

	for _, i := range d.Inventory.ByLocation(item.LocationInventory) {
		if !b.shouldStashIt(d, i, firstRun) {
			continue
		}
		for currentTab < 5 {
			if b.stashItemAction(i, firstRun) {
				r, res := b.CharacterCfg.Runtime.Rules.EvaluateAll(i)

				if res != nip.RuleResultFullMatch && firstRun {
					b.Logger.Info(
						fmt.Sprintf("Item %s [%s] stashed because it was found in the inventory during the first run.", i.Desc().Name, i.Quality.ToString()),
					)
					break
				}

				b.Logger.Info(
					fmt.Sprintf("Item %s [%s] stashed", i.Desc().Name, i.Quality.ToString()),
					slog.String("nipFile", fmt.Sprintf("%s:%d", r.Filename, r.LineNumber)),
					slog.String("rawRule", r.RawLine),
				)
				break
			}
			if currentTab == 5 {
				// TODO: Stop the bot, stash is full
			}
			b.Logger.Debug(fmt.Sprintf("Tab %d is full, switching to next one", currentTab))
			currentTab++
			b.switchTab(currentTab)
		}
	}
}

func (b *Builder) shouldStashIt(d game.Data, i data.Item, firstRun bool) bool {
	// Don't stash items from quests during leveling process, it makes things easier to track
	if _, isLevelingChar := b.ch.(LevelingCharacter); isLevelingChar && i.IsFromQuest() {
		return false
	}

	// Don't stash the Tomes, keys and WirtsLeg
	if i.Name == item.TomeOfTownPortal || i.Name == item.TomeOfIdentify || i.Name == item.Key || i.Name == "WirtsLeg" {
		return false
	}

	if i.Position.Y >= len(b.CharacterCfg.Inventory.InventoryLock) || i.Position.X >= len(b.CharacterCfg.Inventory.InventoryLock[0]) {
		return false
	}

	if i.Location.LocationType == item.LocationInventory && b.CharacterCfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 || i.IsPotion() {
		return false
	}

	// Let's stash everything during first run, we don't want to sell items from the user
	if firstRun {
		return true
	}

	rule, res := d.CharacterCfg.Runtime.Rules.EvaluateAll(i)
	if res == nip.RuleResultFullMatch && b.doesExceedQuantity(i, rule, d) {
		return false
	}

	return true
}

func (b *Builder) stashItemAction(i data.Item, firstRun bool) bool {
	screenPos := b.UIManager.GetScreenCoordsForItem(i)
	b.HID.MovePointer(screenPos.X, screenPos.Y)
	helper.Sleep(170)
	screenshot := b.Reader.Screenshot()
	helper.Sleep(150)
	b.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
	helper.Sleep(500)

	d := b.Reader.GetData(false)
	for _, it := range d.Inventory.ByLocation(item.LocationInventory) {
		if it.UnitID == i.UnitID {
			return false
		}
	}

	// Don't log items that we already have in inventory during first run
	if !firstRun {
		event.Send(event.ItemStashed(event.WithScreenshot(b.Supervisor, fmt.Sprintf("Item %s [%d] stashed", i.Name, i.Quality), screenshot), i))
	}
	return true
}

func (b *Builder) clickStashGoldBtn() {
	helper.Sleep(170)
	if b.Reader.LegacyGraphics() {
		b.HID.Click(game.LeftButton, ui.StashGoldBtnXClassic, ui.StashGoldBtnYClassic)
		helper.Sleep(1000)
		b.HID.Click(game.LeftButton, ui.StashGoldBtnConfirmXClassic, ui.StashGoldBtnConfirmYClassic)
	} else {
		b.HID.Click(game.LeftButton, ui.StashGoldBtnX, ui.StashGoldBtnY)
		helper.Sleep(1000)
		b.HID.Click(game.LeftButton, ui.StashGoldBtnConfirmX, ui.StashGoldBtnConfirmY)
	}

}

func (b *Builder) switchTab(tab int) {
	if b.Reader.LegacyGraphics() {
		x := ui.SwitchStashTabBtnXClassic
		y := ui.SwitchStashTabBtnYClassic

		tabSize := ui.SwitchStashTabBtnTabSizeClassic
		x = x + tabSize*tab - tabSize/2
		b.HID.Click(game.LeftButton, x, y)
		helper.Sleep(500)
	} else {
		x := ui.SwitchStashTabBtnX
		y := ui.SwitchStashTabBtnY

		tabSize := ui.SwitchStashTabBtnTabSize
		x = x + tabSize*tab - tabSize/2
		b.HID.Click(game.LeftButton, x, y)
		helper.Sleep(500)
	}
}

func (b *Builder) SwitchStashTab(tab int) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		if d.LegacyGraphics {
			x := ui.SwitchStashTabBtnXClassic
			y := ui.SwitchStashTabBtnYClassic

			tabSize := ui.SwitchStashTabBtnTabSizeClassic
			x = x + tabSize*tab - tabSize/2
			b.HID.Click(game.LeftButton, x, y)
			helper.Sleep(500)
		} else {
			x := ui.SwitchStashTabBtnX
			y := ui.SwitchStashTabBtnY

			tabSize := ui.SwitchStashTabBtnTabSize
			x = x + tabSize*tab - tabSize/2
			b.HID.Click(game.LeftButton, x, y)
			helper.Sleep(500)
		}

		return []Action{}
	})
}
