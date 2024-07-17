package run

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type Rushing struct {
	baseRun
}

func (a Rushing) Name() string {
	return string(config.RushingRun)
}

func (a Rushing) BuildActions() []action.Action {
	return []action.Action{
		a.rushAct1(),
		a.rushAct2(),
		a.rushAct3(),
		a.rushAct4(),
		a.rushAct5(),
	}
}

// Waypoints
func (a Rushing) GiveAct1WPs() action.Action {
	areas := []area.ID{
		area.StonyField,
		area.DarkWood,
		area.BlackMarsh,
		area.InnerCloister,
		area.OuterCloister,
		area.CatacombsLevel2,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		actions := []action.Action{}

		for _, areaID := range areas {
			actions = append(actions,
				a.builder.WayPoint(areaID),
				a.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
				a.builder.OpenTP(),
				a.builder.Wait(time.Second*5),
			)
		}

		return actions
	})
}

func (a Rushing) GiveAct2WPs() action.Action {
	areas := []area.ID{
		area.SewersLevel2Act2,
		area.HallsOfTheDeadLevel2,
		area.FarOasis,
		area.LostCity,
		area.ArcaneSanctuary,
		area.CanyonOfTheMagi,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		actions := []action.Action{}

		for _, areaID := range areas {
			actions = append(actions,
				a.builder.WayPoint(areaID),
				a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
				a.builder.OpenTP(),
				a.builder.Wait(time.Second*5),
			)
		}

		return actions
	})
}

func (a Rushing) GiveAct3WPs() action.Action {
	areas := []area.ID{
		area.SpiderForest,
		area.GreatMarsh,
		area.FlayerJungle,
		area.LowerKurast,
		area.KurastBazaar,
		area.UpperKurast,
		area.Travincal,
		area.DuranceOfHateLevel2,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		actions := []action.Action{}

		for _, areaID := range areas {
			actions = append(actions,
				a.builder.WayPoint(areaID),
				a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
				a.builder.OpenTP(),
				a.builder.Wait(time.Second*5),
			)
		}

		return actions
	})
}

func (a Rushing) GiveAct4WPs() action.Action {
	areas := []area.ID{
		area.OuterSteppes,
		area.PlainsOfDespair,
		area.RiverOfFlame,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		actions := []action.Action{}

		for _, areaID := range areas {
			actions = append(actions,
				a.builder.WayPoint(areaID),
				a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
				a.builder.OpenTP(),
				a.builder.Wait(time.Second*5),
			)
		}

		return actions
	})
}

func (a Rushing) GiveAct5WPs() action.Action {
	areas := []area.ID{
		area.BloodyFoothills,
		area.CrystallinePassage,
		area.HallsOfPain,
		area.TheAncientsWay,
		area.TheWorldStoneKeepLevel2,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		actions := []action.Action{}

		for _, areaID := range areas {
			actions = append(actions,
				a.builder.WayPoint(areaID),
				a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
				a.builder.OpenTP(),
				a.builder.Wait(time.Second*5),
			)
		}

		return actions
	})
}

// Act 1
func (a Rushing) rushAct1() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {

		actions := []action.Action{
			a.builder.VendorRefill(true, true),
		}

		if a.CharacterCfg.Game.Rushing.GiveWPsA1 {
			actions = append(actions, a.GiveAct1WPs())
		}

		if a.CharacterCfg.Game.Rushing.ClearDen {
			actions = append(actions, a.clearDenQuest())
		}

		if a.CharacterCfg.Game.Rushing.RescueCain {
			actions = append(actions, a.rescueCainQuest())
		}

		if a.CharacterCfg.Game.Rushing.RetrieveHammer {
			actions = append(actions, a.retrieveHammerQuest())
		}

		actions = append(actions,
			a.killAandarielQuest(),
		)

		return actions
	})
}

func (a Rushing) clearDenQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.MoveToArea(area.BloodMoor),
			a.builder.Buff(),
			a.builder.MoveToArea(area.DenOfEvil),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ClearArea(false, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) rescueCainQuest() action.Action {
	var gimpCage = data.Position{
		X: 25140,
		Y: 5145,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			// Go to Tree
			a.builder.WayPoint(area.DarkWood),
			a.builder.OpenTP(),
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
			a.builder.OpenTP(),
			a.builder.ReturnTown(),

			// Go to Stones
			a.builder.WayPoint(area.StonyField),
			a.builder.OpenTP(),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.CairnStoneAlpha {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),

			// Wait for Tristram portal and enter
			action.NewChain(func(d game.Data) []action.Action {
				_, found := d.Objects.FindOne(object.PermanentTownPortal)
				if found {
					return []action.Action{
						a.builder.InteractObject(object.PermanentTownPortal, func(d game.Data) bool {
							return d.PlayerUnit.Area == area.Tristram
						}),
					}
				}
				return nil
			}),
			a.builder.MoveToArea(area.Tristram),
			a.builder.MoveToCoords(gimpCage),
			a.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) retrieveHammerQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.OuterCloister),
			a.builder.OpenTP(),
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
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killAandarielQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.CatacombsLevel2),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.CatacombsLevel3),
			a.builder.MoveToArea(area.CatacombsLevel4),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.MoveToCoords(andarielStartingPosition),
			a.char.KillAndariel(),
			a.builder.ReturnTown(),
			a.builder.WayPoint(area.LutGholein),
		}
	})
}

// Act 2
func (a Rushing) rushAct2() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {

		actions := []action.Action{
			a.builder.VendorRefill(true, true),
		}

		if a.CharacterCfg.Game.Rushing.GiveWPsA2 {
			actions = append(actions, a.GiveAct2WPs())
		}

		if a.CharacterCfg.Game.Rushing.KillRadament {
			actions = append(actions, a.killRadamentQuest())
		}

		if a.CharacterCfg.Game.Difficulty == "normal" {
			actions = append(actions, a.getHoradricCube())
		}

		actions = append(actions,
			a.getStaff(),
			a.getAmulet(),
			a.killSummonerQuest(),
			a.killDurielQuest(),
		)

		return actions
	})
}

func (a Rushing) killRadamentQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.SewersLevel2Act2),
			a.builder.OpenTP(),
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
			}, step.StopAtDistance(50)),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) getHoradricCube() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.HallsOfTheDeadLevel2),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.HallsOfTheDeadLevel3),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				a.logger.Info("Horadric Cube chest found, moving to that room")
				chest, found := d.Objects.FindOne(object.HoradricCubeChest)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) getStaff() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.FarOasis),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.MaggotLairLevel1),
			a.builder.MoveToArea(area.MaggotLairLevel2),
			a.builder.MoveToArea(area.MaggotLairLevel3),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				a.logger.Info("Staff Of Kings chest found, moving to that room")
				chest, found := d.Objects.FindOne(object.StaffOfKingsChest)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) getAmulet() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.LostCity),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.ValleyOfSnakes),
			a.builder.MoveToArea(area.ClawViperTempleLevel1),
			a.builder.MoveToArea(area.ClawViperTempleLevel2),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				a.logger.Info("Altar found, moving closer")
				chest, found := d.Objects.FindOne(object.TaintedSunAltar)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killSummonerQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.ArcaneSanctuary),
			a.builder.OpenTP(),
			a.builder.Buff(),

			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				if summoner, found := d.NPCs.FindOne(npc.Summoner); found {
					return summoner.Positions[0], true
				}
				return data.Position{}, false
			}, step.StopAtDistance(80)),

			a.builder.OpenTP(),
			a.waitForParty(),
			a.char.KillSummoner(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killDurielQuest() action.Action {
	var realTomb area.ID

	for _, tomb := range talRashaTombs {
		_, _, objects, _ := a.Reader.GetCachedMapData(false).NPCsExitsAndObjects(data.Position{}, tomb)
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

	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action

		actions = append(actions,
			a.builder.WayPoint(area.CanyonOfTheMagi),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(realTomb),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				orifice, found := d.Objects.FindOne(object.HoradricOrifice)
				if found {
					return orifice.Position, true
				}
				return data.Position{}, false
			}),
			a.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.Buff(),
		)

		actions = append(actions,
			action.NewChain(func(d game.Data) []action.Action {
				_, found := d.Objects.FindOne(object.DurielsLairPortal)
				if found {
					return []action.Action{
						a.builder.InteractObject(object.DurielsLairPortal, func(d game.Data) bool {
							return d.PlayerUnit.Area == area.DurielsLair
						}),
					}
				}
				return nil
			}),
		)

		actions = append(actions,
			a.builder.MoveToArea(area.DurielsLair),
			a.char.KillDuriel(),
			a.builder.ReturnTown(),
			a.builder.WayPoint(area.KurastDocks),
		)

		return actions
	})
}

// Act 3
func (a Rushing) rushAct3() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {

		actions := []action.Action{
			a.builder.VendorRefill(true, true),
		}

		if a.CharacterCfg.Game.Rushing.GiveWPsA3 {
			actions = append(actions, a.GiveAct3WPs())
		}

		actions = append(actions,
			a.getKhalimsEye(),
			a.getKhalimsBrain(),
			a.getKhalimsHeart(),
		)

		if a.CharacterCfg.Game.Rushing.RetrieveBook {
			actions = append(actions, a.retrieveBookQuest())
		}

		actions = append(actions,
			a.killCouncilQuest(),
			a.killMephistoQuest(),
		)

		return actions
	})
}

func (a Rushing) getKhalimsEye() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.SpiderForest),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.SpiderCavern),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				chest, found := d.Objects.FindOne(object.KhalimChest3)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(25, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) getKhalimsBrain() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.FlayerJungle),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.FlayerDungeonLevel1),
			a.builder.MoveToArea(area.FlayerDungeonLevel2),
			a.builder.MoveToArea(area.FlayerDungeonLevel3),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				chest, found := d.Objects.FindOne(object.KhalimChest2)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) getKhalimsHeart() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.KurastBazaar),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.SewersLevel1Act3),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, l := range d.AdjacentLevels {
					if l.Area == area.SewersLevel2Act3 {
						return l.Position, true
					}
				}
				return data.Position{}, false
			}),
			a.builder.ClearAreaAroundPlayer(10, data.MonsterAnyFilter()),
			a.builder.InteractObject(object.Act3SewerStairsToLevel3, func(d game.Data) bool {
				o, _ := d.Objects.FindOne(object.Act3SewerStairsToLevel3)

				return !o.Selectable
			}),
			a.builder.Wait(time.Second * 3),
			a.builder.InteractObject(object.Act3SewerStairs, func(d game.Data) bool {
				return d.PlayerUnit.Area == area.SewersLevel2Act3
			}),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				chest, found := d.Objects.FindOne(object.KhalimChest1)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) retrieveBookQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.KurastBazaar),
			a.builder.OpenTP(),
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
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killCouncilQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.Travincal),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.Buff(),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, al := range d.AdjacentLevels {
					if al.Area == area.DuranceOfHateLevel1 {
						return data.Position{
							X: al.Position.X - 1,
							Y: al.Position.Y + 3,
						}, true
					}
				}
				return data.Position{}, false
			}),
			a.char.KillCouncil(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killMephistoQuest() action.Action {
	var mephistoSafePosition = data.Position{
		X: 17570,
		Y: 8007,
	}

	var mephistoPosition = data.Position{
		X: 17568,
		Y: 8069,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.DuranceOfHateLevel2),
			a.builder.OpenTP(),
			a.builder.MoveToArea(area.DuranceOfHateLevel3),
			a.builder.MoveToCoords(mephistoSafePosition),
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.Buff(),
			a.builder.MoveToCoords(mephistoPosition),
			a.char.KillMephisto(),
			a.builder.InteractObject(object.HellGate, func(d game.Data) bool {
				return d.PlayerUnit.Area == area.ThePandemoniumFortress
			}),
		}
	})
}

// Act 4
func (a Rushing) rushAct4() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {

		actions := []action.Action{
			a.builder.VendorRefill(true, false),
		}

		if a.CharacterCfg.Game.Rushing.GiveWPsA4 {
			actions = append(actions, a.GiveAct4WPs())
		}

		if a.CharacterCfg.Game.Rushing.KillIzual {
			actions = append(actions, a.killIzualQuest())
		}

		actions = append(actions,
			a.killDiabloQuest(),
		)

		return actions
	})
}

func (a Rushing) killIzualQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.MoveToArea(area.OuterSteppes),
			a.builder.Buff(),
			a.builder.MoveToArea(area.PlainsOfDespair),

			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				if izual, found := d.NPCs.FindOne(npc.Izual); found {
					return izual.Positions[0], true
				}
				return data.Position{}, false
			}, step.StopAtDistance(50)),

			a.builder.OpenTP(),
			a.waitForParty(),
			a.char.KillIzual(),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killDiabloQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action
		actions = append(actions, a.builder.WayPoint(area.RiverOfFlame))
		actions = append(actions, a.builder.Buff())
		actions = append(actions, a.builder.MoveToCoords(diabloSpawnPosition))

		seals := []object.Name{object.DiabloSeal4, object.DiabloSeal5, object.DiabloSeal3, object.DiabloSeal2, object.DiabloSeal1}

		for i, s := range seals {
			seal := s
			sealNumber := i

			actions = append(actions, a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				a.logger.Debug("Moving to next seal", slog.Int("seal", sealNumber+1))
				if obj, found := d.Objects.FindOne(seal); found {
					a.logger.Debug("Seal found, moving closer", slog.Int("seal", sealNumber+1))
					return obj.Position, true
				}
				a.logger.Debug("Seal NOT found", slog.Int("seal", sealNumber+1))
				return data.Position{}, false
			}, step.StopAtDistance(10)))

			actions = append(actions, action.NewStepChain(func(d game.Data) []step.Step {
				if obj, found := d.Objects.FindOne(seal); found {
					pos := a.getLessConcurredCornerAroundSeal(d, obj.Position)
					return []step.Step{step.MoveTo(pos)}
				}
				return []step.Step{}
			}))

			actions = append(actions,
				a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
				a.builder.ItemPickup(false, 40),
			)

			actions = append(actions,
				a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
				action.NewChain(func(d game.Data) []action.Action {
					if i == 0 {
						return []action.Action{
							a.builder.Buff(),
						}
					}
					return nil
				}),
			)

			actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
				obj, _ := d.Objects.FindOne(seal)
				if obj.Position.X == 7773 && obj.Position.Y == 5155 {
					return []action.Action{
						a.builder.MoveToCoords(data.Position{
							X: 7768,
							Y: 5160,
						}),
						action.NewStepChain(func(d game.Data) []step.Step {
							return []step.Step{step.InteractObjectByID(obj.ID, func(d game.Data) bool {
								if obj, found := d.Objects.FindOne(seal); found {
									if !obj.Selectable {
										a.logger.Debug("Seal activated, waiting for elite group to spawn", slog.Int("seal", sealNumber+1))
									}
									return !obj.Selectable
								}
								return false
							})}
						}),
					}
				}

				return []action.Action{a.builder.InteractObject(seal, func(d game.Data) bool {
					if obj, found := d.Objects.FindOne(seal); found {
						if !obj.Selectable {
							a.logger.Debug("Seal activated, waiting for elite group to spawn", slog.Int("seal", sealNumber+1))
						}
						return !obj.Selectable
					}
					return false
				})}
			}))

			if sealNumber != 0 {
				if sealNumber == 2 {
					actions = append(actions, a.builder.MoveToCoords(data.Position{
						X: 7773,
						Y: 5195,
					}))
				}

				startTime := time.Time{}
				actions = append(actions, action.NewStepChain(func(d game.Data) []step.Step {
					if startTime.IsZero() {
						startTime = time.Now()
					}
					for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
						if a.builder.IsMonsterSealElite(m) {
							a.logger.Debug("Seal defender found!")
							return nil
						}
					}

					if time.Since(startTime) < time.Second*5 {
						return []step.Step{step.Wait(time.Millisecond * 100)}
					}

					return nil
				}, action.RepeatUntilNoSteps()))

				actions = append(actions, a.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
					for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
						if a.builder.IsMonsterSealElite(m) {
							_, _, found := a.PathFinder.GetPath(d, m.Position)
							return m.UnitID, found
						}
					}
					return 0, false
				}, nil))
			}

			actions = append(actions, a.builder.ItemPickup(false, 40))
		}

		actions = append(actions,
			a.builder.MoveToCoords(data.Position{
				X: 7767,
				Y: 5252,
			}),
			a.builder.OpenTP(),
			a.waitForParty(),
			a.builder.Buff(),
			a.builder.MoveToCoords(diabloSpawnPosition),
			a.char.KillDiablo(),
			a.builder.ReturnTown(),
			a.builder.WayPoint(area.Harrogath),
		)

		return actions
	})
}

func (a Rushing) getLessConcurredCornerAroundSeal(d game.Data, sealPosition data.Position) data.Position {
	corners := [4]data.Position{
		{
			X: sealPosition.X + 7,
			Y: sealPosition.Y + 7,
		},
		{
			X: sealPosition.X - 7,
			Y: sealPosition.Y + 7,
		},
		{
			X: sealPosition.X - 7,
			Y: sealPosition.Y - 7,
		},
		{
			X: sealPosition.X + 7,
			Y: sealPosition.Y - 7,
		},
	}

	bestCorner := 0
	bestCornerDistance := 0
	for i, c := range corners {
		averageDistance := 0
		monstersFound := 0
		for _, m := range d.Monsters.Enemies() {
			distance := pather.DistanceFromPoint(c, m.Position)

			if distance < 5 {
				monstersFound++
				averageDistance += pather.DistanceFromPoint(c, m.Position)
			}
		}
		if averageDistance > bestCornerDistance {
			bestCorner = i
			bestCornerDistance = averageDistance
		}

		if monstersFound == 0 {
			a.logger.Debug("Moving to corner", slog.Int("corner", i), slog.Int("monsters", monstersFound))
			return corners[i]
		}
		a.logger.Debug("Corner", slog.Int("corner", i), slog.Int("monsters", monstersFound), slog.Int("distance", averageDistance))
	}

	a.logger.Debug("Moving to corner", slog.Int("corner", bestCorner), slog.Int("monsters", bestCornerDistance))

	return corners[bestCorner]
}

// Act 5
func (a Rushing) rushAct5() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {

		actions := []action.Action{
			a.builder.VendorRefill(true, false),
		}

		if a.CharacterCfg.Game.Rushing.GiveWPsA5 {
			actions = append(actions, a.GiveAct5WPs())
		}

		actions = append(actions,
			a.builder.VendorRefill(true, false),
			a.rescueAnyaQuest(),
			a.killAncientsQuest(),
			a.killBaalQuest(),
		)

		return actions
	})
}

func (a Rushing) rescueAnyaQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.CrystallinePassage),
			a.builder.Buff(),
			a.builder.MoveToArea(area.FrozenRiver),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				anya, found := d.NPCs.FindOne(793)
				return anya.Positions[0], found
			}),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				anya, found := d.Objects.FindOne(object.FrozenAnya)
				return anya.Position, found
			}),
			a.builder.OpenTP(),
			// a.waitForParty(d),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killAncientsQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.TheAncientsWay),
			a.builder.Buff(),
			a.builder.MoveToArea(area.ArreatSummit),
			a.builder.OpenTP(),
			// a.waitForParty(d),
			a.builder.Buff(),
			a.builder.InteractObject(object.AncientsAltar, func(d game.Data) bool {
				if len(d.Monsters.Enemies()) > 0 {
					return true
				}
				a.HID.Click(game.LeftButton, 300, 300)
				helper.Sleep(1000)
				return false
			}),
			a.builder.ClearAreaAroundPlayer(50, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
		}
	})
}

func (a Rushing) killBaalQuest() action.Action {
	var baalThronePosition = data.Position{
		X: 15095,
		Y: 5042,
	}

	var safebaalThronePosition = data.Position{
		X: 15116,
		Y: 5052,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action

		actions = append(actions,
			a.builder.WayPoint(area.TheWorldStoneKeepLevel2),
			a.builder.OpenTP(),
			a.builder.Buff(),
			a.builder.MoveToArea(area.TheWorldStoneKeepLevel3),
			a.builder.MoveToArea(area.ThroneOfDestruction),
			a.builder.MoveToCoords(baalThronePosition),
			a.builder.ClearAreaAroundPlayer(50, data.MonsterAnyFilter()),
			a.builder.MoveToCoords(safebaalThronePosition),
			a.builder.OpenTP(),
			a.builder.MoveToCoords(baalThronePosition),
		)

		lastWave := false
		actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
			if !lastWave {
				if _, found := d.Monsters.FindOne(npc.BaalsMinion, data.MonsterTypeMinion); found {
					lastWave = true
				}

				enemies := false
				for _, e := range d.Monsters.Enemies() {
					dist := pather.DistanceFromPoint(baalThronePosition, e.Position)
					if dist < 50 {
						enemies = true
					}
				}

				if !enemies {
					return []action.Action{
						a.builder.ItemPickup(false, 50),
						a.builder.MoveToCoords(baalThronePosition),
					}
				}

				return []action.Action{
					a.builder.ClearAreaAroundPlayer(50, data.MonsterAnyFilter()),
					a.builder.MoveToCoords(safebaalThronePosition),
					a.builder.Buff(),
					a.builder.Wait(time.Second * 4),
				}
			}
			return nil
		}, action.RepeatUntilNoSteps()))

		actions = append(actions, a.builder.ItemPickup(false, 30))

		actions = append(actions,
			a.builder.ReturnTown(),
			a.builder.VendorRefill(true, true),
			a.builder.UsePortalInTown(),
			a.builder.Buff(),
			a.builder.InteractObject(object.BaalsPortal, func(d game.Data) bool {
				return d.PlayerUnit.Area == area.TheWorldstoneChamber
			}),
			a.builder.OpenTP(),
			// a.waitForParty(d),
			a.char.KillBaal(),
			a.builder.ItemPickup(true, 50),
			a.builder.ReturnTown(),
			a.builder.Wait(time.Second*600),
		)

		return actions
	})
}

// Other functions
func (a Rushing) waitForParty() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action
		for {
			data := a.Container.Reader.GetData(false)

			var shouldContinue bool

			for _, c := range data.Roster {
				if c.Area.Area() == d.PlayerUnit.Area.Area() {
					shouldContinue = true
					break
				}
			}
			if shouldContinue {
				break
			} else {
				helper.Sleep(1000) // sleep 1
			}
		}

		return actions
	})
}
