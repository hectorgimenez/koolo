package action

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

const (
	maxGoldPerStashTab = 2500000
)

func Stash(forceStash bool) error {
	ctx := context.Get()
	ctx.SetLastAction("Stash")

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
	// Clear messages like TZ change or public game spam.  Prevent bot from clicking on messages
	ClearMessages()
	stashGold()
	orderInventoryPotions()
	stashInventory(forceStash)
	step.CloseAllMenus()

	return nil
}

func orderInventoryPotions() {
	ctx := context.Get()
	ctx.SetLastStep("orderInventoryPotions")

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
	ctx.SetLastStep("isStashingRequired")

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
	ctx.SetLastAction("stashGold")

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
			SwitchStashTab(tab + 1)
			clickStashGoldBtn()
			utils.Sleep(500)
		}
	}

	ctx.Logger.Info("All stash tabs are full of gold :D")
}

func stashInventory(firstRun bool) {
	ctx := context.Get()
	ctx.SetLastAction("stashInventory")

	currentTab := 1
	if ctx.CharacterCfg.Character.StashToShared {
		currentTab = 2
	}
	SwitchStashTab(currentTab)

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		stashIt, matchedRule, ruleFile := shouldStashIt(i, firstRun)

		if !stashIt {
			continue
		}

		// Always stash unique charms to the shared stash
		if (i.Name == "grandcharm" || i.Name == "smallcharm" || i.Name == "largecharm") && i.Quality == item.QualityUnique {
			currentTab = 2
			SwitchStashTab(currentTab)
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
				ctx.Logger.Info("Stash is full ...")
				//TODO: Stash is full stop the bot
			}
			ctx.Logger.Debug(fmt.Sprintf("Tab %d is full, switching to next one", currentTab))
			currentTab++
			SwitchStashTab(currentTab)
		}
	}
}

func shouldStashIt(i data.Item, firstRun bool) (bool, string, string) {
	ctx := context.Get()
	ctx.SetLastStep("shouldStashIt")

	// Don't stash items in protected slots
	if ctx.CharacterCfg.Inventory.InventoryLock[i.Position.Y][i.Position.X] == 0 {
		return false, "", ""
	}

	// Don't stash items from quests during leveling process, it makes things easier to track
	if _, isLevelingChar := ctx.Char.(context.LevelingCharacter); isLevelingChar && i.IsFromQuest() && i.Name != "HoradricCube" {
		return false, "", ""
	}

	if i.IsRuneword {
		return true, "Runeword", ""
	}

	// Stash items that are part of a recipe which are not covered by the NIP rules
	if shouldKeepRecipeItem(i) {
		return true, "Item is part of a enabled recipe", ""
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
	if res == nip.RuleResultFullMatch && doesExceedQuantity(rule) {
		return false, "", ""
	}

	// Full rule match
	if res == nip.RuleResultFullMatch {
		return true, rule.RawLine, rule.Filename + ":" + strconv.Itoa(rule.LineNumber)
	}
	return false, "", ""
}

func shouldKeepRecipeItem(i data.Item) bool {
	ctx := context.Get()
	ctx.SetLastStep("shouldKeepRecipeItem")

	// No items with quality higher than magic can be part of a recipe
	if i.Quality > item.QualityMagic {
		return false
	}

	itemInStashNotMatchingRule := false

	// Check if we already have the item in our stash and if it doesn't match any of our pickit rules
	for _, it := range ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash) {
		if it.Name == i.Name {
			_, res := ctx.CharacterCfg.Runtime.Rules.EvaluateAll(it)
			if res != nip.RuleResultFullMatch {
				itemInStashNotMatchingRule = true
			}
		}
	}

	recipeMatch := false

	// Check if the item is part of a recipe and if that recipe is enabled
	for _, recipe := range Recipes {
		if slices.Contains(recipe.Items, string(i.Name)) && slices.Contains(ctx.CharacterCfg.CubeRecipes.EnabledRecipes, recipe.Name) {
			recipeMatch = true
			break
		}
	}

	if recipeMatch && !itemInStashNotMatchingRule {
		return true
	}

	return false
}

func stashItemAction(i data.Item, rule string, ruleFile string, skipLogging bool) bool {
	ctx := context.Get()
	ctx.SetLastAction("stashItemAction")

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

	dropLocation := "unknown"

	// log the contents of picked up items
	ctx.Logger.Info(fmt.Sprintf("Picked up items: %v", ctx.CurrentGame.PickedUpItems))

	if _, found := ctx.CurrentGame.PickedUpItems[int(i.UnitID)]; found {
		areaId := ctx.CurrentGame.PickedUpItems[int(i.UnitID)]
		dropLocation = area.ID(ctx.CurrentGame.PickedUpItems[int(i.UnitID)]).Area().Name

		if slices.Contains(ctx.Data.TerrorZones, area.ID(areaId)) {
			dropLocation += " (terrorized)"
		}
	}

	// Don't log items that we already have in inventory during first run or that we don't want to notify about (gems, low runes .. etc)
	if !skipLogging && shouldNotifyAboutStashing(i) && ruleFile != "" {
		event.Send(event.ItemStashed(event.WithScreenshot(ctx.Name, fmt.Sprintf("Item %s [%d] stashed", i.Name, i.Quality), screenshot), data.Drop{Item: i, Rule: rule, RuleFile: ruleFile, DropLocation: dropLocation}))
	}

	return true
}

func shouldNotifyAboutStashing(i data.Item) bool {
	ctx := context.Get()

	ctx.Logger.Debug(fmt.Sprintf("Checking if we should notify about stashing %s %v", i.Name, i.Desc()))
	// Don't notify about gems
	if strings.Contains(i.Desc().Type, "gem") {
		return false
	}

	// Skip low runes (below lem)
	lowRunes := []string{"elrune", "eldrune", "tirrune", "nefrune", "ethrune", "ithrune", "talrune", "ralrune", "ortrune", "thulrune", "amnrune", "solrune", "shaelrune", "dolrune", "helrune", "iorune", "lumrune", "korune", "falrune"}
	if i.Desc().Type == item.TypeRune {
		itemName := strings.ToLower(string(i.Name))
		for _, runeName := range lowRunes {
			if itemName == runeName {
				return false
			}
		}
	}

	return true
}

func clickStashGoldBtn() {
	ctx := context.Get()
	ctx.SetLastStep("clickStashGoldBtn")

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

func SwitchStashTab(tab int) {
	ctx := context.Get()
	ctx.SetLastStep("switchTab")

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

func OpenStash() error {
	ctx := context.Get()
	ctx.SetLastAction("OpenStash")

	bank, found := ctx.Data.Objects.FindOne(object.Bank)
	if !found {
		return errors.New("stash not found")
	}
	InteractObject(bank,
		func() bool {
			return ctx.Data.OpenMenus.Stash
		},
	)

	return nil
}

func CloseStash() error {
	ctx := context.Get()
	ctx.SetLastAction("CloseStash")

	if ctx.Data.OpenMenus.Stash {
		ctx.HID.PressKey(win.VK_ESCAPE)
	} else {
		return errors.New("stash is not open")
	}

	return nil
}

func TakeItemsFromStash(stashedItems []data.Item) error {
	ctx := context.Get()
	ctx.SetLastAction("TakeItemsFromStash")

	if ctx.Data.OpenMenus.Stash {
		err := OpenStash()
		if err != nil {
			return err
		}
	}

	utils.Sleep(250)

	for _, i := range stashedItems {

		if i.Location.LocationType != item.LocationStash && i.Location.LocationType != item.LocationSharedStash {
			continue
		}

		// Make sure we're on the correct tab
		SwitchStashTab(i.Location.Page + 1)

		// Move the item to the inventory
		screenPos := ui.GetScreenCoordsForItem(i)
		ctx.HID.MovePointer(screenPos.X, screenPos.Y)
		ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
		utils.Sleep(500)
	}

	return nil
}
