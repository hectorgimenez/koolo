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
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func (a Leveling) act5() error {
	if a.ctx.Data.PlayerUnit.Area != area.Harrogath {
		return nil
	}

	if a.ctx.Data.Quests[quest.Act5RiteOfPassage].Completed() {
		a.ctx.Logger.Info("Starting Baal run...")
		Baal{}.Run()

		lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
		if a.ctx.Data.PlayerUnit.Area == area.TheWorldstoneChamber && len(a.ctx.Data.Monsters.Enemies()) == 0 {
			switch a.ctx.CharacterCfg.Game.Difficulty {
			case difficulty.Normal:
				if lvl.Value >= 46 {
					a.ctx.CharacterCfg.Game.Difficulty = difficulty.Nightmare
				}
			case difficulty.Nightmare:
				if lvl.Value >= 65 {
					a.ctx.CharacterCfg.Game.Difficulty = difficulty.Hell
				}
			}
		}
		return nil

	}

	wp, _ := a.ctx.Data.Objects.FindOne(object.ExpansionWaypoint)
	action.MoveToCoords(wp.Position)

	if _, found := a.ctx.Data.Monsters.FindOne(npc.Drehya, data.MonsterTypeNone); !found {
		a.anya()
	}

	err := a.ancients()
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) anya() error {
	a.ctx.Logger.Info("Rescuing Anya...")

	err := action.WayPoint(area.CrystallinePassage)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.FrozenRiver)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		anya, found := a.ctx.Data.NPCs.FindOne(793)
		return anya.Positions[0], found
	})
	if err != nil {
		return err
	}

	err = action.MoveTo(func() (data.Position, bool) {
		anya, found := a.ctx.Data.Objects.FindOne(object.FrozenAnya)
		return anya.Position, found
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	anya, found := a.ctx.Data.Objects.FindOne(object.FrozenAnya)
	if !found {
		a.ctx.Logger.Debug("Frozen Anya not found")
	}

	err = action.InteractObject(anya, nil)
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	action.IdentifyAll(false)
	action.Stash(false)
	action.ReviveMerc()
	action.Repair()
	action.VendorRefill(false, true)

	err = action.InteractNPC(npc.Malah)
	if err != nil {
		return err
	}

	err = action.UsePortalInTown()
	if err != nil {
		return err
	}

	err = action.InteractObject(anya, nil)
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	time.Sleep(8000)

	err = action.InteractNPC(npc.Malah)
	if err != nil {
		return err
	}

	a.ctx.HID.PressKey(win.VK_ESCAPE)
	a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)
	itm, _ := a.ctx.Data.Inventory.Find("ScrollOfResistance")
	screenPos := ui.GetScreenCoordsForItem(itm)
	utils.Sleep(200)
	a.ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
	a.ctx.HID.PressKey(win.VK_ESCAPE)

	return nil
}

func (a Leveling) ancients() error {
	char := a.ctx.Char.(context.LevelingCharacter)

	a.ctx.Logger.Info("Kill the Ancients...")

	err := action.WayPoint(area.TheAncientsWay)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.ArreatSummit)
	if err != nil {
		return err
	}
	action.Buff()

	action.ReturnTown()
	action.InRunReturnTownRoutine()
	action.UsePortalInTown()
	action.Buff()

	ancientsaltar, found := a.ctx.Data.Objects.FindOne(object.AncientsAltar)
	if !found {
		a.ctx.Logger.Debug("Ancients Altar not found")
	}

	err = action.InteractObject(ancientsaltar, func() bool {
		if len(a.ctx.Data.Monsters.Enemies()) > 0 {
			return true
		}
		a.ctx.HID.Click(game.LeftButton, 300, 300)
		utils.Sleep(1000)
		return false
	})
	if err != nil {
		return err
	}

	err = char.KillAncients()
	if err != nil {
		return err
	}

	summitdoor, found := a.ctx.Data.Objects.FindOne(object.ArreatSummitDoorToWorldstone)
	if !found {
		a.ctx.Logger.Debug("Worldstone Door not found")
	}

	err = action.InteractObject(summitdoor, func() bool {
		obj, _ := a.ctx.Data.Objects.FindOne(object.ArreatSummitDoorToWorldstone)
		return !obj.Selectable
	})
	if err != nil {
		return err
	}

	time.Sleep(5000)

	err = action.MoveToArea(area.TheWorldStoneKeepLevel1)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.TheWorldStoneKeepLevel2)
	if err != nil {
		return err
	}

	action.DiscoverWaypoint()

	return nil
}
