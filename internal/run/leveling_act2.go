package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	ui2 "github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func (a Leveling) act2() error {
	running := false

	if running || a.ctx.Data.PlayerUnit.Area != area.LutGholein {
		return nil
	}

	running = true
	// Find Horadric Cube
	_, found := a.ctx.Data.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("Horadric Cube found, skipping quest")
	} else {
		a.ctx.Logger.Info("Horadric Cube not found, starting quest")
		return a.findHoradricCube()
	}

	if a.ctx.Data.Quests[quest.Act2TheSummoner].Completed() {
		// Try to get level 21 before moving to Duriel and Act3

		if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 18 {
			return TalRashaTombs{}.Run()
		}

		return a.duriel(a.ctx.Data.Quests[quest.Act2TheHoradricStaff].Completed(), *a.ctx.Data)
	}

	_, horadricStaffFound := a.ctx.Data.Inventory.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)

	// Find Staff of Kings
	_, found = a.ctx.Data.Inventory.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if found || horadricStaffFound {
		a.ctx.Logger.Info("StaffOfKings found, skipping quest")
	} else {
		a.ctx.Logger.Info("StaffOfKings not found, starting quest")
		return a.findStaff()
	}

	// Find Amulet
	_, found = a.ctx.Data.Inventory.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if found || horadricStaffFound {
		a.ctx.Logger.Info("Amulet of the Viper found, skipping quest")
	} else {
		a.ctx.Logger.Info("Amulet of the Viper not found, starting quest")
		return a.findAmulet()
	}

	// Summoner
	a.ctx.Logger.Info("Starting summoner quest")
	return a.summoner()

}

func (a Leveling) findHoradricCube() error {
	err := action2.WayPoint(area.HallsOfTheDeadLevel2)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.HallsOfTheDeadLevel3)
	if err != nil {
		return err
	}

	err = action2.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.HoradricCubeChest)
		if found {
			a.ctx.Logger.Info("Horadric Cube chest found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action2.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.HoradricCubeChest)
	if !found {
		return err
	}

	err = action2.InteractObject(obj, func() bool {
		updatedObj, found := a.ctx.Data.Objects.FindOne(object.HoradricCubeChest)
		if found {
			if !updatedObj.Selectable {
				a.ctx.Logger.Debug("Interacted with Horadric Cube Chest")
			}
			return !updatedObj.Selectable
		}
		return false
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) findStaff() error {
	err := action2.WayPoint(area.FarOasis)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.MaggotLairLevel1)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.MaggotLairLevel2)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.MaggotLairLevel3)
	if err != nil {
		return err
	}
	err = action2.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
		if found {
			a.ctx.Logger.Info("Staff Of Kings chest found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action2.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
	if !found {
		return err
	}

	err = action2.InteractObject(obj, func() bool {
		updatedObj, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
		if found {
			if !updatedObj.Selectable {
				a.ctx.Logger.Debug("Interacted with Staff Of Kings Chest")
			}
			return !updatedObj.Selectable
		}
		return false
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) findAmulet() error {
	err := action2.WayPoint(area.LostCity)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.ValleyOfSnakes)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.ClawViperTempleLevel1)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.ClawViperTempleLevel2)
	if err != nil {
		return err
	}
	err = action2.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.TaintedSunAltar)
		if found {
			a.ctx.Logger.Info("Tainted Sun Altar found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action2.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.TaintedSunAltar)
	if !found {
		return err
	}

	err = action2.InteractObject(obj, func() bool {
		updatedObj, found := a.ctx.Data.Objects.FindOne(object.TaintedSunAltar)
		if found {
			if !updatedObj.Selectable {
				a.ctx.Logger.Debug("Interacted with Tainted Sun Altar")
			}
			return !updatedObj.Selectable
		}
		return false
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) summoner() error {
	// Start Summoner run to find and kill Summoner
	err := Summoner{}.Run()
	if err != nil {
		return err
	}

	tome, found := a.ctx.Data.Objects.FindOne(object.YetAnotherTome)
	if !found {
		return err
	}

	// Try to use the portal and discover the waypoint
	err = action2.InteractObject(tome, func() bool {
		_, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
		return found
	})
	if err != nil {
		return err
	}

	portal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return err
	}

	err = action2.InteractObject(portal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.CanyonOfTheMagi
	})
	if err != nil {
		return err
	}

	err = action2.DiscoverWaypoint()
	if err != nil {
		return err
	}

	err = action2.ReturnTown()
	if err != nil {
		return err
	}

	err = action2.InteractNPC(npc.Atma)
	if err != nil {
		return err
	}

	a.ctx.HID.PressKey(win.VK_ESCAPE)

	return nil
}

func (a Leveling) prepareStaff() error {
	horadricStaff, found := a.ctx.Data.Inventory.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if found {
		a.ctx.Logger.Info("Horadric Staff found!")
		if horadricStaff.Location.LocationType == item.LocationStash {
			a.ctx.Logger.Info("It's in the stash, let's pick it up")

			bank, found := a.ctx.Data.Objects.FindOne(object.Bank)
			if !found {
				a.ctx.Logger.Info("bank object not found")
			}

			err := action2.InteractObject(bank, func() bool {
				return a.ctx.Data.OpenMenus.Stash
			})
			if err != nil {
				return err
			}

			screenPos := ui2.GetScreenCoordsForItem(horadricStaff)
			a.ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
			utils.Sleep(300)
			a.ctx.HID.PressKey(win.VK_ESCAPE)

			return nil
		}
	}

	staff, found := a.ctx.Data.Inventory.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if !found {
		a.ctx.Logger.Info("Staff of Kings not found, skipping")
		return nil
	}

	amulet, found := a.ctx.Data.Inventory.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if !found {
		a.ctx.Logger.Info("Amulet of the Viper not found, skipping")
		return nil
	}

	err := action2.CubeAddItems(staff, amulet)
	if err != nil {
		return err
	}

	err = action2.CubeTransmute()
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) duriel(staffAlreadyUsed bool, d game.Data) error {
	a.ctx.Logger.Info("Starting Duriel....")

	var realTomb area.ID
	for _, tomb := range talRashaTombs {
		for _, obj := range a.ctx.Data.Areas[tomb].Objects {
			if obj.Name == object.HoradricOrifice {
				realTomb = tomb
				break
			}
		}
	}

	if realTomb == 0 {
		a.ctx.Logger.Info("Could not find the real tomb :(")
		return nil
	}

	if !staffAlreadyUsed {
		a.prepareStaff()
	}

	// Move close to the Horadric Orifice
	action2.WayPoint(area.CanyonOfTheMagi)
	action2.Buff()
	action2.MoveToArea(realTomb)
	action2.MoveTo(func() (data.Position, bool) {
		orifice, found := d.Objects.FindOne(object.HoradricOrifice)
		if !found {
			return data.Position{}, false
		}
		return orifice.Position, true
	})

	// If staff has not been used, then put it in the orifice and wait for the entrance to open
	if !staffAlreadyUsed {
		action2.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())

		orifice, found := a.ctx.Data.Objects.FindOne(object.HoradricOrifice)
		if !found {
			a.ctx.Logger.Info("Horadric Orifice not found")
			return nil
		}

		action2.InteractObject(orifice, func() bool {
			return d.OpenMenus.Anvil
		})

		staff, _ := d.Inventory.Find("HoradricStaff", item.LocationInventory)
		screenPos := ui2.GetScreenCoordsForItem(staff)

		a.ctx.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
		utils.Sleep(300)
		a.ctx.HID.Click(game.LeftButton, ui2.AnvilCenterX, ui2.AnvilCenterY)
		utils.Sleep(500)
		a.ctx.HID.Click(game.LeftButton, ui2.AnvilBtnX, ui2.AnvilBtnY)
		utils.Sleep(20000)
	}

	potsToBuy := 4
	if d.MercHPPercent() > 0 {
		potsToBuy = 8
	}

	// Return to the city, ensure we have pots and everything, and get some thawing potions
	action2.ReturnTown()
	action2.ReviveMerc()
	action2.VendorRefill(false, true)
	action2.BuyAtVendor(npc.Lysander, action2.VendorItemRequest{
		Item:     "ThawingPotion",
		Quantity: potsToBuy,
		Tab:      4,
	})

	a.ctx.HID.PressKeyBinding(d.KeyBindings.Inventory)
	x := 0
	for _, itm := range d.Inventory.ByLocation(item.LocationInventory) {
		if itm.Name != "ThawingPotion" {
			continue
		}

		pos := ui2.GetScreenCoordsForItem(itm)
		utils.Sleep(500)

		if x > 3 {
			a.ctx.HID.Click(game.LeftButton, pos.X, pos.Y)
			utils.Sleep(300)
			if d.LegacyGraphics {
				a.ctx.HID.Click(game.LeftButton, ui2.MercAvatarPositionXClassic, ui2.MercAvatarPositionYClassic)
			} else {
				a.ctx.HID.Click(game.LeftButton, ui2.MercAvatarPositionX, ui2.MercAvatarPositionY)
			}
		} else {
			a.ctx.HID.Click(game.RightButton, pos.X, pos.Y)
		}
		x++
	}

	a.ctx.HID.PressKey(win.VK_ESCAPE)

	action2.UsePortalInTown()
	action2.Buff()

	duriellair, found := d.Objects.FindOne(object.DurielsLairPortal)
	if found {
		action2.InteractObject(duriellair, func() bool {
			return d.PlayerUnit.Area == area.DurielsLair
		})
	}

	a.ctx.Char.KillDuriel()

	action2.MoveToCoords(data.Position{
		X: 22577,
		Y: 15613,
	})

	action2.InteractNPC(npc.Tyrael)
	a.ctx.HID.PressKey(win.VK_ESCAPE)

	action2.ReturnTown()
	action2.MoveToCoords(data.Position{
		X: 5092,
		Y: 5144,
	})

	action2.InteractNPC(npc.Jerhyn)
	a.ctx.HID.PressKey(win.VK_ESCAPE)

	action2.MoveToCoords(data.Position{
		X: 5195,
		Y: 5060,
	})
	action2.InteractNPC(npc.Meshif)
	a.ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)

	return nil
}
