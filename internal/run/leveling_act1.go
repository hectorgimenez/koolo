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
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
)

const scrollOfInifuss = "ScrollOfInifuss"
const keyToTheCairnStones = "KeyToTheCairnStones"

func (a Leveling) act1() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.RogueEncampment {
			return nil
		}

		running = true

		// Clear Blood Moor until level 3
		if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 3 {
			return a.bloodMoor()
		}

		// Clear Cold Plains until level 6
		if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 6 {
			return a.coldPlains()
		}

		// Do Den of Evil quest
		if !d.Quests[quest.Act1DenOfEvil].Completed() {
			return a.denOfEvil()
		}

		// Clear Stony Field until level 8
		if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 8 {
			return a.stonyField()
		}

		// Too many depended factors that can go wrong for automatic leveling, removing for now
		// if !a.isCainInTown(d) && !d.Quests[quest.Act1TheSearchForCain].Completed() {
		// 	if a.isCairnKeyInInventory(d) {
		// 		a.logger.Info("Found Cairn key already, running tristram.")
		// 		var actions []action.Action
		// 		actions = append(actions, Tristram{baseRun: a.baseRun}.BuildActions()...)
		// 		return actions
		// 	}
		// 	return a.deckardCain()
		// }
		// do Tristram Runs until level 14
		// if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 14 {
		// 	return a.tristram()
		// }

		// Clear areas up to the Tower before Countess runs until lvl 11
		if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 11 {
			return a.ClearAreasBeforeTower()
		}

		// Do Countess Runs until level 15
		if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 15 {
			return a.countess()
		}

		// Clear Catacomb 2-3 until level 17
		if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 17 {
			return a.catacombs()
		}

		// Kill Andariel to progress to Act 2
		return a.andariel(d)
	})
}

func (a Leveling) bloodMoor() []action.Action {
	a.logger.Info("Starting Blood Moor run")
	return []action.Action{
		a.builder.MoveToArea(area.BloodMoor),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
	}
}

func (a Leveling) coldPlains() []action.Action {
	a.logger.Info("Starting Blood Moor run")
	return []action.Action{
		a.builder.WayPoint(area.ColdPlains),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
	}
}

func (a Leveling) denOfEvil() []action.Action {
	a.logger.Info("Starting Den of Evil Quest")
	return []action.Action{
		a.builder.MoveToArea(area.BloodMoor),
		a.builder.Buff(),
		a.builder.MoveToArea(area.DenOfEvil),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
		a.builder.ReturnTown(),
		a.builder.InteractNPC(
			npc.Akara,
			step.KeySequence(win.VK_ESCAPE),
		),
	}
}

func (a Leveling) stonyField() []action.Action {
	a.logger.Info("Starting Blood Moor run")
	return []action.Action{
		a.builder.WayPoint(area.StonyField),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
	}
}

func (a Leveling) isCainInTown(d game.Data) bool {
	_, found := d.Monsters.FindOne(npc.DeckardCain5, data.MonsterTypeNone)

	return found
}

func (a Leveling) isCairnKeyInInventory(d game.Data) bool {
	_, found := d.Inventory.Find(keyToTheCairnStones)
	return found
}

func (a Leveling) deckardCain() []action.Action {
	a.logger.Info("Starting Rescue Cain Quest")
	var actions []action.Action

	actions = append(actions,
		a.builder.WayPoint(area.RogueEncampment),
		a.builder.WayPoint(area.DarkWood),
		a.builder.Buff(),

		// after clearing the area, go save Cain
		a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			for _, o := range d.Objects {
				if o.Name == object.InifussTree {
					return o.Position, true
				}
			}
			return data.Position{}, false
		}),
		a.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
		a.builder.InteractObject(object.InifussTree, func(d game.Data) bool {
			_, found := d.Inventory.Find(scrollOfInifuss)
			return found
		}),
		a.builder.ItemPickup(true, 0),
		a.builder.ReturnTown(),
		a.builder.InteractNPC(
			npc.Akara,
			step.KeySequence(win.VK_ESCAPE),
		),
	)
	// Reuse Tristram Run actions
	actions = append(actions, Tristram{baseRun: a.baseRun}.BuildActions()...)

	return actions
}

func (a Leveling) tristram() []action.Action {
	a.logger.Info("Starting Tristram run")
	return Tristram{baseRun: a.baseRun}.BuildActions()
}

func (a Leveling) ClearAreasBeforeTower() []action.Action {
	a.logger.Info("Starting Area Clearing before Tower run")
	return []action.Action{
		a.builder.WayPoint(area.DarkWood),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
		a.builder.WayPoint(area.BlackMarsh),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
	}
}

func (a Leveling) countess() []action.Action {
	a.logger.Info("Starting Countess run")
	return Countess{baseRun: a.baseRun}.BuildActions()
}

func (a Leveling) catacombs() []action.Action {
	a.logger.Info("Starting Catacombs run")
	return []action.Action{
		a.builder.WayPoint(area.CatacombsLevel2),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
		a.builder.MoveToArea(area.CatacombsLevel3),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
	}
}

func (a Leveling) andariel(d game.Data) []action.Action {
	a.logger.Info("Starting Andariel run")
	actions := []action.Action{
		a.builder.WayPoint(area.CatacombsLevel2),
		a.builder.Buff(),
		a.builder.MoveToArea(area.CatacombsLevel3),
		a.builder.MoveToArea(area.CatacombsLevel4),
	}
	actions = append(actions, a.builder.ReturnTown()) // Return town to pickup pots and heal, just in case...

	potsToBuy := 4
	if d.MercHPPercent() > 0 {
		potsToBuy = 8
	}

	// Return to the city, ensure we have pots and everything, and get some antidote potions
	actions = append(actions,
		a.builder.ReturnTown(),
		a.builder.VendorRefill(false, true),
		a.builder.BuyAtVendor(npc.Akara, action.VendorItemRequest{
			Item:     "AntidotePotion",
			Quantity: potsToBuy,
			Tab:      4,
		}),
		action.NewStepChain(func(d game.Data) []step.Step {
			return []step.Step{
				step.SyncStep(func(d game.Data) error {
					a.HID.PressKeyBinding(d.KeyBindings.Inventory)
					x := 0
					for _, itm := range d.Inventory.ByLocation(item.LocationInventory) {
						if itm.Name != "AntidotePotion" {
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

	actions = append(actions,
		a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			return andarielStartingPosition, true
		}),
		a.char.KillAndariel(),
		a.builder.ReturnTown(),
		a.builder.InteractNPC(npc.Warriv, step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)),
	)

	return actions
}
