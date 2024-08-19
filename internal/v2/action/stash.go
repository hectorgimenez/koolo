package action

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/action/step"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/ui"
	"github.com/hectorgimenez/koolo/internal/v2/utils"
)

const (
	maxGoldPerStashTab = 2500000
)

func Stash(forceStash bool) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "Stash"

	ctx.Logger.Debug("Checking for items to stash...")
	if !isStashingRequired(forceStash) {
		return nil
	}

	ctx.Logger.Info("Stashing items...")

	switch ctx.Data.PlayerUnit.Area {
	case area.KurastDocks:
		MoveToCoords(data.Position{X: 5146, Y: 5067})
	case area.LutGholein:
		MoveToCoords(data.Position{X: 5130, Y: 5086})
	}

	bank, _ := ctx.Data.Objects.FindOne(object.Bank)
	InteractObject(bank,
		func() bool {
			return ctx.Data.OpenMenus.Stash
		},
	)

	stashGold()
	orderInventoryPotions()
	stashInventory(forceStash)
	step.CloseAllMenus()

	return nil
}

func orderInventoryPotions() {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "orderInventoryPotions"

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if i.IsPotion() {
			if ctx.CharacterCfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 {
				continue
			}

			screenPos := ui.GetScreenCoordsForItem(i)
			utils.Sleep(100)
			ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
			utils.Sleep(200)
		}
	}
}

func isStashingRequired(firstRun bool) bool {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "isStashingRequired"

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		stashIt, _, _ := shouldStashIt(i, firstRun)
		if stashIt {
			return true
		}
	}

	isStashFull := true
	for _, goldInStash := range ctx.Data.Inventory.StashedGold {
		if goldInStash < maxGoldPerStashTab {
			isStashFull = false
		}
	}

	if ctx.Data.Inventory.Gold > ctx.Data.PlayerUnit.MaxGold()/3 && !isStashFull {
		return true
	}

	return false
}

func stashGold() {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "stashGold"

	if ctx.Data.Inventory.Gold == 0 {
		return
	}

	ctx.Logger.Info("Stashing gold...", slog.Int("gold", ctx.Data.Inventory.Gold))

	for tab, goldInStash := range ctx.Data.Inventory.StashedGold {
		ctx.RefreshGameData()
		if ctx.Data.Inventory.Gold == 0 {
			return
		}

		if goldInStash < maxGoldPerStashTab {
			switchTab(tab + 1)
			clickStashGoldBtn()
			utils.Sleep(500)
		}
	}

	ctx.Logger.Info("All stash tabs are full of gold :D")
}

func stashInventory(firstRun bool) {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "stashInventory"

	currentTab := 1
	if ctx.CharacterCfg.Character.StashToShared {
		currentTab = 2
	}
	switchTab(currentTab)

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		stashIt, matchedRule, ruleFile := shouldStashIt(i, firstRun)

		if !stashIt {
			continue
		}
		for currentTab < 5 {
			if stashItemAction(i, matchedRule, ruleFile, firstRun) {
				r, res := ctx.CharacterCfg.Runtime.Rules.EvaluateAll(i)

				if res != nip.RuleResultFullMatch && firstRun {
					ctx.Logger.Info(
						fmt.Sprintf("Item %s [%s] stashed because it was found in the inventory during the first run.", i.Desc().Name, i.Quality.ToString()),
					)
					break
				}

				ctx.Logger.Info(
					fmt.Sprintf("Item %s [%s] stashed", i.Desc().Name, i.Quality.ToString()),
					slog.String("nipFile", fmt.Sprintf("%s:%d", r.Filename, r.LineNumber)),
					slog.String("rawRule", r.RawLine),
				)
				break
			}
			if currentTab == 5 {
				// TODO: Stop the bot, stash is full
			}
			ctx.Logger.Debug(fmt.Sprintf("Tab %d is full, switching to next one", currentTab))
			currentTab++
			switchTab(currentTab)
		}
	}
}

func shouldStashIt(i data.Item, firstRun bool) (bool, string, string) {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "shouldStashIt"

	// Don't stash items from quests during leveling process, it makes things easier to track
	if _, isLevelingChar := ctx.Char.(context.LevelingCharacter); isLevelingChar && i.IsFromQuest() {
		return false, "", ""
	}

	if i.IsRuneword {
		return true, "runeword", ""
	}

	// Don't stash the Tomes, keys and WirtsLeg
	if i.Name == item.TomeOfTownPortal || i.Name == item.TomeOfIdentify || i.Name == item.Key || i.Name == "WirtsLeg" {
		return false, "", ""
	}

	if i.Position.Y >= len(ctx.CharacterCfg.Inventory.InventoryLock) || i.Position.X >= len(ctx.CharacterCfg.Inventory.InventoryLock[0]) {
		return false, "", ""
	}

	if i.Location.LocationType == item.LocationInventory && ctx.CharacterCfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 || i.IsPotion() {
		return false, "", ""
	}

	// Let's stash everything during first run, we don't want to sell items from the user
	if firstRun {
		return true, "FirstRun", ""
	}

	rule, res := ctx.CharacterCfg.Runtime.Rules.EvaluateAll(i)
	if res == nip.RuleResultFullMatch && doesExceedQuantity(i, rule) {
		return false, "", ""
	}

	return true, rule.RawLine, rule.Filename + ":" + strconv.Itoa(rule.LineNumber)
}

func stashItemAction(i data.Item, rule string, ruleFile string, firstRun bool) bool {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "stashItemAction"

	screenPos := ui.GetScreenCoordsForItem(i)
	ctx.HID.MovePointer(screenPos.X, screenPos.Y)
	utils.Sleep(170)
	screenshot := ctx.GameReader.Screenshot()
	utils.Sleep(150)
	ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
	utils.Sleep(500)

	for _, it := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if it.UnitID == i.UnitID {
			return false
		}
	}

	// Don't log items that we already have in inventory during first run
	if !firstRun {
		event.Send(event.ItemStashed(event.WithScreenshot(ctx.Name, fmt.Sprintf("Item %s [%d] stashed", i.Name, i.Quality), screenshot), data.Drop{Item: i, Rule: rule, RuleFile: ruleFile}))
	}

	return true
}

func clickStashGoldBtn() {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "clickStashGoldBtn"

	utils.Sleep(170)
	if ctx.GameReader.LegacyGraphics() {
		ctx.HID.Click(game.LeftButton, ui.StashGoldBtnXClassic, ui.StashGoldBtnYClassic)
		utils.Sleep(1000)
		ctx.HID.Click(game.LeftButton, ui.StashGoldBtnConfirmXClassic, ui.StashGoldBtnConfirmYClassic)
	} else {
		ctx.HID.Click(game.LeftButton, ui.StashGoldBtnX, ui.StashGoldBtnY)
		utils.Sleep(1000)
		ctx.HID.Click(game.LeftButton, ui.StashGoldBtnConfirmX, ui.StashGoldBtnConfirmY)
	}
}

func switchTab(tab int) {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "switchTab"

	if ctx.GameReader.LegacyGraphics() {
		x := ui.SwitchStashTabBtnXClassic
		y := ui.SwitchStashTabBtnYClassic

		tabSize := ui.SwitchStashTabBtnTabSizeClassic
		x = x + tabSize*tab - tabSize/2
		ctx.HID.Click(game.LeftButton, x, y)
		utils.Sleep(500)
	} else {
		x := ui.SwitchStashTabBtnX
		y := ui.SwitchStashTabBtnY

		tabSize := ui.SwitchStashTabBtnTabSize
		x = x + tabSize*tab - tabSize/2
		ctx.HID.Click(game.LeftButton, x, y)
		utils.Sleep(500)
	}
}
