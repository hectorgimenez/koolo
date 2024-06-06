package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/lxn/win"
)

func (a Leveling) act5() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		if d.PlayerUnit.Area != area.Harrogath {
			return nil
		}

		if d.Quests[quest.Act5RiteOfPassage].Completed() {
			a.logger.Info("Starting Baal run...")
			actions := Baal{baseRun: a.baseRun}.BuildActions()
			return append(actions, action.NewStepChain(func(d game.Data) []step.Step {
				lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
				if d.PlayerUnit.Area == area.TheWorldstoneChamber && len(d.Monsters.Enemies()) == 0 {
					switch a.CharacterCfg.Game.Difficulty {
					case difficulty.Normal:
						if lvl.Value >= 46 {
							a.CharacterCfg.Game.Difficulty = difficulty.Nightmare
						}
					case difficulty.Nightmare:
						if lvl.Value >= 65 {
							a.CharacterCfg.Game.Difficulty = difficulty.Hell
						}
					}
				}
				return nil
			}))
		}

		wp, _ := d.Objects.FindOne(object.ExpansionWaypoint)
		actions := []action.Action{a.builder.MoveToCoords(wp.Position)}
		actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
			if _, found := d.Monsters.FindOne(npc.Drehya, data.MonsterTypeNone); !found {
				return a.anya()
			}

			return a.ancients()
		}))

		return actions
	})
}

func (a Leveling) anya() []action.Action {
	a.logger.Info("Rescuing Anya...")
	return []action.Action{
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
		a.builder.IdentifyAll(false),
		a.builder.Stash(false),
		a.builder.ReviveMerc(),
		a.builder.Repair(),
		a.builder.VendorRefill(false, true),
		a.builder.InteractNPC(npc.Malah),
		a.builder.UsePortalInTown(),
		a.builder.InteractObject(object.FrozenAnya, nil),
		a.builder.ReturnTown(),
		a.builder.Wait(time.Second * 8),
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
}

func (a Leveling) ancients() []action.Action {
	char := a.char.(action.LevelingCharacter)
	a.logger.Info("Kill the Ancients...")
	actions := []action.Action{
		a.builder.WayPoint(area.TheAncientsWay),
		a.builder.MoveToArea(area.ArreatSummit),
		a.builder.ReturnTown(),
	}

	actions = append(actions, a.builder.InRunReturnTownRoutine()...)

	return append(actions,
		a.builder.UsePortalInTown(),
		a.builder.Buff(),
		a.builder.InteractObject(object.AncientsAltar, func(d game.Data) bool {
			if len(d.Monsters.Enemies()) > 0 {
				return true
			}
			a.HID.Click(game.LeftButton, 300, 300)
			helper.Sleep(1000)
			return false
		}),
		char.KillAncients(),
		a.builder.InteractObject(object.ArreatSummitDoorToWorldstone, func(d game.Data) bool {
			obj, _ := d.Objects.FindOne(object.ArreatSummitDoorToWorldstone)
			return !obj.Selectable
		}),
		a.builder.Wait(time.Second*5), // Wait until the door is open
		a.builder.MoveToArea(area.TheWorldStoneKeepLevel1),
		a.builder.MoveToArea(area.TheWorldStoneKeepLevel2),
		a.builder.DiscoverWaypoint(),
	)
}
