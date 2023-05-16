package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type Leveling struct {
	baseRun
}

func (a Leveling) Name() string {
	return "Leveling"
}

func (a Leveling) BuildActions() (actions []action.Action) {
	// ACT 1
	// Den of Evil
	actions = append(actions, a.denOfEvil())

	// Blood Raven
	actions = append(actions, a.bloodRaven()...)

	//// Cairn
	//actions = append(actions, a.cairn()...)

	//// Countess
	//actions = append(actions, a.countess()...)

	return
}

func (a Leveling) denOfEvil() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		if d.PlayerUnit.Stats[stat.Level] >= 5 {
			// TODO: Check if we have the Den of Evil quest completed
			a.logger.Info("Skipping Den of Evil farming, character is already level 5")
			return []action.Action{}
		}

		return []action.Action{
			a.builder.MoveToAreaAndKill(area.BloodMoor),
			a.char.Buff(),
			a.builder.MoveToAreaAndKill(area.DenOfEvil),
			a.char.Buff(),
			a.builder.ClearArea(false, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
		}
	})

}

func (a Leveling) bloodRaven() []action.Action {
	return []action.Action{
		a.builder.MoveToAreaAndKill(area.BloodMoor),
		a.char.Buff(),
		a.builder.MoveToAreaAndKill(area.ColdPlains),
		a.DiscoverWaypoint(),
		a.builder.MoveToAreaAndKill(area.BurialGrounds),
		a.char.Buff(),
		action.BuildStatic(func(d data.Data) []step.Step {
			for _, l := range d.AdjacentLevels {
				if l.Area == area.Mausoleum {
					return []step.Step{step.MoveTo(l.Position, step.StopAtDistance(50))}
				}
			}

			return []step.Step{}
		}),
		a.char.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies() {
				if pather.DistanceFromMe(d, m.Position) < 3 {
					return m.UnitID, true
				}

				if m.Name == npc.BloodRaven {
					return m.UnitID, true
				}
			}

			return 0, false
		}, nil, step.Distance(5, 15)),
		a.builder.ReturnTown(),
	}
}

func (a Leveling) countess() (actions []action.Action) {
	// Moving to starting point (Black Marsh)
	actions = append(actions, a.builder.MoveToAreaAndKill(area.BlackMarsh))
	actions = append(actions, a.DiscoverWaypoint())
	//actions = append(actions, a.builder.WayPoint(area.BlackMarsh))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to boss level
	actions = append(actions,
		a.builder.MoveToAreaAndKill(area.ForgottenTower),
		a.builder.MoveToAreaAndKill(area.TowerCellarLevel1),
		a.builder.MoveToAreaAndKill(area.TowerCellarLevel2),
		a.builder.MoveToAreaAndKill(area.TowerCellarLevel3),
		a.builder.MoveToAreaAndKill(area.TowerCellarLevel4),
		a.builder.MoveToAreaAndKill(area.TowerCellarLevel5),
	)

	// Try to move around Countess area
	actions = append(actions, a.builder.MoveAndKill(func(d data.Data) (data.Position, bool) {
		if countess, found := d.Monsters.FindOne(npc.DarkStalker, data.MonsterTypeSuperUnique); found {
			return countess.Position, true
		}

		for _, o := range d.Objects {
			if o.Name == object.GoodChest {
				return o.Position, true
			}
		}

		return data.Position{}, false
	}))

	return actions
}

func (a Leveling) cairn() []action.Action {
	return []action.Action{
		//a.builder.WayPoint(area.ColdPlains),
		a.char.Buff(),
		//a.builder.MoveToAreaAndKill(area.StonyField),
		//a.DiscoverWaypoint(),
		//a.builder.MoveToAreaAndKill(area.UndergroundPassageLevel1),
		//a.builder.MoveToAreaAndKill(area.DarkWood),
		//a.DiscoverWaypoint(),
		action.BuildStatic(func(d data.Data) []step.Step {
			for _, o := range d.Objects {
				if o.Name == object.InifussTree {
					return []step.Step{
						step.MoveTo(o.Position, step.StopAtDistance(20)),
						step.InteractObject(o.Name, func(d data.Data) bool {
							for _, o := range d.Objects {
								if o.Name == object.InifussTree {
									return o.Selectable
								}
							}

							return false
						}),
					}
				}
			}

			return []step.Step{}
		}),
		a.builder.ReturnTown(),
	}
}

func (a Leveling) DiscoverWaypoint() action.Action {
	interacted := false

	return action.NewFactory(func(d data.Data) action.Action {
		if interacted {
			return nil
		}

		for _, o := range d.Objects {
			if o.IsWaypoint() {
				if pather.DistanceFromMe(d, o.Position) < 15 {
					return action.BuildStatic(func(d data.Data) []step.Step {
						return []step.Step{
							step.MoveTo(o.Position),
							step.InteractObject(o.Name, func(d data.Data) bool {
								return d.OpenMenus.Waypoint
							}),
							step.SyncStep(func(d data.Data) error {
								helper.Sleep(1000)
								hid.PressKey("esc")
								interacted = true
								return nil
							}),
						}
					})
				}

				return a.builder.MoveAndKill(func(d data.Data) (data.Position, bool) {
					for _, o := range d.Objects {
						if o.IsWaypoint() {
							return o.Position, true
						}
					}

					return data.Position{}, false
				})
			}
		}

		return nil
	})
}
