package run

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

type Rush struct {
	ctx *context.Status
}

func NewRush() *Rush {
	return &Rush{
		ctx: context.Get(),
	}
}

func (a Rush) Name() string {
	return string(config.RushRun)
}

// Rush Run will need you to manually enter the name and password of the game before starting the bot
// More could be done with some sort of follow, or whisper system
func (a Rush) Run() error {
	if a.ctx.CharacterCfg.Game.Rush.ClearDen && a.ctx.CharacterCfg.Game.Rush.ClearAct1 {
		_ = a.clearDenQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.RescueCain && a.ctx.CharacterCfg.Game.Rush.ClearAct1 {
		_ = a.rescueCainQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.RetrieveHammer && a.ctx.CharacterCfg.Game.Rush.ClearAct1 {
		_ = a.retrieveHammerQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.ClearAct1 {
		err := a.finishAct1Quest()
		if err != nil {
			return err
		}
	}

	if a.ctx.CharacterCfg.Game.Rush.KillRadament && a.ctx.CharacterCfg.Game.Rush.ClearAct2 {
		_ = a.killRadamentQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.GetCube && a.ctx.CharacterCfg.Game.Rush.ClearAct2 {
		err := a.getHoradricCube()
		if err != nil {
			return err
		}
	}

	if a.ctx.CharacterCfg.Game.Rush.ClearAct2 {
		err := a.finishAct2Quest()
		if err != nil {
			return err
		}
	}

	if a.ctx.CharacterCfg.Game.Rush.RetrieveBook && a.ctx.CharacterCfg.Game.Rush.ClearAct3 {
		_ = a.retrieveBookQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.ClearAct3 {
		err := a.finishAct3Quest()
		if err != nil {
			return err
		}
	}

	if a.ctx.CharacterCfg.Game.Rush.KillIzual && a.ctx.CharacterCfg.Game.Rush.ClearAct4 {
		_ = a.killIzualQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.ClearAct4 {
		err := a.finishAct4Quest()
		if err != nil {
			return err
		}
	}

	if a.ctx.CharacterCfg.Game.Rush.KillShenk && a.ctx.CharacterCfg.Game.Rush.ClearAct5 {
		_ = a.killShenkQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.RescueAnya && a.ctx.CharacterCfg.Game.Rush.ClearAct5 {
		_ = a.rescueAnyaQuest()
	}

	if a.ctx.CharacterCfg.Game.Rush.KillAncients && a.ctx.CharacterCfg.Game.Rush.ClearAct5 {
		err := a.killAncientsQuest()
		if err != nil {
			return err
		}
	}

	if a.ctx.CharacterCfg.Game.Rush.ClearAct5 {
		err := a.finishAct5Quest()
		if err != nil {
			return err
		}
	}

	return nil
}

func (a Rush) clearDenQuest() error {
	a.ctx.Logger.Info("Starting Den of Evil Quest...")

	err := action.WayPoint(area.RogueEncampment)
	err = action.WayPoint(area.ColdPlains)
	action.Buff()
	err = action.MoveToArea(area.BloodMoor)
	err = action.MoveToArea(area.DenOfEvil)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(5000)
	action.Buff()
	err = action.ClearCurrentLevel(false, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	return nil
}

func (a Rush) rescueCainQuest() error {
	a.ctx.Logger.Info("Starting Rescue Cain Quest...")

	err := action.WayPoint(area.RogueEncampment)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.DarkWood)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
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

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	err = action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(10000)

	err = action.WayPoint(area.StonyField)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(5000)
	action.Buff()

	// Find the Cairn Stone Alpha
	cairnStone := data.Object{}
	for _, o := range a.ctx.Data.Objects {
		if o.Name == object.CairnStoneAlpha {
			cairnStone = o
		}
	}

	// Move to the cairnStone
	_ = action.MoveToCoords(cairnStone.Position)

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	err = action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	utils.Sleep(25000)

	// Find the portal object
	tristPortal, _ := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)

	// Interact with the portal
	if err = action.InteractObject(tristPortal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.Tristram && a.ctx.Data.AreaData.IsInside(a.ctx.Data.PlayerUnit.Position)
	}); err != nil {
		return err
	}

	// Open a TP if we're the leader
	_ = action.OpenTPIfLeader()

	err = action.ClearCurrentLevel(false, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(10000)

	return nil
}

func (a Rush) retrieveHammerQuest() error {
	a.ctx.Logger.Info("Starting Retrieve Hammer Quest...")

	err := action.WayPoint(area.RogueEncampment)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.OuterCloister)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	action.Buff()

	err = action.MoveToArea(area.Barracks)
	if err != nil {
		return err
	}

	err = action.MoveTo(func() (data.Position, bool) {
		for _, o := range a.ctx.Data.Objects {
			if o.Name == object.Malus {
				return data.Position{X: o.Position.X - 10, Y: o.Position.Y - 10}, true
			}
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	action.Buff()
	utils.Sleep(2000)

	err = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(10000)

	return nil
}

func (a Rush) finishAct1Quest() error {
	a.ctx.Logger.Info("Starting Finish Act1 Quest...")

	err := action.WayPoint(area.CatacombsLevel2)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	action.Buff()
	err = action.MoveToArea(area.CatacombsLevel3)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.CatacombsLevel4)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	utils.Sleep(5000)
	action.Buff()

	err = action.ClearCurrentLevel(false, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	err = action.WayPoint(area.LutGholein)
	if err != nil {
		return err
	}

	utils.Sleep(15000)

	return nil
}

func (a Rush) killRadamentQuest() error {
	a.ctx.Logger.Info("Starting Kill Radament Quest...")

	err := action.WayPoint(area.LutGholein)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.SewersLevel2Act2)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
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
				return data.Position{X: o.Position.X - 10, Y: o.Position.Y - 10}, true
			}
		}

		return data.Position{}, false
	})
	if err != nil {
		err = action.MoveTo(func() (data.Position, bool) {
			radament, found := a.ctx.Data.NPCs.FindOne(npc.Radament)
			if !found {
				return data.Position{}, false
			}

			return radament.Positions[0], true
		})
	}

	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	action.Buff()
	utils.Sleep(10000)

	err = action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(10000)
	return nil
}

func (a Rush) getHoradricCube() error {
	a.ctx.Logger.Info("Starting Retrieve the Cube Quest...")

	err := action.WayPoint(area.LutGholein)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.HallsOfTheDeadLevel2)
	if err != nil {
		return err
	}
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(5000)
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

	err = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
	if err != nil {
		return err
	}
	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)
	return nil
}

func (a Rush) finishAct2Quest() error {
	a.ctx.Logger.Info("Starting Finish Act2 Quest...")
	err := a.retrieveStaffOfKings()
	if err != nil {
		return err
	}
	err = a.retrieveViperAmulet()
	if err != nil {
		return err
	}
	err = a.killSummoner()
	if err != nil {
		return err
	}
	err = a.killDuriel()
	if err != nil {
		return err
	}
	return nil
}

func (a Rush) retrieveStaffOfKings() error {
	err := action.WayPoint(area.FarOasis)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.MaggotLairLevel1)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.MaggotLairLevel2)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.MaggotLairLevel3)
	if err != nil {
		return err
	}

	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
		if found {
			a.ctx.Logger.Info("Staff of Kings chest found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	err = action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())
	if err != nil {
		return err
	}
	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)
	return nil
}

func (a Rush) retrieveViperAmulet() error {
	err := action.WayPoint(area.LostCity)
	if err != nil {
		return err
	}
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.ValleyOfSnakes)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.ClawViperTempleLevel1)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.ClawViperTempleLevel2)
	if err != nil {
		return err
	}

	action.Buff()
	err = action.ClearCurrentLevel(false, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(60000)
	return nil
}

func (a Rush) killSummoner() error {
	err := action.WayPoint(area.ArcaneSanctuary)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	action.Buff()

	// Get the Summoner's position from the cached map data
	areaData := a.ctx.Data.Areas[area.ArcaneSanctuary]
	summonerNPC, found := areaData.NPCs.FindOne(npc.Summoner)
	if !found || len(summonerNPC.Positions) == 0 {
		return err
	}

	smnPos := summonerNPC.Positions[0]

	// Move to the Summoner's position using the static coordinates from map data
	_ = action.MoveToCoords(data.Position{X: smnPos.X, Y: smnPos.Y + 10})
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(10000)
	err = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
	if err != nil {
		return err
	}
	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(10000)
	return nil
}

func (a Rush) killDuriel() error {
	err := action.WayPoint(area.CanyonOfTheMagi)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	action.Buff()

	// Find and move to the real Tal Rasha tomb
	realTalRashaTomb, err := a.findCorrectTomb()
	if err != nil {
		return err
	}

	err = action.MoveToArea(realTalRashaTomb)
	if err != nil {
		return err
	}

	// Wait for area to fully load and get synchronized
	utils.Sleep(500)
	a.ctx.RefreshGameData()

	action.Buff()

	// Find orifice with retry logic
	var orifice data.Object
	var found bool

	for attempts := 0; attempts < maxOrificeAttempts; attempts++ {
		orifice, found = a.ctx.Data.Objects.FindOne(object.HoradricOrifice)
		if found && orifice.Mode == mode.ObjectModeOpened {
			break
		}
		utils.Sleep(orificeCheckDelay)
		a.ctx.RefreshGameData()
	}

	// Move to orifice and clear the area
	err = action.MoveToCoords(orifice.Position)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	err = action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	utils.Sleep(30000)
	// Pre-fight buff
	action.Buff()

	// Find portal and enter Duriel's Lair
	var portal data.Object
	for attempts := 0; attempts < maxOrificeAttempts; attempts++ {
		portal, found = a.ctx.Data.Objects.FindOne(object.DurielsLairPortal)
		if found && portal.Mode == mode.ObjectModeOpened {
			break
		}
		utils.Sleep(orificeCheckDelay)
		a.ctx.RefreshGameData()
	}

	err = action.InteractObject(portal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.DurielsLair
	})
	if err != nil {
		return nil
	}

	// Final refresh before fight
	a.ctx.RefreshGameData()

	utils.Sleep(700)

	err = a.ctx.Char.KillDuriel()
	if err != nil {
		return err
	}
	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	err = action.WayPoint(area.KurastDocks)
	if err != nil {
		return err
	}

	utils.Sleep(60000)
	return nil
}

func (a Rush) findCorrectTomb() (area.ID, error) {
	var realTomb area.ID

	for _, tomb := range talTombs {
		for _, obj := range a.ctx.Data.Areas[tomb].Objects {
			if obj.Name == object.HoradricOrifice {
				realTomb = tomb
				break
			}
		}
	}

	return realTomb, nil
}

func (a Rush) retrieveBookQuest() error {
	a.ctx.Logger.Info("Starting Retrieve Book Quest...")

	err := action.WayPoint(area.KurastDocks)
	if err != nil {
		return err
	}

	err = action.WayPoint(area.KurastBazaar)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	action.Buff()

	err = action.MoveToArea(area.RuinedTemple)
	if err != nil {
		return err
	}

	action.Buff()

	_ = action.MoveTo(func() (data.Position, bool) {
		for _, o := range a.ctx.Data.Objects {
			if o.Name == object.LamEsensTome {
				return o.Position, true
			}
		}

		return data.Position{}, false
	})

	err = action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)

	return nil
}

func (a Rush) finishAct3Quest() error {
	err := a.getEyeOfKhalim()
	if err != nil {
		return err
	}
	err = a.getBrainOfKhalim()
	if err != nil {
		return err
	}
	err = a.getHeartOfKhalim()
	if err != nil {
		return err
	}
	err = a.clearTravincal()
	if err != nil {
		return err
	}
	err = a.killMephisto()
	if err != nil {
		return err
	}
	return nil
}

func (a Rush) getEyeOfKhalim() error {
	err := action.WayPoint(area.SpiderForest)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(5000)
	action.Buff()
	err = action.MoveToArea(area.SpiderCavern)
	if err != nil {
		return err
	}

	action.Buff()
	err = action.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.KhalimChest2)
		if found {
			a.ctx.Logger.Info("Eye of Khalim chest found, moving to that room")
			return chest.Position, true
		} else {
			chest, found = a.ctx.Data.Objects.FindOne(object.KhalimChest1)
			if found {
				a.ctx.Logger.Info("Eye of Khalim chest found, moving to that room")
				return chest.Position, true
			} else {
				chest, found = a.ctx.Data.Objects.FindOne(object.KhalimChest3)
				if found {
					a.ctx.Logger.Info("Eye of Khalim chest found, moving to that room")
					return chest.Position, true
				}
			}
		}
		return data.Position{}, false
	})
	if err != nil {
		err = action.ClearCurrentLevel(true, data.MonsterAnyFilter())
	}

	err = action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	_ = action.Stash(false)
	utils.Sleep(5000)
	return nil
}

func (a Rush) getBrainOfKhalim() error {
	err := action.WayPoint(area.FlayerJungle)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(5000)
	action.Buff()
	err = action.MoveToArea(area.FlayerDungeonLevel1)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.FlayerDungeonLevel2)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.FlayerDungeonLevel3)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.KhalimChest2)
		if found {
			a.ctx.Logger.Info("Brain of Khalim chest found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	err = action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)
	return nil
}

func (a Rush) getHeartOfKhalim() error {
	err := action.WayPoint(area.KurastBazaar)
	if err != nil {
		return err
	}
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(5000)
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

	_ = action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())

	stairs, found := a.ctx.Data.Objects.FindOne(object.Act3SewerStairsToLevel3)
	if !found {
		a.ctx.Logger.Debug("Stairs in Act 3 Sewers not found")
	}

	_ = action.MoveToCoords(stairs.Position)

	_ = action.InteractObject(stairs, func() bool {
		o, _ := a.ctx.Data.Objects.FindOne(object.Act3SewerStairsToLevel3)

		return !o.Selectable
	})

	time.Sleep(3000)

	err = action.MoveToArea(area.SewersLevel2Act3)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.ClearCurrentLevel(true, a.ctx.Data.MonsterFilterAnyReachable())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)
	return nil
}

func (a Rush) clearTravincal() error {
	err := action.WayPoint(area.Travincal)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	action.Buff()
	utils.Sleep(10000)
	action.Buff()

	err = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	councilPosition := a.findCouncilPosition()

	err = action.ClearThroughPath(councilPosition, 30, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ClearAreaAroundPosition(councilPosition, 50, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(30000)
	return nil
}

func (a Rush) findCouncilPosition() data.Position {
	for _, al := range a.ctx.Data.AdjacentLevels {
		if al.Area == area.DuranceOfHateLevel1 {
			return data.Position{
				X: al.Position.X - 1,
				Y: al.Position.Y + 3,
			}
		}
	}

	return data.Position{}
}

func (a Rush) killMephisto() error {
	err := action.WayPoint(area.DuranceOfHateLevel2)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	utils.Sleep(10000)
	action.Buff()
	err = action.MoveToArea(area.DuranceOfHateLevel3)
	if err != nil {
		return err
	}
	action.Buff()

	// Move to the Safe position
	action.MoveToCoords(data.Position{
		X: 17588,
		Y: 8069,
	})
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(10000)

	_ = a.ctx.Char.KillMephisto()

	a.ctx.Logger.Debug("Moving to bridge")
	action.MoveToCoords(data.Position{X: 17588, Y: 8068})
	action.ClearAreaAroundPlayer(40, data.MonsterAnyFilter())
	//Wait for bridge to rise
	utils.Sleep(1000)

	a.ctx.Logger.Debug("Moving to red portal")
	portal, _ := a.ctx.Data.Objects.FindOne(object.HellGate)
	action.MoveToCoords(portal.Position)

	action.InteractObject(portal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.ThePandemoniumFortress
	})
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	err = action.WayPoint(area.ThePandemoniumFortress)
	if err != nil {
		return err
	}

	utils.Sleep(30000)
	return nil
}

func (a Rush) killIzualQuest() error {
	a.ctx.Logger.Info("Starting Kill Izual Quest...")

	_ = action.WayPoint(area.ThePandemoniumFortress)
	_ = action.WayPoint(area.CityOfTheDamned)
	action.OpenTPIfLeader()
	utils.Sleep(3000)
	action.Buff()

	_ = action.MoveToArea(area.PlainsOfDespair)
	_ = action.MoveTo(func() (data.Position, bool) {
		izual, found := a.ctx.Data.NPCs.FindOne(npc.Izual)
		if !found {
			return data.Position{}, false
		}

		return izual.Positions[0], true
	})

	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	_ = a.ctx.Char.KillIzual()

	_ = action.ReturnTown()
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	return nil
}

func (a Rush) finishAct4Quest() error {
	err := action.WayPoint(area.RiverOfFlame)
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	utils.Sleep(5000)
	action.Buff()

	_ = action.MoveToArea(area.ChaosSanctuary)
	_ = action.MoveToCoords(data.Position{X: 7792, Y: 5294})
	action.Buff()
	_ = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
	_ = action.MoveToCoords(data.Position{X: 7792, Y: 5294})
	_ = action.OpenTPIfLeader()
	utils.Sleep(3000)
	action.Buff()
	err = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	//path through towards vizier
	sealGroups := map[string][]object.Name{
		"Vizier":       {object.DiabloSeal4, object.DiabloSeal5}, // Vizier
		"Lord De Seis": {object.DiabloSeal3},                     // Lord De Seis
		"Infector":     {object.DiabloSeal1, object.DiabloSeal2}, // Infector
	}

	// Thanks Go for the lack of ordered maps
	for _, bossName := range []string{"Vizier", "Lord De Seis", "Infector"} {
		a.ctx.Logger.Debug("Heading to", bossName)

		for _, sealID := range sealGroups[bossName] {
			seal, found := a.ctx.Data.Objects.FindOne(sealID)
			if !found {
				return fmt.Errorf("seal not found: %d", sealID)
			}

			err := action.ClearThroughPath(seal.Position, 40, a.ctx.Data.MonsterFilterAnyReachable())
			if err != nil {
				return err
			}

			// Handle the special case for DiabloSeal3
			if sealID == object.DiabloSeal3 && seal.Position.X == 7773 && seal.Position.Y == 5155 {
				if err = action.MoveToCoords(data.Position{X: 7768, Y: 5160}); err != nil {
					return fmt.Errorf("failed to move to bugged seal position: %w", err)
				}
			}

			// Clear everything around the seal
			action.ClearAreaAroundPlayer(20, a.ctx.Data.MonsterFilterAnyReachable())

			//Buff refresh before Infector
			if object.DiabloSeal1 == sealID {
				action.Buff()
			}

			_ = action.InteractObject(seal, func() bool {
				seal, _ = a.ctx.Data.Objects.FindOne(sealID)
				return !seal.Selectable
			})

			// Infector spawns when first seal is enabled
			if object.DiabloSeal1 == sealID {
				if err = a.killSealElite(bossName); err != nil {
					return err
				}
			}
		}

		// Skip Infector boss because was already killed
		if bossName != "Infector" {
			// Wait for the boss to spawn and kill it.
			// Lord De Seis sometimes it's far, and we can not detect him, but we will kill him anyway heading to the next seal
			if err := a.killSealElite(bossName); err != nil && bossName != "Lord De Seis" {
				return err
			}
		}

	}
	action.Buff()

	_ = action.MoveToCoords(diabloSpawnPosition)

	_ = a.ctx.Char.KillDiablo()

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)
	err = action.WayPoint(area.Harrogath)
	if err != nil {
		return err
	}
	utils.Sleep(5000)
	return nil
}

func (a Rush) killSealElite(boss string) error {
	a.ctx.Logger.Debug(fmt.Sprintf("Starting kill sequence for %s", boss))
	startTime := time.Now()
	timeout := 4 * time.Second

	for time.Since(startTime) < timeout {
		for _, m := range a.ctx.Data.Monsters.Enemies(a.ctx.Data.MonsterFilterAnyReachable()) {
			if action.IsMonsterSealElite(m) {
				a.ctx.Logger.Debug(fmt.Sprintf("Seal elite found: %s at position X: %d, Y: %d", m.Name, m.Position.X, m.Position.Y))

				return action.ClearAreaAroundPosition(m.Position, 30, a.ctx.Data.MonsterFilterAnyReachable())
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (a Rush) killShenkQuest() error {
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
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	_ = action.ClearAreaAroundPlayer(50, a.ctx.Data.MonsterFilterAnyReachable())
	utils.Sleep(5000)
	action.Buff()

	err = action.ClearThroughPath(shenkPosition, 40, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	_ = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)

	return nil
}

func (a Rush) rescueAnyaQuest() error {
	a.ctx.Logger.Info("Starting Rescuing Anya Quest...")

	err := action.WayPoint(area.CrystallinePassage)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
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
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}
	_ = action.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	time.Sleep(60000)

	return nil
}

func (a Rush) killAncientsQuest() error {
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
	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	action.Buff()

	err = action.MoveToArea(area.ArreatSummit)
	if err != nil {
		return err
	}
	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	utils.Sleep(15000)
	action.Buff()

	_ = action.MoveToCoords(ancientsAltar)

	utils.Sleep(1000)
	a.ctx.HID.Click(game.LeftButton, 720, 260)
	utils.Sleep(1000)
	a.ctx.HID.PressKey(win.VK_RETURN)
	utils.Sleep(2000)

	_ = action.ClearAreaAroundPlayer(50, data.MonsterEliteFilter())

	err = action.ReturnTown()
	if err != nil {
		return err
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	utils.Sleep(5000)
	return nil
}

func (a Rush) finishAct5Quest() error {
	err := action.WayPoint(area.TheWorldStoneKeepLevel2)
	if err != nil {
		return err
	}

	_ = action.OpenTPIfLeader()
	utils.Sleep(5000)
	action.Buff()
	err = action.MoveToArea(area.TheWorldStoneKeepLevel3)
	if err != nil {
		return err
	}
	err = action.MoveToArea(area.ThroneOfDestruction)
	if err != nil {
		return err
	}

	err = action.MoveToCoords(data.Position{
		X: 15116,
		Y: 5071,
	})
	if err != nil {
		return err
	}

	err = action.OpenTPIfLeader()
	if err != nil {
		return err
	}

	err = action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	// Force rebuff before waves
	action.Buff()

	// Come back to previous position
	err = action.MoveToCoords(baalThronePosition)
	if err != nil {
		return err
	}

	// Handle Baal waves
	lastWave := false
	for !lastWave {
		// Check for last wave
		if _, found := a.ctx.Data.Monsters.FindOne(npc.BaalsMinion, data.MonsterTypeMinion); found {
			lastWave = true
		}

		// Clear current wave
		err = a.clearWave()
		if err != nil {
			return err
		}

		// Return to throne position between waves
		err = action.MoveToCoords(baalThronePosition)
		if err != nil {
			return err
		}
		action.Buff()
		// Small delay to allow next wave to spawn if not last wave
		if !lastWave {
			utils.Sleep(500)
		}
	}

	utils.Sleep(12000)
	action.Buff()

	// Exception: Baal portal has no destination in memory
	baalPortal, _ := a.ctx.Data.Objects.FindOne(object.BaalsPortal)
	err = action.InteractObject(baalPortal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.TheWorldstoneChamber
	})
	if err != nil {
		return err
	}

	_ = action.MoveToCoords(data.Position{X: 15136, Y: 5943})

	return a.ctx.Char.KillBaal()
}

func (a Rush) clearWave() error {
	return a.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies(data.MonsterAnyFilter()) {
			dist := pather.DistanceFromPoint(baalThronePosition, m.Position)
			if d.AreaData.IsWalkable(m.Position) && dist <= 45 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}
