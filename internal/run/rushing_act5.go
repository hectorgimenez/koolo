package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Rushing) rushAct5() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.Harrogath {
			return nil
		}

		running = true

		if a.CharacterCfg.Game.Rushing.GiveWPs {
			return []action.Action{
				a.builder.VendorRefill(true, false),
				a.GiveAct5WPs(),
				a.rescueAnyaQuest(),
				a.killAncientsQuest(),
				a.killBaalQuest(),			
			}
		}

		return []action.Action{
			a.builder.VendorRefill(true, false),
			a.rescueAnyaQuest(),
			a.killAncientsQuest(),
			a.killBaalQuest(),			
		}
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
			a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
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
            a.builder.Buff(),
            a.builder.InteractObject(object.BaalsPortal, func(d game.Data) bool {
                return d.PlayerUnit.Area == area.TheWorldstoneChamber
            }),
			a.builder.OpenTP(),
            // a.waitForParty(d),
            a.char.KillBaal(),
            a.builder.ItemPickup(true, 50),
            a.builder.ReturnTown(),
            a.builder.Wait(time.Second * 600),
        )

        return actions
    })
}

