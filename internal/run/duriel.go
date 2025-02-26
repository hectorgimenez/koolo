package run

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxOrificeAttempts = 10
	orificeCheckDelay  = 200
)

var talTombs = []area.ID{area.TalRashasTomb1, area.TalRashasTomb2, area.TalRashasTomb3, area.TalRashasTomb4, area.TalRashasTomb5, area.TalRashasTomb6, area.TalRashasTomb7}

type Duriel struct {
	ctx *context.Status
}

func NewDuriel() *Duriel {
	return &Duriel{
		ctx: context.Get(),
	}
}

func (d Duriel) Name() string {
	return string(config.DurielRun)
}

func (d Duriel) Run() error {
	err := action.WayPoint(area.CanyonOfTheMagi)
	if err != nil {
		return err
	}

	// Find and move to the real Tal Rasha tomb
	realTalRashaTomb, err := d.findRealTomb()
	if err != nil {
		return err
	}

	err = action.MoveToArea(realTalRashaTomb)
	if err != nil {
		return err
	}

	// Wait for area to fully load and get synchronized
	utils.Sleep(500)
	d.ctx.RefreshGameData()

	// Find orifice with retry logic
	var portal data.Object
	var found bool

	for attempts := 0; attempts < maxOrificeAttempts; attempts++ {
		portal, found = d.ctx.Data.Objects.FindOne(object.DurielsLairPortal)
		if found && portal.Mode == mode.ObjectModeOpened {
			break
		}
		utils.Sleep(orificeCheckDelay)
		d.ctx.RefreshGameData()
	}

	if !found {
		return errors.New("failed to find Duriel's portal after multiple attempts")
	}

	if d.ctx.CharacterCfg.Game.Duriel.UseThawing {
		reHidePortraits := false
		action.ReturnTown()

		potsToBuy := 4
		if d.ctx.Data.MercHPPercent() > 0 {
			potsToBuy = 8
			if d.ctx.CharacterCfg.HidePortraits && !d.ctx.Data.OpenMenus.PortraitsShown {
				d.ctx.CharacterCfg.HidePortraits = false
				reHidePortraits = true
				d.ctx.HID.PressKey(d.ctx.Data.KeyBindings.ShowPortraits.Key1[0])
			}
		}

		action.VendorRefill(false, true)
		action.BuyAtVendor(npc.Lysander, action.VendorItemRequest{
			Item:     "ThawingPotion",
			Quantity: potsToBuy,
			Tab:      4,
		})

		d.ctx.HID.PressKeyBinding(d.ctx.Data.KeyBindings.Inventory)

		x := 0
		for _, itm := range d.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
			if itm.Name != "ThawingPotion" {
				continue
			}
			pos := ui.GetScreenCoordsForItem(itm)
			utils.Sleep(500)

			if x > 3 {

				d.ctx.HID.Click(game.LeftButton, pos.X, pos.Y)
				utils.Sleep(300)
				if d.ctx.Data.LegacyGraphics {
					d.ctx.HID.Click(game.LeftButton, ui.MercAvatarPositionXClassic, ui.MercAvatarPositionYClassic)
				} else {
					d.ctx.HID.Click(game.LeftButton, ui.MercAvatarPositionX, ui.MercAvatarPositionY)
				}

			} else {
				d.ctx.HID.Click(game.RightButton, pos.X, pos.Y)
			}
			x++
		}
		step.CloseAllMenus()
		if reHidePortraits {
			d.ctx.CharacterCfg.HidePortraits = true
		}
		action.HidePortraits()
	}

	err = action.InteractObject(portal, func() bool {
		return d.ctx.Data.PlayerUnit.Area == area.DurielsLair
	})
	if err != nil {
		return err
	}

	// Final refresh before fight
	d.ctx.RefreshGameData()

	utils.Sleep(700)

	return d.ctx.Char.KillDuriel()
}

func (d Duriel) findRealTomb() (area.ID, error) {
	var realTomb area.ID

	for _, tomb := range talTombs {
		for _, obj := range d.ctx.Data.Areas[tomb].Objects {
			if obj.Name == object.HoradricOrifice {
				realTomb = tomb
				break
			}
		}
	}

	if realTomb == 0 {
		return 0, errors.New("failed to find the real Tal Rasha tomb")
	}

	return realTomb, nil
}
