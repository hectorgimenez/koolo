package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (a Rushing) rushAct1() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.RogueEncampment {
			return nil
		}

		running = true

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
