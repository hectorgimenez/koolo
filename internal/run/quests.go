package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/lxn/win"
)

type Quests struct {
	baseRun
}

func (a Quests) Name() string {
	return string(config.QuestsRun)
}

func (a Quests) BuildActions() []action.Action {
	//var actions []action.Action

	return []action.Action{
		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.ClearDen && !d.Quests[quest.Act1DenOfEvil].Completed() {
				return []action.Action{a.clearDenQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.RescueCain && !d.Quests[quest.Act1TheSearchForCain].Completed() {
				return []action.Action{a.rescueCainQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.RetrieveHammer && !d.Quests[quest.Act1ToolsOfTheTrade].Completed() {
				return []action.Action{a.retrieveHammerQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.KillRadament && !d.Quests[quest.Act2RadamentsLair].Completed() {
				return []action.Action{a.killRadamentQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.GetCube {
				_, found := d.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
				if !found {
					return []action.Action{a.getHoradricCube()}
				}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.RetrieveBook && !d.Quests[quest.Act3LamEsensTome].Completed() {
				return []action.Action{a.retrieveBookQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.KillIzual && !d.Quests[quest.Act4TheFallenAngel].Completed() {
				return []action.Action{a.killIzualQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.KillShenk && !d.Quests[quest.Act5SiegeOnHarrogath].Completed() {
				return []action.Action{a.killShenkQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.RescueAnya && !d.Quests[quest.Act5PrisonOfIce].Completed() {
				return []action.Action{a.rescueAnyaQuest()}
			}
			return nil
		}),

		action.NewChain(func(d game.Data) []action.Action {
			if a.CharacterCfg.Game.Quests.KillAncients && !d.Quests[quest.Act5RiteOfPassage].Completed() {
				return []action.Action{a.killAncientsQuest()}
			}
			return nil
		}),
	}
}

func (a Quests) openQuestLog() action.Action {
	return action.NewStepChain(func(d game.Data) []step.Step {
		return []step.Step{
			step.SyncStep(func(g game.Data) error {
				a.HID.PressKeyBinding(d.KeyBindings.QuestLog)
				return nil
			}),
			step.Wait(time.Second * 1),
			step.SyncStep(func(g game.Data) error {
				a.HID.PressKey(win.VK_ESCAPE)
				return nil
			}),
		}
	})
}

func (a Quests) clearDenQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.RogueEncampment),
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
	})
}

func (a Quests) rescueCainQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action
		actions = append(actions,
			a.builder.WayPoint(area.RogueEncampment),
			a.builder.WayPoint(area.DarkWood),
			a.builder.Buff(),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.InifussTree {
						return o.Position, true
					}
				}
				return data.Position{}, false
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
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
	})
}

func (a Quests) retrieveHammerQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.RogueEncampment),
			a.builder.WayPoint(area.OuterCloister),
			a.builder.Buff(),
			a.builder.MoveToArea(area.Barracks),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.Malus {
						return o.Position, true
					}
				}
				return data.Position{}, false
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.InteractObject(object.Malus, nil),
			a.builder.ItemPickup(false, 30),
			a.builder.ReturnTown(),
			a.builder.InteractNPC(
				npc.Charsi,
				step.KeySequence(win.VK_ESCAPE),
			),
		}
	})
}

func (a Quests) killRadamentQuest() action.Action {
	var startingPositionAtma = data.Position{
		X: 5138,
		Y: 5057,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.LutGholein),
			a.builder.WayPoint(area.SewersLevel2Act2),
			a.builder.Buff(),
			a.builder.MoveToArea(area.SewersLevel3Act2),
			// cant find npc.Radament for some reason, using the sparkly chest with ID 355 next him to find him
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.Name(355) {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			a.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
			// Sometimes it moves too far away from the book to pick it up, making sure it moves back to the chest
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.Name(355) {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			a.builder.ItemPickup(true, 50),
			a.builder.ReturnTown(),
			a.builder.MoveToCoords(startingPositionAtma),
			a.builder.InteractNPC(npc.Atma,
				step.SyncStep(func(d game.Data) error {
					a.HID.PressKey(win.VK_ESCAPE)
					a.HID.PressKeyBinding(d.KeyBindings.Inventory)
					itm, _ := d.Inventory.Find("BookofSkill")
					screenPos := a.UIManager.GetScreenCoordsForItem(itm)
					helper.Sleep(200)
					a.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
					a.HID.PressKey(win.VK_ESCAPE)

					return nil
				}),
			),
		}
	})
}

func (a Quests) getHoradricCube() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.LutGholein),
			a.builder.WayPoint(area.HallsOfTheDeadLevel2),
			a.builder.Buff(),
			a.builder.MoveToArea(area.HallsOfTheDeadLevel3),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				chest, found := d.Objects.FindOne(object.HoradricCubeChest)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.InteractObject(object.HoradricCubeChest, func(d game.Data) bool {
				chest, _ := d.Objects.FindOne(object.HoradricCubeChest)
				return !chest.Selectable
			}),
			a.builder.ItemPickup(true, 10),
			a.builder.ReturnTown(),
		}
	})
}

func (a Quests) retrieveBookQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.KurastDocks),
			a.builder.WayPoint(area.KurastBazaar),
			a.builder.Buff(),
			a.builder.MoveToArea(area.RuinedTemple),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.LamEsensTome {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			a.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
			a.builder.InteractObject(object.LamEsensTome, nil),
			a.builder.ItemPickup(true, 30),
			a.builder.ReturnTown(),
			a.builder.InteractNPC(
				npc.Alkor,
				step.KeySequence(win.VK_ESCAPE),
			),
		}
	})
}

func (a Quests) killIzualQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.ThePandemoniumFortress),
			a.builder.MoveToArea(area.OuterSteppes),
			a.builder.Buff(),
			a.builder.MoveToArea(area.PlainsOfDespair),

			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				if izual, found := d.NPCs.FindOne(npc.Izual); found {
					return izual.Positions[0], true
				}
				return data.Position{}, false
			}, step.StopAtDistance(50)),

			a.char.KillIzual(),
			a.builder.ReturnTown(),
			a.builder.InteractNPC(
				npc.Tyrael2,
				step.KeySequence(win.VK_ESCAPE),
			),
		}
	})
}

func (a Quests) killShenkQuest() action.Action {
	var shenkPosition = data.Position{
		X: 3895,
		Y: 5120,
	}
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.Harrogath),
			a.builder.WayPoint(area.FrigidHighlands),
			a.builder.Buff(),
			a.builder.MoveToArea(area.BloodyFoothills),
			a.builder.MoveToCoords(shenkPosition),
			a.builder.ClearAreaAroundPlayer(25, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
			a.builder.InteractNPC(
				npc.Larzuk,
				step.KeySequence(win.VK_ESCAPE),
			),
		}
	})
}

func (a Quests) rescueAnyaQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.Harrogath),
			a.builder.WayPoint(area.CrystallinePassage),
			a.builder.MoveToArea(area.FrozenRiver),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				anya, found := d.NPCs.FindOne(793)
				return anya.Positions[0], found
			}),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				anya, found := d.Objects.FindOne(object.FrozenAnya)
				return anya.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
			a.builder.InteractObject(object.FrozenAnya, nil),
			a.builder.ReturnTown(),
			a.builder.InteractNPC(npc.Malah),
			a.builder.UsePortalInTown(),
			a.builder.InteractObject(object.FrozenAnya, nil),
			a.builder.ReturnTown(),
			a.builder.Wait(time.Second * 4),
			a.builder.InteractNPC(npc.Malah,
				step.SyncStep(func(d game.Data) error {
					a.HID.PressKey(win.VK_ESCAPE)
					a.HID.PressKeyBinding(d.KeyBindings.Inventory)
					itm, _ := d.Inventory.Find("ScrollOfResistance")
					screenPos := a.UIManager.GetScreenCoordsForItem(itm)
					helper.Sleep(200)
					a.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
					a.HID.PressKey(win.VK_ESCAPE)

					return nil
				}),
			),
		}
	})
}

func (a Quests) killAncientsQuest() action.Action {
	var ancientsAltar = data.Position{
		X: 10049,
		Y: 12623,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.Harrogath),
			a.builder.WayPoint(area.TheAncientsWay),
			a.builder.Buff(),
			a.builder.MoveToArea(area.ArreatSummit),
			a.builder.Buff(),
			a.builder.MoveToCoords(ancientsAltar),

			action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{
					step.SyncStep(func(g game.Data) error {
						helper.Sleep(1000)
						a.HID.Click(game.LeftButton, 720, 260)
						helper.Sleep(1000)
						a.HID.PressKey(win.VK_RETURN)
						helper.Sleep(2000)
						return nil
					}),
				}
			}),

			a.builder.ClearAreaAroundPlayer(50, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
		}
	})
}
