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
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
)

func (a Leveling) act2() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.LutGholein {
			return nil
		}

		running = true
		// Find Horadric Cube
		_, found := d.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if found {
			a.logger.Info("Horadric Cube found, skipping quest")
		} else {
			a.logger.Info("Horadric Cube not found, starting quest")
			return a.findHoradricCube()
		}

		if d.Quests[quest.Act2TheSummoner].Completed() {
			// Try to get level 21 before moving to Duriel and Act3

			if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 18 {
				return TalRashaTombs{a.baseRun}.BuildActions()
			}

			return a.duriel(d.Quests[quest.Act2TheHoradricStaff].Completed(), d)
		}

		_, horadricStaffFound := d.Inventory.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)

		// Find Staff of Kings
		_, found = d.Inventory.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if found || horadricStaffFound {
			a.logger.Info("StaffOfKings found, skipping quest")
		} else {
			a.logger.Info("StaffOfKings not found, starting quest")
			return a.findStaff()
		}

		// Find Amulet
		_, found = d.Inventory.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if found || horadricStaffFound {
			a.logger.Info("Amulet of the Viper found, skipping quest")
		} else {
			a.logger.Info("Amulet of the Viper not found, starting quest")
			return a.findAmulet()
		}

		// Summoner
		a.logger.Info("Starting summoner quest")
		return a.summoner()
	})
}

//func (a Leveling) radament() action.Action {
//	return action.NewChain(func(d game.Data) (actions []action.Action) {
//		actions = append(actions,
//			a.builder.WayPoint(area.SewersLevel2Act2),
//			a.builder.MoveToArea(area.SewersLevel3Act2),
//		)
//
//		// TODO: Find Radament (use 355 object to locate him)
//		return
//	})
//}

func (a Leveling) findHoradricCube() []action.Action {
	return []action.Action{
		a.builder.WayPoint(area.HallsOfTheDeadLevel2),
		a.builder.MoveToArea(area.HallsOfTheDeadLevel3),
		a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			a.logger.Info("Horadric Cube chest found, moving to that room")
			chest, found := d.Objects.FindOne(object.HoradricCubeChest)

			return chest.Position, found
		}),
		a.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
		a.builder.InteractObject(object.HoradricCubeChest, func(d game.Data) bool {
			chest, _ := d.Objects.FindOne(object.HoradricCubeChest)
			return !chest.Selectable
		}),
		a.builder.ItemPickup(true, 10),
	}
}

func (a Leveling) findStaff() []action.Action {
	return []action.Action{
		a.builder.WayPoint(area.FarOasis),
		a.builder.MoveToArea(area.MaggotLairLevel1),
		a.builder.MoveToArea(area.MaggotLairLevel2),
		a.builder.MoveToArea(area.MaggotLairLevel3),
		a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			a.logger.Info("Staff Of Kings chest found, moving to that room")
			chest, found := d.Objects.FindOne(object.StaffOfKingsChest)

			return chest.Position, found
		}),
		a.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
		a.builder.InteractObject(object.StaffOfKingsChest, func(d game.Data) bool {
			chest, _ := d.Objects.FindOne(object.StaffOfKingsChest)
			return !chest.Selectable
		}),
		a.builder.ItemPickup(true, 10),
	}
}

func (a Leveling) findAmulet() []action.Action {
	return []action.Action{
		a.builder.WayPoint(area.LostCity),
		a.builder.MoveToArea(area.ValleyOfSnakes),
		a.builder.MoveToArea(area.ClawViperTempleLevel1),
		a.builder.MoveToArea(area.ClawViperTempleLevel2),
		a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			a.logger.Info("Altar found, moving closer")
			chest, found := d.Objects.FindOne(object.TaintedSunAltar)

			return chest.Position, found
		}),
		a.builder.InteractObject(object.TaintedSunAltar, func(d game.Data) bool {
			chest, _ := d.Objects.FindOne(object.TaintedSunAltar)
			return !chest.Selectable
		}),
		a.builder.ItemPickup(true, 10),
	}
}

func (a Leveling) summoner() []action.Action {
	actions := Summoner{baseRun: a.baseRun}.BuildActions()

	// Try to use the portal and discover the waypoint
	actions = append(actions,
		a.builder.InteractObject(object.YetAnotherTome, func(d game.Data) bool {
			_, found := d.Objects.FindOne(object.PermanentTownPortal)
			return found
		}),
		a.builder.InteractObject(object.PermanentTownPortal, func(d game.Data) bool {
			return d.PlayerUnit.Area == area.CanyonOfTheMagi
		}),
		a.builder.DiscoverWaypoint(),
		a.builder.ReturnTown(),
		a.builder.InteractNPC(npc.Atma, step.KeySequence(win.VK_ESCAPE)),
	)

	return actions
}

func (a Leveling) prepareStaff() action.Action {
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		horadricStaff, found := d.Inventory.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if found {
			a.logger.Info("Horadric Staff found!")
			if horadricStaff.Location.LocationType == item.LocationStash {
				a.logger.Info("It's in the stash, let's pickup it (not done yet)")

				return []action.Action{
					a.builder.InteractObject(object.Bank, func(d game.Data) bool {
						return d.OpenMenus.Stash
					},
						step.SyncStep(func(d game.Data) error {
							screenPos := a.UIManager.GetScreenCoordsForItem(horadricStaff)

							a.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
							helper.Sleep(300)
							a.HID.PressKey(win.VK_ESCAPE)
							return nil
						}),
					),
				}
			}

			return nil
		}

		staff, found := d.Inventory.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.logger.Info("Staff of Kings not found, skipping")
			return nil
		}

		amulet, found := d.Inventory.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.logger.Info("AmuletOfTheViper not found, skipping")
			return nil
		}

		return []action.Action{
			a.builder.CubeAddItems(staff, amulet),
			a.builder.CubeTransmute(),
		}
	})
}

func (a Leveling) duriel(staffAlreadyUsed bool, d game.Data) (actions []action.Action) {
	a.logger.Info("Starting Duriel....")

	var realTomb area.ID
	for _, tomb := range talRashaTombs {
		_, _, objects, _ := a.Reader.CachedMapData.NPCsExitsAndObjects(data.Position{}, tomb)
		for _, obj := range objects {
			if obj.Name == object.HoradricOrifice {
				realTomb = tomb
				break
			}
		}
	}

	if realTomb == 0 {
		a.logger.Info("Could not find the real tomb :(")
		return nil
	}

	if !staffAlreadyUsed {
		actions = append(actions, a.prepareStaff())
	}

	// Move close to the Horadric Orifice
	actions = append(actions,
		a.builder.WayPoint(area.CanyonOfTheMagi),
		a.builder.Buff(),
		a.builder.MoveToArea(realTomb),
		a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			orifice, _ := d.Objects.FindOne(object.HoradricOrifice)

			return orifice.Position, true
		}),
	)

	// If staff has not been used, then put it in the orifice and wait for the entrance to open
	if !staffAlreadyUsed {
		actions = append(actions,
			a.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
			a.builder.InteractObject(object.HoradricOrifice, func(d game.Data) bool {
				return d.OpenMenus.Anvil
			},
				step.SyncStep(func(d game.Data) error {
					staff, _ := d.Inventory.Find("HoradricStaff", item.LocationInventory)

					screenPos := a.UIManager.GetScreenCoordsForItem(staff)

					a.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
					helper.Sleep(300)
					a.HID.Click(game.LeftButton, ui.AnvilCenterX, ui.AnvilCenterY)
					helper.Sleep(500)
					a.HID.Click(game.LeftButton, ui.AnvilBtnX, ui.AnvilBtnY)
					helper.Sleep(20000)

					return nil
				}),
			),
		)
	}

	potsToBuy := 4
	if d.MercHPPercent() > 0 {
		potsToBuy = 8
	}

	// Return to the city, ensure we have pots and everything, and get some thawing potions
	actions = append(actions,
		a.builder.ReturnTown(),
		a.builder.ReviveMerc(),
		a.builder.VendorRefill(false, true),
		a.builder.BuyAtVendor(npc.Lysander, action.VendorItemRequest{
			Item:     "ThawingPotion",
			Quantity: potsToBuy,
			Tab:      4,
		}),
		action.NewStepChain(func(d game.Data) []step.Step {
			return []step.Step{
				step.SyncStep(func(d game.Data) error {
					a.HID.PressKeyBinding(d.KeyBindings.Inventory)
					x := 0
					for _, itm := range d.Inventory.ByLocation(item.LocationInventory) {
						if itm.Name != "ThawingPotion" {
							continue
						}

						pos := a.UIManager.GetScreenCoordsForItem(itm)
						helper.Sleep(500)

						if x > 3 {
							a.HID.Click(game.LeftButton, pos.X, pos.Y)
							helper.Sleep(300)
							if d.LegacyGraphics {
								a.HID.Click(game.LeftButton, ui.MercAvatarPositionXClassic, ui.MercAvatarPositionYClassic)
							} else {
								a.HID.Click(game.LeftButton, ui.MercAvatarPositionX, ui.MercAvatarPositionY)
							}
						} else {
							a.HID.Click(game.RightButton, pos.X, pos.Y)
						}
						x++
					}

					a.HID.PressKey(win.VK_ESCAPE)
					return nil
				}),
			}
		}),
		a.builder.UsePortalInTown(),
		a.builder.Buff(),
	)

	return append(actions,
		action.NewChain(func(d game.Data) []action.Action {
			_, found := d.Objects.FindOne(object.DurielsLairPortal)
			if found {
				return []action.Action{a.builder.InteractObject(object.DurielsLairPortal, func(d game.Data) bool {
					return d.PlayerUnit.Area == area.DurielsLair
				})}
			}
			return nil
		}),
		a.char.KillDuriel(),
		a.builder.ItemPickup(true, 30),
		a.builder.MoveToCoords(data.Position{
			X: 22577,
			Y: 15613,
		}),
		a.builder.InteractNPCWithCheck(npc.Tyrael, func(d game.Data) bool {
			obj, found := d.Objects.FindOne(object.TownPortal)
			if found && pather.DistanceFromMe(d, obj.Position) < 10 {
				return true
			}

			return false
		}),
		a.builder.ReturnTown(),
		a.builder.MoveToCoords(data.Position{
			X: 5092,
			Y: 5144,
		}),
		a.builder.InteractNPC(npc.Jerhyn, step.KeySequence(win.VK_ESCAPE)),
		a.builder.MoveToCoords(data.Position{
			X: 5195,
			Y: 5060,
		}),
		a.builder.InteractNPC(npc.Meshif, step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)),
	)
}
