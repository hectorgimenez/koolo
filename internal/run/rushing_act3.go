package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (a Rushing) rushAct3() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.KurastDocks {
			return nil
		}

		running = true

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
			a.builder.WayPoint(area.UpperKurast),
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
