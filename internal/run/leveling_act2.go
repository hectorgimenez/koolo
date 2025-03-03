package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func (a Leveling) act2() error {
	running := false

	if running || a.ctx.Data.PlayerUnit.Area != area.LutGholein {
		return nil
	}

	running = true

	if a.ctx.Data.Quests[quest.Act2TheSevenTombs].HasStatus(quest.StatusInProgress5) {
		action.MoveToCoords(data.Position{
			X: 5092,
			Y: 5144,
		})

		action.InteractNPC(npc.Jerhyn)
	}

	if a.ctx.Data.Quests[quest.Act2TheSevenTombs].Completed() {
		action.MoveToCoords(data.Position{
			X: 5195,
			Y: 5060,
		})
		action.InteractNPC(npc.Meshif)
		a.ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
		return nil
	}
	// Find Horadric Cube
	_, found := a.ctx.Data.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("Horadric Cube found, skipping quest")
	} else {
		a.ctx.Logger.Info("Horadric Cube not found, starting quest")
		return NewQuests().getHoradricCube()
	}

	// Duriel quest only starts when we click the journal. If we haven't clicked the journal we probably don't even have the Canyon WP.
	if !a.ctx.Data.Quests[quest.Act2TheSevenTombs].HasStatus(quest.StatusQuestNotStarted) {
		// Try to get level 21 before moving to Duriel and Act3

		if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 21 {
			return NewTalRashaTombs().Run()
		}

		a.prepareStaff()

		return a.duriel()
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
		a.findAmulet()
	}

	if !a.ctx.Data.Quests[quest.Act2TheSummoner].Completed() {
		// Summoner
		a.ctx.Logger.Info("Starting summoner quest")
		err := NewSummoner().Run()
		if err != nil {
			return err
		}

		// This block can be removed when https://github.com/hectorgimenez/koolo/pull/642 gets merged
		tome, found := a.ctx.Data.Objects.FindOne(object.YetAnotherTome)
		if !found {
			return err
		}

		// Try to use the portal and discover the waypoint
		err = action.InteractObject(tome, func() bool {
			_, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
			return found
		})
		if err != nil {
			return err
		}
		a.ctx.Logger.Debug("Moving to red portal")
		portal, _ := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)

		action.InteractObject(portal, func() bool {
			return a.ctx.Data.PlayerUnit.Area == area.CanyonOfTheMagi && a.ctx.Data.AreaData.IsInside(a.ctx.Data.PlayerUnit.Position)
		})
		// End of block for removal
		err = action.DiscoverWaypoint()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a Leveling) findStaff() error {
	err := action.WayPoint(area.FarOasis)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.MaggotLairLevel1)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.MaggotLairLevel2)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.MaggotLairLevel3)
	if err != nil {
		return err
	}

	obj, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
	if !found {
		return err
	}

	err = action.InteractObject(obj, func() bool {
		updatedObj, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
		if found {
			return !updatedObj.Selectable
		}
		return false
	})
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) findAmulet() error {
	err := action.WayPoint(area.LostCity)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.ValleyOfSnakes)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.ClawViperTempleLevel1)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.ClawViperTempleLevel2)
	if err != nil {
		return err
	}
	err = action.MoveTo(func() (data.Position, bool) {
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

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.TaintedSunAltar)
	if !found {
		return err
	}

	err = action.InteractObject(obj, func() bool {
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

	action.ReturnTown()

	// This stops us being blocked from getting into Palace
	action.InteractNPC(npc.Drognan)

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

			err := action.InteractObject(bank, func() bool {
				return a.ctx.Data.OpenMenus.Stash
			})
			if err != nil {
				return err
			}

			screenPos := ui.GetScreenCoordsForItem(horadricStaff)
			a.ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
			utils.Sleep(300)
			step.CloseAllMenus()

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

	err := action.CubeAddItems(staff, amulet)
	if err != nil {
		return err
	}

	err = action.CubeTransmute()
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) duriel() error {
	a.ctx.Logger.Info("Starting Duriel....")
	a.ctx.CharacterCfg.Game.Duriel.UseThawing = true
	if err := NewDuriel().Run(); err != nil {
		return err
	}
	duriel, found := a.ctx.Data.Monsters.FindOne(npc.Duriel, data.MonsterTypeUnique)
	if !found || duriel.Stats[stat.Life] <= 0 || a.ctx.Data.Quests[quest.Act2TheSevenTombs].HasStatus(quest.StatusInProgress3) {

		action.MoveToCoords(data.Position{
			X: 22577,
			Y: 15600,
		})
		action.InteractNPC(npc.Tyrael)

	}

	action.ReturnTown()

	return nil
}
