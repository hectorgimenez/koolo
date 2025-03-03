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
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

type Quests struct {
	ctx *context.Status
}

func NewQuests() *Quests {
	return &Quests{
		ctx: context.Get(),
	}
}

func (a Quests) Name() string {
	return string(config.QuestsRun)
}

func (a Quests) Run() error {
	if a.ctx.CharacterCfg.Game.Quests.ClearDen && !a.ctx.Data.Quests[quest.Act1DenOfEvil].Completed() {
		a.clearDenQuest()
	}

	if a.ctx.CharacterCfg.Game.Quests.RescueCain && !a.ctx.Data.Quests[quest.Act1TheSearchForCain].Completed() {
		a.rescueCainQuest()
	}

	if a.ctx.CharacterCfg.Game.Quests.RetrieveHammer && !a.ctx.Data.Quests[quest.Act1ToolsOfTheTrade].Completed() {
		a.retrieveHammerQuest()
	}

	if a.ctx.CharacterCfg.Game.Quests.KillRadament && !a.ctx.Data.Quests[quest.Act2RadamentsLair].Completed() {
		a.killRadamentQuest()
	}

	if a.ctx.CharacterCfg.Game.Quests.GetCube {
		_, found := a.ctx.Data.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if !found {
			a.getHoradricCube()
		}
	}

	if a.ctx.CharacterCfg.Game.Quests.RetrieveBook && !a.ctx.Data.Quests[quest.Act3LamEsensTome].Completed() {
		a.retrieveBookQuest()
	}

	if a.ctx.CharacterCfg.Game.Quests.KillIzual && !a.ctx.Data.Quests[quest.Act4TheFallenAngel].Completed() {
		a.killIzualQuest()
	}

	if a.ctx.CharacterCfg.Game.Quests.KillShenk && !a.ctx.Data.Quests[quest.Act5SiegeOnHarrogath].Completed() {
		a.killShenkQuest()
	}

	if a.ctx.CharacterCfg.Game.Quests.RescueAnya && !a.ctx.Data.Quests[quest.Act5PrisonOfIce].Completed() {
		a.rescueAnyaQuest()
	}
	if a.ctx.CharacterCfg.Game.Quests.KillAncients && !a.ctx.Data.Quests[quest.Act5RiteOfPassage].Completed() {
		a.killAncientsQuest()
	}

	return nil
}

func (a Quests) clearDenQuest() error {
	a.ctx.Logger.Info("Starting Den of Evil Quest...")

	err := action.MoveToArea(area.BloodMoor)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.DenOfEvil)
	if err != nil {
		return err
	}

	action.ClearCurrentLevel(false, data.MonsterAnyFilter())

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Akara)
	if err != nil {
		return err
	}

	step.CloseAllMenus()

	return nil
}

func (a Quests) rescueCainQuest() error {
	a.ctx.Logger.Info("Starting Rescue Cain Quest...")

	err := action.WayPoint(area.RogueEncampment)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.DarkWood)
	if err != nil {
		return err
	}

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

	action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.InifussTree)
	if !found {
		a.ctx.Logger.Debug("InifussTree not found")
	}

	err = action.InteractObject(obj, func() bool {
		updatedObj, found := a.ctx.Data.Objects.FindOne(object.InifussTree)
		if found {
			if !updatedObj.Selectable {
				a.ctx.Logger.Debug("Interacted with InifussTree")
			}
			return !updatedObj.Selectable
		}
		return false
	})
	if err != nil {
		return err
	}
	utils.Sleep(1000)

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Akara)
	if err != nil {
		return err
	}

	step.CloseAllMenus()

	//Reuse Tristram Run actions
	err = NewTristram().Run()
	if err != nil {
		return err
	}

	action.ReturnTown()

	return nil
}

func (a Quests) retrieveHammerQuest() error {
	a.ctx.Logger.Info("Starting Retrieve Hammer Quest...")

	err := action.WayPoint(area.RogueEncampment)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.OuterCloister)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.Barracks)
	if err != nil {
		return err
	}

	err = action.MoveTo(func() (data.Position, bool) {
		for _, o := range a.ctx.Data.Objects {
			if o.Name == object.Malus {
				return o.Position, true
			}
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())

	malus, found := a.ctx.Data.Objects.FindOne(object.Malus)
	if !found {
		a.ctx.Logger.Debug("Malus not found")
	}

	err = action.InteractObject(malus, nil)
	if err != nil {
		return err
	}

	action.ItemPickup(0)

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Charsi)
	if err != nil {
		return err
	}

	step.CloseAllMenus()

	return nil
}

func (a Quests) killRadamentQuest() error {
	var startingPositionAtma = data.Position{
		X: 5138,
		Y: 5057,
	}

	a.ctx.Logger.Info("Starting Kill Radament Quest...")

	err := action.WayPoint(area.LutGholein)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.SewersLevel2Act2)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.SewersLevel3Act2)
	if err != nil {
		return err
	}
	action.Buff()

	// cant find npc.Radament for some reason, using the sparkly chest with ID 355 next him to find him
	err = action.MoveTo(func() (data.Position, bool) {
		for _, o := range a.ctx.Data.Objects {
			if o.Name == object.Name(355) {
				return o.Position, true
			}
		}

		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())

	// Sometimes it moves too far away from the book to pick it up, making sure it moves back to the chest
	err = action.MoveTo(func() (data.Position, bool) {
		for _, o := range a.ctx.Data.Objects {
			if o.Name == object.Name(355) {
				return o.Position, true
			}
		}

		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	// If its still too far away, we're making sure it detects it
	action.ItemPickup(50)

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.MoveToCoords(startingPositionAtma)
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Atma)
	if err != nil {
		return err
	}

	step.CloseAllMenus()
	a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)
	itm, _ := a.ctx.Data.Inventory.Find("BookofSkill")
	screenPos := ui.GetScreenCoordsForItem(itm)
	utils.Sleep(200)
	a.ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
	step.CloseAllMenus()

	return nil
}

func (a Quests) getHoradricCube() error {
	a.ctx.Logger.Info("Starting Retrieve the Cube Quest...")

	err := action.WayPoint(area.LutGholein)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.HallsOfTheDeadLevel2)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.HallsOfTheDeadLevel3)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.HoradricCubeChest)
		if found {
			a.ctx.Logger.Info("Horadric Cube chest found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.HoradricCubeChest)
	if !found {
		return err
	}

	err = action.InteractObject(obj, func() bool {
		updatedObj, found := a.ctx.Data.Objects.FindOne(object.HoradricCubeChest)
		if found {
			if !updatedObj.Selectable {
				a.ctx.Logger.Debug("Interacted with Horadric Cube Chest")
			}
			return !updatedObj.Selectable
		}
		return false
	})
	if err != nil {
		return err
	}

	// Making sure we pick up the cube
	action.ItemPickup(10)

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	return nil
}

func (a Quests) retrieveBookQuest() error {
	a.ctx.Logger.Info("Starting Retrieve Book Quest...")

	err := action.WayPoint(area.KurastDocks)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.KurastBazaar)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.RuinedTemple)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		for _, o := range a.ctx.Data.Objects {
			if o.Name == object.LamEsensTome {
				return o.Position, true
			}
		}

		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())

	tome, found := a.ctx.Data.Objects.FindOne(object.LamEsensTome)
	if !found {
		return err
	}

	err = action.InteractObject(tome, nil)
	if err != nil {
		return err
	}

	// Making sure we pick up the tome
	action.ItemPickup(10)

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Alkor)
	if err != nil {
		return err
	}

	step.CloseAllMenus()

	return nil
}

func (a Quests) killIzualQuest() error {
	a.ctx.Logger.Info("Starting Kill Izual Quest...")

	err := action.MoveToArea(area.OuterSteppes)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.PlainsOfDespair)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		izual, found := a.ctx.Data.NPCs.FindOne(npc.Izual)
		if !found {
			return data.Position{}, false
		}

		return izual.Positions[0], true
	})
	if err != nil {
		return err
	}

	err = a.ctx.Char.KillIzual()
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Tyrael2)
	if err != nil {
		return err
	}

	return nil
}

func (a Quests) killShenkQuest() error {
	var shenkPosition = data.Position{
		X: 3895,
		Y: 5120,
	}

	a.ctx.Logger.Info("Starting Kill Shenk Quest...")

	err := action.WayPoint(area.Harrogath)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.FrigidHighlands)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.BloodyFoothills)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToCoords(shenkPosition)
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(25, data.MonsterAnyFilter())

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Larzuk)
	if err != nil {
		return err
	}

	step.CloseAllMenus()

	return nil
}

func (a Quests) rescueAnyaQuest() error {
	a.ctx.Logger.Info("Starting Rescuing Anya Quest...")

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

	step.CloseAllMenus()
	a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)
	itm, _ := a.ctx.Data.Inventory.Find("ScrollOfResistance")
	screenPos := ui.GetScreenCoordsForItem(itm)
	utils.Sleep(200)
	a.ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
	step.CloseAllMenus()

	return nil
}

func (a Quests) killAncientsQuest() error {
	var ancientsAltar = data.Position{
		X: 10049,
		Y: 12623,
	}

	err := action.WayPoint(area.Harrogath)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.WayPoint(area.TheAncientsWay)
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

	action.MoveToCoords(ancientsAltar)

	utils.Sleep(1000)
	a.ctx.HID.Click(game.LeftButton, 720, 260)
	utils.Sleep(1000)
	a.ctx.HID.PressKey(win.VK_RETURN)
	utils.Sleep(2000)

	action.ClearAreaAroundPlayer(50, data.MonsterEliteFilter())

	action.ReturnTown()

	return nil
}
