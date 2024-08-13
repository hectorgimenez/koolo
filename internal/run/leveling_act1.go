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

func (a Leveling) act1() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.RogueEncampment {
			return nil
		}

		running = true
		if a.CharacterCfg.Game.Leveling.ExtraFarmBloodMoor {
			if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 3 {
				return a.farmBloodMoor()
			}
		}
		if !d.Quests[quest.Act1DenOfEvil].Completed() {
			return a.denOfEvil()
		}

		if a.CharacterCfg.Game.Leveling.ExtraFarmStonyField {
			if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 7 {
				return a.farmStonyField()
			}
		}

		if lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 13 {
			return a.countess()
		}

		if !a.isCainInTown(d) && !d.Quests[quest.Act1TheSearchForCain].Completed() {
			return a.deckardCain(d)
		}

		return a.andariel(d)
	})
}

func (a Leveling) denOfEvil() []action.Action {
	a.logger.Info("Starting Den of Evil run")
	return []action.Action{
		a.builder.MoveToArea(area.BloodMoor),
		a.builder.Buff(),
		a.builder.MoveToArea(area.DenOfEvil),
		a.builder.Buff(),
		a.builder.ClearArea(false, data.MonsterAnyFilter()),
	}
}
func (a Leveling) farmBloodMoor() []action.Action {
	a.logger.Info("Starting Blood Moor Farm-run")
	return []action.Action{
		a.builder.MoveToArea(area.BloodMoor),
		a.builder.Buff(),
		action.NewStepChain(func(d game.Data) []step.Step {
			for _, l := range d.AdjacentLevels {
				if l.Area == area.DenOfEvil {
					return []step.Step{step.MoveTo(l.Position, step.StopAtDistance(50))}
				}
			}

			return []step.Step{}
		}),
		a.builder.ClearArea(true, data.MonsterAnyFilter()),
	}
}
func (a Leveling) farmStonyField() []action.Action {
	a.logger.Info("Starting Stony Field Farm-run")
	return []action.Action{
		a.builder.WayPoint(area.StonyField),
		a.builder.Buff(),
		a.builder.ClearArea(true, data.MonsterAnyFilter()),
	}
}

//func (a Leveling) bloodRaven() action.Action {
//	return action.NewChain(func(d game.Data) []action.Action {
//		a.logger.Info("Starting Blood Raven quest")
//		return []action.Action{
//			a.builder.WayPoint(area.ColdPlains),
//			a.builder.MoveToArea(area.BurialGrounds),
//			a.char.Buff(),
//			action.NewStepChain(func(d game.Data) []step.Step {
//				for _, l := range d.AdjacentLevels {
//					if l.Area == area.Mausoleum {
//						return []step.Step{step.MoveTo(l.Position, step.StopAtDistance(50))}
//					}
//				}
//
//				return []step.Step{}
//			}),
//			a.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
//				for _, m := range d.Monsters.Enemies() {
//					if pather.DistanceFromMe(d, m.Position) < 3 {
//						return m.UnitID, true
//					}
//
//					if m.Name == npc.BloodRaven {
//						return m.UnitID, true
//					}
//				}
//
//				return 0, false
//			}, nil, step.Distance(5, 15)),
//		}
//	})
//}

func (a Leveling) countess() []action.Action {
	a.logger.Info("Starting Countess run")
	return Countess{baseRun: a.baseRun}.BuildActions()
}

func (a Leveling) deckardCain(d game.Data) (actions []action.Action) {
	a.logger.Info("Rescuing Cain")
	if _, found := d.Inventory.Find("KeyToTheCairnStones"); !found {
		actions = []action.Action{
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
			a.builder.InteractObject(object.InifussTree, func(d game.Data) bool {
				_, found := d.Inventory.Find(scrollOfInifuss)
				return found
			}),
			a.builder.ItemPickup(false, 30),
			a.builder.ReturnTown(),
			a.builder.InteractNPC(
				npc.Akara,
				step.KeySequence(win.VK_ESCAPE),
			),
		}

		// Heal and refill pots
		actions = append(actions,
			a.builder.ReturnTown(),
			a.builder.RecoverCorpse(),
			a.builder.IdentifyAll(false),
			a.builder.Stash(false),
			a.builder.VendorRefill(false, true),
		)

		if a.CharacterCfg.Game.Leveling.EnsurePointsAllocation {
			actions = append(actions,
				a.builder.EnsureStatPoints(),
				a.builder.EnsureSkillPoints(),
			)
		}

		if a.CharacterCfg.Game.Leveling.EnsureKeyBinding {
			actions = append(actions,
				a.builder.EnsureSkillBindings(),
			)
		}

		actions = append(actions,
			a.builder.Heal(),
			a.builder.ReviveMerc(),
			a.builder.HireMerc(),
			a.builder.Repair(),
		)
	}

	// Reuse Tristram Run actions
	actions = append(actions, Tristram{baseRun: a.baseRun}.BuildActions()...)

	return actions
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
		a.builder.UsePortalInTown(),
		a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			return andarielStartingPosition, true
		}),
		a.char.KillAndariel(),
		a.builder.ReturnTown(),
		a.builder.InteractNPC(npc.Warriv, step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)),
	)

	return actions
}

func (a Leveling) isCainInTown(d game.Data) bool {
	_, found := d.Monsters.FindOne(npc.DeckardCain5, data.MonsterTypeNone)

	return found
}
