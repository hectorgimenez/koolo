package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"

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

	_, potionFound := a.ctx.Data.Inventory.Find("PotionOfLife", item.LocationInventory)
	q := a.ctx.Data.Quests[quest.Act3TheGoldenBird]
	if (q.Completed() && potionFound) ||
		(!q.HasStatus(quest.StatusQuestNotStarted) && !q.Completed()) {
		a.jadefigurine()
	}

	_, willFound := a.ctx.Data.Inventory.Find("KhalimsWill", item.LocationInventory, item.LocationStash, item.LocationEquipped)

	if willFound {
		a.ctx.Logger.Debug("Khalim's Will found, skipping Khalims Will quests")
		a.openMephistoStairs()
	}

	if a.ctx.Data.Quests[quest.Act3KhalimsWill].Completed() {
		err := NewMephisto(nil).Run()
		if err != nil {
			return err
		}
		if a.ctx.Data.Quests[quest.Act3TheGuardian].Completed() {
			hellgate, found := a.ctx.Data.Objects.FindOne(object.HellGate)
			if !found {
				a.ctx.Logger.Info("Gate to Pandemonium Fortress not found")
			}
			return action.InteractObject(hellgate, func() bool {
				utils.Sleep(500)
				return a.ctx.Data.PlayerUnit.Area == area.ThePandemoniumFortress
			})
		}
	}

	// Find KhalimsEye
	_, found = a.ctx.Data.Inventory.Find("KhalimsEye", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("KhalimsEye found, skipping quest")
	} else if !willFound && !a.ctx.Data.Quests[quest.Act3KhalimsWill].Completed() {
		a.ctx.Logger.Info("KhalimsEye not found, starting quest")
		a.findKhalimsEye()
	}

	// Find KhalimsBrain
	_, found = a.ctx.Data.Inventory.Find("KhalimsBrain", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("KhalimsBrain found, skipping quest")
	} else if !willFound && !a.ctx.Data.Quests[quest.Act3KhalimsWill].Completed() {
		a.ctx.Logger.Info("KhalimsBrain not found, starting quest")
		NewEndugu().Run()
	}

	// Find KhalimsHeart
	_, found = a.ctx.Data.Inventory.Find("KhalimsHeart", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("KhalimsHeart found, skipping quest")
	} else if !willFound && !a.ctx.Data.Quests[quest.Act3KhalimsWill].Completed() {
		a.ctx.Logger.Info("KhalimsHeart not found, starting quest")
		a.findKhalimsHeart()
	}

	// Find KhalimsFlail
	_, found = a.ctx.Data.Inventory.Find("KhalimsFlail", item.LocationInventory, item.LocationStash)
	if found {
		// Trav
		a.openMephistoStairs()
	} else if !willFound && !a.ctx.Data.Quests[quest.Act3KhalimsWill].Completed() {
		a.ctx.Logger.Info("KhalimsFlail not found, starting quest")
		NewTravincal().Run()
	}

	return nil
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

	time.Sleep(4000)

	err = action.MoveToArea(area.SewersLevel2Act3)
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

func (a Leveling) prepareWill() error {
	_, found := a.ctx.Data.Inventory.Find("KhalimsWill", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if !found {

		eye, found := a.ctx.Data.Inventory.Find("KhalimsEye", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.ctx.Logger.Info("Khalim's Eye not found, skipping")
			return nil
		}

		brain, found := a.ctx.Data.Inventory.Find("KhalimsBrain", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.ctx.Logger.Info("Khalim's Brain not found, skipping")
			return nil
		}

		heart, found := a.ctx.Data.Inventory.Find("KhalimsHeart", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.ctx.Logger.Info("Khalim's Heart not found, skipping")
			return nil
		}

		flail, found := a.ctx.Data.Inventory.Find("KhalimsFlail", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.ctx.Logger.Info("Khalim's Flail not found, skipping")
			return nil
		}

		err := action.CubeAddItems(eye, brain, heart, flail)
		if err != nil {
			return err
		}

		err = action.CubeTransmute()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a Leveling) openMephistoStairs() error {
	// Use Travincal/Council run to kill the council
	khalimsWill, kwfound := a.ctx.Data.Inventory.Find("KhalimsWill", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if !kwfound {
		a.prepareWill()
	}

	if !a.ctx.Data.Quests[quest.Act3TheGuardian].Completed() && kwfound {
		a.ctx.Logger.Info("Khalim's Will found!")
		if khalimsWill.Location.LocationType != item.LocationEquipped {
			if a.ctx.Data.ActiveWeaponSlot == 0 {
				a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.SwapWeapons)
			}
			if khalimsWill.Location.LocationType == item.LocationStash {
				a.ctx.Logger.Info("It's in the stash, let's pick it up")

				bank, found := a.ctx.Data.Objects.FindOne(object.Bank)
				if !found {
					a.ctx.Logger.Info("bank object not found")
				}
				utils.Sleep(300)
				err := action.InteractObject(bank, func() bool {
					return a.ctx.Data.OpenMenus.Stash
				})
				if err != nil {
					return err
				}
			}
			if khalimsWill.Location.LocationType == item.LocationInventory && !a.ctx.Data.OpenMenus.Inventory {
				a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)
			}
			screenPos := ui.GetScreenCoordsForItem(khalimsWill)
			a.ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.ShiftKey)
			utils.Sleep(300)
			if a.ctx.Data.ActiveWeaponSlot == 1 {
				a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.SwapWeapons)
			}
			step.CloseAllMenus()
		}
	}

	err := action.WayPoint(area.Travincal)
	if err != nil {
		return err
	}

	// Interact with the Compelling Orb to open the stairs
	compellingorb, found := a.ctx.Data.Objects.FindOne(object.CompellingOrb)
	if !found {
		a.ctx.Logger.Debug("Compelling Orb not found")
	}
	// Swap weapons, equip Khalim's Will

	action.MoveToCoords(compellingorb.Position)
	if a.ctx.Data.ActiveWeaponSlot == 0 {
		a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.SwapWeapons)
	}

	err = action.InteractObject(compellingorb, func() bool {
		o, _ := a.ctx.Data.Objects.FindOne(object.CompellingOrb)
		return !o.Selectable
	})
	if err != nil {
		return err
	}
	utils.Sleep(300)
	if a.ctx.Data.ActiveWeaponSlot == 1 {
		a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.SwapWeapons)
	}

	time.Sleep(12000)

	if a.ctx.Data.Quests[quest.Act3TheBlackenedTemple].Completed() {
		// Interact with the stairs to go to Durance of Hate Level 1
		stairsr, found := a.ctx.Data.Objects.FindOne(object.StairSR)
		if !found {
			a.ctx.Logger.Debug("Stairs to Durance not found")
		}

		err := action.InteractObject(stairsr, func() bool {
			return a.ctx.Data.PlayerUnit.Area == area.DuranceOfHateLevel1
		})
		if err != nil {
			return err
		}

		// Move to Durance of Hate Level 2 and discover the waypoint
		action.MoveToArea(area.DuranceOfHateLevel2)
		err = action.DiscoverWaypoint()
		if err != nil {
			return err
		}
	}

	return nil
}

func (a Leveling) jadefigurine() error {
	_, jadefigureFound := a.ctx.Data.Inventory.Find("AJadeFigurine", item.LocationInventory)
	if jadefigureFound {
		action.InteractNPC(npc.Meshif2)
	}

	_, goldenbirdFound := a.ctx.Data.Inventory.Find("TheGoldenBird", item.LocationInventory)
	if goldenbirdFound {
		// Talk to Alkor
		action.InteractNPC(npc.Alkor)
		action.InteractNPC(npc.Ormus)
		action.InteractNPC(npc.Alkor)
	}
	lifepotion, lifepotfound := a.ctx.Data.Inventory.Find("PotionOfLife", item.LocationInventory)
	if lifepotfound {
		a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)
		screenPos := ui.GetScreenCoordsForItem(lifepotion)
		a.ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
		step.CloseAllMenus()
	}
	return nil
}
