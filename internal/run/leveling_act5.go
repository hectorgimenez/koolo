package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func (a Leveling) act5() action.Action {
	return action.NewFactory(func(d data.Data) action.Action {
		if d.PlayerUnit.Area != area.Harrogath {
			return nil
		}
		quests := a.builder.GetCompletedQuests(5)
		if quests[4] {
			return action.NewChain(func(d data.Data) (actions []action.Action) {
				a.logger.Info("Starting Baal run...")
				actions = Baal{baseRun: a.baseRun}.BuildActions()
				actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
					if d.PlayerUnit.Area == area.TheWorldstoneChamber && len(d.Monsters.Enemies()) == 0 {
						switch config.Config.Game.Difficulty {
						case difficulty.Normal:
							if d.PlayerUnit.Stats[stat.Level] >= 46 {
								config.Config.Game.Difficulty = difficulty.Nightmare
							}
						case difficulty.Nightmare:
							if d.PlayerUnit.Stats[stat.Level] >= 65 {
								config.Config.Game.Difficulty = difficulty.Hell
							}
						}
					}
					return nil
				}))

				return
			})
		}

		wp, _ := d.Objects.FindOne(object.ExpansionWaypoint)
		if dist := pather.DistanceFromMe(d, wp.Position); dist > 10 {
			return a.builder.MoveToCoords(wp.Position)
		}

		if _, found := d.Monsters.FindOne(npc.Drehya, data.MonsterTypeNone); !found {
			return a.anya()
		}

		return a.ancients()
	})
}

func (a Leveling) anya() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		a.logger.Info("Rescuing Anya...")
		return []action.Action{
			a.builder.WayPoint(area.CrystallinePassage),
			a.builder.MoveToArea(area.FrozenRiver),
			a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
				anya, found := d.NPCs.FindOne(793)
				return anya.Positions[0], found
			}),
			action.NewChain(func(d data.Data) (actions []action.Action) {
				return []action.Action{
					a.builder.ClearAreaAroundPlayer(15),
					action.BuildStatic(func(d data.Data) []step.Step {
						return []step.Step{
							step.InteractObject(object.FrozenAnya, nil),
						}
					}),
				}
			}),
			a.builder.ReturnTown(),
			a.builder.IdentifyAll(false),
			a.builder.Stash(false),
			a.builder.ReviveMerc(),
			a.builder.Repair(),
			a.builder.VendorRefill(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Malah),
				}
			}),
			a.builder.UsePortalInTown(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractObject(object.FrozenAnya, nil),
				}
			}),
			a.builder.ReturnTown(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Malah),
					step.SyncStep(func(d data.Data) error {
						hid.PressKey("esc")
						hid.PressKey(config.Config.Bindings.OpenInventory)
						itm, _ := d.Items.Find("ScrollOfResistance")
						screenPos := ui.GetScreenCoordsForItem(itm)
						hid.MovePointer(screenPos.X, screenPos.Y)
						helper.Sleep(200)
						hid.Click(hid.RightButton)
						hid.PressKey("esc")

						return nil
					}),
				}
			}),
		}
	})
}

func (a Leveling) ancients() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		a.logger.Info("Kill the Ancients...")
		actions = append(actions,
			a.builder.WayPoint(area.TheAncientsWay),
			a.builder.MoveToArea(area.ArreatSummit),
			a.builder.ReturnTown(),
		)

		actions = append(actions, a.builder.InRunReturnTownRoutine()...)

		actions = append(actions,
			a.builder.UsePortalInTown(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractObject(object.AncientsAltar, func(d data.Data) bool {
						if len(d.Monsters.Enemies()) > 0 {
							return true
						}
						hid.Click(hid.LeftButton)
						helper.Sleep(1000)
						return false
					}),
				}
			}),
			a.char.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
				for _, monster := range d.Monsters.Enemies() {
					m, found := d.Monsters.FindOne(monster.Name, monster.Type)
					if found {
						return m.UnitID, true
					}
				}

				return 0, false
			}, nil),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractObject(object.ArreatSummitDoorToWorldstone, func(d data.Data) bool {
						obj, _ := d.Objects.FindOne(object.ArreatSummitDoorToWorldstone)
						return !obj.Selectable
					}),
				}
			}),
			a.builder.MoveToArea(area.TheWorldStoneKeepLevel1),
			a.builder.MoveToArea(area.TheWorldStoneKeepLevel2),
			a.builder.DiscoverWaypoint(),
		)

		return
	})
}
