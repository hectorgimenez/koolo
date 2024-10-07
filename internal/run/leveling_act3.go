package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
)

func (a Leveling) act3() error {
	running := false

	if running || a.ctx.Data.PlayerUnit.Area != area.KurastDocks {
		return nil
	}

	// Try to find Hratli at pier, if he's there, talk to him, so he will move to the normal position later
	hratli, found := a.ctx.Data.Monsters.FindOne(npc.Hratli, data.MonsterTypeNone)
	if found {
		action.InteractNPC(hratli.Name)
	}

	running = true
	_, willFound := a.ctx.Data.Inventory.Find("KhalimsWill", item.LocationInventory, item.LocationStash)
	if willFound {
		a.openMephistoStairs()
	}

	hellgate, found := a.ctx.Data.Objects.FindOne(object.HellGate)
	if !found {
		a.ctx.Logger.Info("Gate to Pandemonium Fortress not found")
	}

	if a.ctx.Data.Quests[quest.Act3KhalimsWill].Completed() {
		//Mephisto{}.Run()
		action.InteractObject(hellgate, func() bool {
			return a.ctx.Data.PlayerUnit.Area == area.ThePandemoniumFortress
		})
	}

	// Find KhalimsEye
	_, found = a.ctx.Data.Inventory.Find("KhalimsEye", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("KhalimsEye found, skipping quest")
	} else {
		a.ctx.Logger.Info("KhalimsEye not found, starting quest")
		a.findKhalimsEye()
	}

	// Find KhalimsBrain
	_, found = a.ctx.Data.Inventory.Find("KhalimsBrain", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("KhalimsBrain found, skipping quest")
	} else {
		a.ctx.Logger.Info("KhalimsBrain not found, starting quest")
		a.findKhalimsBrain()
	}

	// Find KhalimsHeart
	_, found = a.ctx.Data.Inventory.Find("KhalimsHeart", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("KhalimsHeart found, skipping quest")
	} else {
		a.ctx.Logger.Info("KhalimsHeart not found, starting quest")
		a.findKhalimsHeart()
	}

	// Trav
	a.openMephistoStairs()

	return a.ctx.Char.KillMephisto()
}

func (a Leveling) findKhalimsEye() error {
	err := action.WayPoint(area.SpiderForest)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.SpiderCavern)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		for _, o := range a.ctx.Data.Objects {
			if o.Name == object.InifussTree {
				return o.Position, true
			}
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	err = action.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.KhalimChest3)
		if found {
			a.ctx.Logger.Info("Khalm Chest found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	kalimchest3, found := a.ctx.Data.Objects.FindOne(object.KhalimChest3)
	if !found {
		a.ctx.Logger.Debug("Khalim Chest not found")
	}

	err = action.InteractObject(kalimchest3, func() bool {
		chest, _ := a.ctx.Data.Objects.FindOne(object.KhalimChest3)
		return !chest.Selectable
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) findKhalimsBrain() error {
	err := action.WayPoint(area.FlayerJungle)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.FlayerDungeonLevel1)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.FlayerDungeonLevel2)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.FlayerDungeonLevel3)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		a.ctx.Logger.Info("Khalm Chest found, moving to that room")
		chest, found := a.ctx.Data.Objects.FindOne(object.KhalimChest2)

		return chest.Position, found
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	kalimchest2, found := a.ctx.Data.Objects.FindOne(object.KhalimChest2)
	if !found {
		a.ctx.Logger.Debug("Khalim Chest not found")
	}

	err = action.InteractObject(kalimchest2, func() bool {
		chest, _ := a.ctx.Data.Objects.FindOne(object.KhalimChest2)
		return !chest.Selectable
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) findKhalimsHeart() error {
	err := action.WayPoint(area.KurastBazaar)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.SewersLevel1Act3)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		for _, l := range a.ctx.Data.AdjacentLevels {
			if l.Area == area.SewersLevel2Act3 {
				return l.Position, true
			}
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())

	stairs, found := a.ctx.Data.Objects.FindOne(object.Act3SewerStairsToLevel3)
	if !found {
		a.ctx.Logger.Debug("Khalim Chest not found")
	}

	err = action.InteractObject(stairs, func() bool {
		o, _ := a.ctx.Data.Objects.FindOne(object.Act3SewerStairsToLevel3)

		return !o.Selectable
	})
	if err != nil {
		return err
	}

	time.Sleep(3000)

	err = action.InteractObject(stairs, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.SewersLevel2Act3
	})
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		a.ctx.Logger.Info("Khalm Chest found, moving to that room")
		chest, found := a.ctx.Data.Objects.FindOne(object.KhalimChest1)

		return chest.Position, found
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	kalimchest1, found := a.ctx.Data.Objects.FindOne(object.KhalimChest1)
	if !found {
		a.ctx.Logger.Debug("Khalim Chest not found")
	}

	err = action.InteractObject(kalimchest1, func() bool {
		chest, _ := a.ctx.Data.Objects.FindOne(object.KhalimChest1)
		return !chest.Selectable
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) openMephistoStairs() error {
	// Use Travincal/Council run to kill the council
	err := Council{}.Run()
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	eye, _ := a.ctx.Data.Inventory.Find("KhalimsEye", item.LocationInventory, item.LocationStash)
	brain, _ := a.ctx.Data.Inventory.Find("KhalimsBrain", item.LocationInventory, item.LocationStash)
	heart, _ := a.ctx.Data.Inventory.Find("KhalimsHeart", item.LocationInventory, item.LocationStash)
	flail, _ := a.ctx.Data.Inventory.Find("KhalimsFlail", item.LocationInventory, item.LocationStash)

	// Combine Khalim's items in the Horadric Cube to create Khalim's Will
	err = action.CubeAddItems(eye, brain, heart, flail)
	if err != nil {
		return err
	}

	err = action.CubeTransmute()
	if err != nil {
		return err
	}

	err = action.UsePortalInTown()
	if err != nil {
		return err
	}

	// Assume we don't have a secondary weapon equipped, so swap to it and equip Khalim's Will
	khalimsWill, found := a.ctx.Data.Inventory.Find("KhalimsWill", item.LocationInventory, item.LocationStash)
	if !found {
		a.ctx.Logger.Info("Khalim's Will not found, aborting mission.")
		return nil
	}

	// Swap weapons, equip Khalim's Will
	a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.SwapWeapons)
	utils.Sleep(1000)
	a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)

	screenPos := ui.GetScreenCoordsForItem(khalimsWill)
	a.ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.ShiftKey)
	utils.Sleep(300)
	a.ctx.HID.PressKey(win.VK_ESCAPE)

	// Interact with the Compelling Orb to open the stairs
	compellingorb, found := a.ctx.Data.Objects.FindOne(object.CompellingOrb)
	if !found {
		a.ctx.Logger.Debug("Khalim Chest not found")
	}

	err = action.InteractObject(compellingorb, func() bool {
		o, _ := a.ctx.Data.Objects.FindOne(object.CompellingOrb)
		return !o.Selectable
	})
	if err != nil {
		return err
	}

	utils.Sleep(1000)
	a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.SwapWeapons)

	time.Sleep(12000)

	// Interact with the stairs to go to Durance of Hate Level 1
	stairsr, found := a.ctx.Data.Objects.FindOne(object.StairSR)
	if !found {
		a.ctx.Logger.Debug("Khalim Chest not found")
	}

	err = action.InteractObject(stairsr, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.DuranceOfHateLevel1
	})
	if err != nil {
		return err
	}

	// Move to Durance of Hate Level 2 and discover the waypoint
	action.MoveToArea(area.DuranceOfHateLevel2)
	action.DiscoverWaypoint()

	return nil
}
