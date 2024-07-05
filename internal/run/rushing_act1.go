package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

func (r Rushing) getRushedAct1() action.Action {
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		rusherStatus := r.getRusherStatus(r.CharacterCfg.Companion.LeaderName)
		if rusherStatus == None {
			actions = append(actions, r.builder.Wait(time.Second))
		}

		if rusherStatus == GivingWPs {
			// Check town for portals
			// Take a portal if open
			// Take waypoint
			// Go back to town
		}

		if rusherStatus == ClearingDen {
			// Take portal to den if open, otherwise we wait
			// When rusher leaves den, we portal to town
		}

		if rusherStatus == FreeingCain {
			// Wait for portal to dark wood
			// Take portal
			// Take scroll
			// Go back to town
			// Talk to Akara
			// Wait for portal to stony field
			// Interact with stones
			// Go to town
			// Wait for portal to tristram
			// Take tristram portal
			// Interact with gimp cage
			// Go to town
		}

		if rusherStatus == RetrievingHammer {
			// Wait for portal to barracks
			// Take portal to barracks
			// Get hammer
			// Portal back to town
		}

		if rusherStatus == KillingAndy {
			// Wait for portal to catacombs level 4
			// Wait for andy death
			// Portal back to town
			// Talk with the blue teletubby
			// Go to act 2
		}
		return actions
	})
}

func (r Rushing) rushAct1() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.RogueEncampment {
			return nil
		}

		running = true

		actions := []action.Action{
			r.builder.VendorRefill(true, true),
		}

		if r.CharacterCfg.Game.Rushing.GiveWPsA1 {
			actions = append(actions, r.GiveAct1WPs())
		}

		if r.CharacterCfg.Game.Rushing.ClearDen {
			actions = append(actions, r.clearDenQuest())
		}

		if r.CharacterCfg.Game.Rushing.RescueCain {
			actions = append(actions, r.rescueCainQuest())
		}

		if r.CharacterCfg.Game.Rushing.RetrieveHammer {
			actions = append(actions, r.retrieveHammerQuest())
		}

		actions = append(actions,
			r.killAandarielQuest(),
		)

		return actions
	})
}

func (r Rushing) GiveAct1WPs() action.Action {
	areas := []area.ID{
		area.StonyField,
		area.DarkWood,
		area.BlackMarsh,
		area.InnerCloister,
		area.OuterCloister,
		area.CatacombsLevel2,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		r.setRusherStatus(r.CharacterCfg.CharacterName, GivingWPs)
		actions := []action.Action{}

		for _, areaID := range areas {
			actions = append(actions,
				r.builder.WayPoint(areaID),
				r.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
				r.builder.OpenTP(),
				r.builder.WaitForPartyToEnterPortal(r.CharacterCfg.CharacterName),
			)
		}

		return actions
	})
}

func (r Rushing) clearDenQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		r.setRusherStatus(r.CharacterCfg.CharacterName, ClearingDen)
		return []action.Action{
			r.builder.MoveToArea(area.BloodMoor),
			r.builder.Buff(),
			r.builder.MoveToArea(area.DenOfEvil),
			r.builder.OpenTP(),
			r.builder.WaitForParty(r.CharacterCfg.CharacterName),
			r.builder.ClearArea(false, data.MonsterAnyFilter()),
			r.builder.ReturnTown(),
		}
	})
}

func (r Rushing) rescueCainQuest() action.Action {
	var gimpCage = data.Position{
		X: 25140,
		Y: 5145,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		r.setRusherStatus(r.CharacterCfg.CharacterName, FreeingCain)
		return []action.Action{
			// Go to Tree
			r.builder.WayPoint(area.DarkWood),
			r.builder.OpenTP(),
			r.builder.Buff(),
			r.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.InifussTree {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			r.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			r.builder.WaitForPartyToEnterPortal(r.CharacterCfg.CharacterName),
			r.builder.ReturnTown(),

			// Go to Stones
			r.builder.WayPoint(area.StonyField),
			r.builder.OpenTP(),
			r.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.CairnStoneAlpha {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			r.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			r.builder.OpenTP(),
			r.builder.WaitForParty(r.CharacterCfg.CharacterName),

			// Wait for Tristram portal and enter
			action.NewChain(func(d game.Data) []action.Action {
				_, found := d.Objects.FindOne(object.PermanentTownPortal)
				if found {
					return []action.Action{
						r.builder.InteractObject(object.PermanentTownPortal, func(d game.Data) bool {
							return d.PlayerUnit.Area == area.Tristram
						}),
					}
				}
				return nil
			}),
			r.builder.MoveToArea(area.Tristram),
			r.builder.MoveToCoords(gimpCage),
			r.builder.ClearAreaAroundPlayer(30, data.MonsterAnyFilter()),
			r.builder.WaitForParty(r.CharacterCfg.CharacterName),
			r.builder.ReturnTown(),
		}
	})
}

func (r Rushing) retrieveHammerQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		r.setRusherStatus(r.CharacterCfg.CharacterName, RetrievingHammer)
		return []action.Action{
			r.builder.WayPoint(area.OuterCloister),
			r.builder.OpenTP(),
			r.builder.Buff(),
			r.builder.MoveToArea(area.Barracks),
			r.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.Malus {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			r.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			r.builder.OpenTP(),
			r.builder.WaitForParty(r.CharacterCfg.CharacterName),
			r.builder.ReturnTown(),
		}
	})
}

func (r Rushing) killAandarielQuest() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		r.setRusherStatus(r.CharacterCfg.CharacterName, KillingAndy)
		return []action.Action{
			r.builder.WayPoint(area.CatacombsLevel2),
			r.builder.OpenTP(),
			r.builder.Buff(),
			r.builder.MoveToArea(area.CatacombsLevel3),
			r.builder.MoveToArea(area.CatacombsLevel4),
			r.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
			r.builder.OpenTP(),
			r.builder.WaitForParty(r.CharacterCfg.CharacterName),
			r.builder.MoveToCoords(andarielStartingPosition),
			r.char.KillAndariel(),
			r.builder.ReturnTown(),
			r.builder.WayPoint(area.LutGholein),
		}
	})
}
