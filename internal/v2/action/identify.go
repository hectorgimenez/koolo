package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/nip"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/action/step"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/ui"
	"github.com/hectorgimenez/koolo/internal/v2/utils"
)

func IdentifyAll(skipIdentify bool) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "IdentifyAll"

	items := itemsToIdentify()

	ctx.Logger.Debug("Checking for items to identify...")
	if len(items) == 0 || skipIdentify {
		ctx.Logger.Debug("No items to identify...")
		return nil
	}

	idTome, found := ctx.Data.Inventory.Find(item.TomeOfIdentify, item.LocationInventory)
	if !found {
		ctx.Logger.Warn("ID Tome not found, not identifying items")
		return nil
	}

	if st, statFound := idTome.FindStat(stat.Quantity, 0); !statFound || st.Value < len(items) {
		ctx.Logger.Info("Not enough ID scrolls, refilling...")
		VendorRefill(true, false)
	}

	ctx.Logger.Info(fmt.Sprintf("Identifying %d items...", len(items)))
	step.CloseAllMenus()
	for !ctx.Data.OpenMenus.Inventory {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
		utils.Sleep(300)
	}
	for _, i := range items {
		identifyItem(idTome, i)
	}
	step.CloseAllMenus()

	return nil
}

func itemsToIdentify() (items []data.Item) {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "itemsToIdentify"

	for _, i := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if i.Identified || i.Quality == item.QualityNormal || i.Quality == item.QualitySuperior {
			continue
		}

		// Skip identifying items that fully match a rule when unid
		if _, result := ctx.CharacterCfg.Runtime.Rules.EvaluateAll(i); result == nip.RuleResultFullMatch {
			continue
		}

		items = append(items, i)
	}

	return
}

func identifyItem(idTome data.Item, i data.Item) {
	ctx := context.Get()
	screenPos := ui.GetScreenCoordsForItem(idTome)

	utils.Sleep(500)
	ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
	utils.Sleep(1000)

	screenPos = ui.GetScreenCoordsForItem(i)

	ctx.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
	utils.Sleep(350)
}
