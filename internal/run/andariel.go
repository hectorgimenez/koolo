package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

var andarielStartingPosition = data.Position{
	X: 22561,
	Y: 9553,
}

var andarielClearPos1 = data.Position{
	X: 22570,
	Y: 9591,
}

var andarielClearPos2 = data.Position{
	X: 22547,
	Y: 9593,
}

var andarielClearPos3 = data.Position{
	X: 22533,
	Y: 9591,
}

var andarielClearPos4 = data.Position{
	X: 22535,
	Y: 9579,
}

var andarielClearPos5 = data.Position{
	X: 22548,
	Y: 9580,
}

var andarielAttackPos1 = data.Position{
	X: 22548,
	Y: 9570,
}

// Placeholder for second attack position
//var andarielAttackPos2 = data.Position{
//	X: 22548,
//	Y: 9590,
//}

type Andariel struct {
	ctx *context.Status
}

func NewAndariel() *Andariel {
	return &Andariel{
		ctx: context.Get(),
	}
}

func (a Andariel) Name() string {
	return string(config.AndarielRun)
}

func (a Andariel) Run() error {
	// Moving to Catacombs Level 4
	a.ctx.Logger.Info("Moving to Catacombs 4")
	err := action.WayPoint(area.CatacombsLevel2)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.CatacombsLevel3)
	action.MoveToArea(area.CatacombsLevel4)
	if err != nil {
		return err
	}

	if a.ctx.CharacterCfg.Game.Andariel.UseAntidoes {
		reHidePortraits := false
		action.ReturnTown()

		potsToBuy := 4
		if a.ctx.Data.MercHPPercent() > 0 {
			potsToBuy = 8
			if a.ctx.CharacterCfg.HidePortraits && !a.ctx.Data.OpenMenus.PortraitsShown {
				a.ctx.CharacterCfg.HidePortraits = false
				reHidePortraits = true
				a.ctx.HID.PressKey(a.ctx.Data.KeyBindings.ShowPortraits.Key1[0])
			}
		}

		action.VendorRefill(false, true)
		action.BuyAtVendor(npc.Akara, action.VendorItemRequest{
			Item:     "AntidotePotion",
			Quantity: potsToBuy,
			Tab:      4,
		})

		a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)

		x := 0
		for _, itm := range a.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
			if itm.Name != "AntidotePotion" {
				continue
			}
			pos := ui.GetScreenCoordsForItem(itm)
			utils.Sleep(500)

			if x > 3 {

				a.ctx.HID.Click(game.LeftButton, pos.X, pos.Y)
				utils.Sleep(300)
				if a.ctx.Data.LegacyGraphics {
					a.ctx.HID.Click(game.LeftButton, ui.MercAvatarPositionXClassic, ui.MercAvatarPositionYClassic)
				} else {
					a.ctx.HID.Click(game.LeftButton, ui.MercAvatarPositionX, ui.MercAvatarPositionY)
				}

			} else {
				a.ctx.HID.Click(game.RightButton, pos.X, pos.Y)
			}
			x++
		}
		step.CloseAllMenus()

		if reHidePortraits {
			a.ctx.CharacterCfg.HidePortraits = true
		}
		action.HidePortraits()

		action.UsePortalInTown()
	}

	if a.ctx.CharacterCfg.Game.Andariel.ClearRoom {
		// Clearing inside room
		a.ctx.Logger.Info("Clearing inside room")
		action.MoveToCoords(andarielClearPos1)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos2)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos3)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos4)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielClearPos5)
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
		action.MoveToCoords(andarielAttackPos1)
		action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())

	} else {
		action.MoveToCoords(andarielStartingPosition)
	}

	// Disable item pickup while fighting Andariel (prevent picking up items if nearby monsters die)
	a.ctx.DisableItemPickup()

	// Attacking Andariel
	a.ctx.Logger.Info("Killing Andariel")
	err = a.ctx.Char.KillAndariel()

	// Enable item pickup after the fight
	a.ctx.EnableItemPickup()

	return err
}
