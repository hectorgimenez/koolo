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

const scrollOfInifuss = "ScrollOfInifuss"

func (a Leveling) act1() (actions []action.Action) {
	// ACT 1
	// Den of Evil - Farm until level 6
	actions = append(actions, a.denOfEvil())

	// Blood Raven
	//actions = append(actions, a.bloodRaven())

	// Deckard Cain - Will try after level 15, it's easier to farm Countess by this stupid useless bot
	actions = append(actions, a.deckardCain())

	// Countess - Farm until level 15
	actions = append(actions, a.countess())

	// Andariel and move to Act 2
	actions = append(actions, a.andariel())

	return
}

func (a Leveling) denOfEvil() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		if a.isCainInTown(d) || d.PlayerUnit.Stats[stat.Level] > 5 {
			// TODO: Check if we have the Den of Evil quest completed
			a.logger.Info("Skipping Den of Evil farming, character is already level 5 or cain is in town")
			return []action.Action{}
		}

		a.logger.Info("Starting Den of Evil run")
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

func (a Leveling) bloodRaven() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		//if a.isCainInTown(d) || d.PlayerUnit.Stats[stat.Level] < 5 || d.PlayerUnit.Stats[stat.Level] > 14 {
		//	// TODO: Check if we have the Blood Raven quest completed
		//	a.logger.Info("Skipping Blood Raven conditions not met")
		//	return []action.Action{}
		//}

		a.logger.Info("Starting Blood Raven quest")
		return []action.Action{
			a.builder.WayPoint(area.ColdPlains),
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
	})
}

func (a Leveling) countess() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		if a.isCainInTown(d) || d.PlayerUnit.Stats[stat.Level] < 6 || d.PlayerUnit.Stats[stat.Level] > 14 {
			a.logger.Info("Skipping Countess farm, already level 15 or Cain in town")
			return
		}

		a.logger.Info("Starting Countess run")
		// Moving to starting point (Black Marsh)
		actions = append(actions, a.builder.WayPoint(area.BlackMarsh))

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
	})
}

func (a Leveling) deckardCain() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		if a.isCainInTown(d) || d.PlayerUnit.Stats[stat.Level] < 15 {
			a.logger.Info("Skipping Deckard Cain quest, he is already in town or still not level 15")
			return []action.Action{}
		}

		a.logger.Info("Rescuing Cain")
		actions := []action.Action{
			a.builder.WayPoint(area.DarkWood),
			a.char.Buff(),
			a.builder.MoveAndKill(func(d data.Data) (data.Position, bool) {
				for _, o := range d.Objects {
					if o.Name == object.InifussTree {
						return o.Position, true
					}
				}

				return data.Position{}, false
			}),
			action.BuildStatic(func(d data.Data) []step.Step {
				for _, o := range d.Objects {
					if o.Name == object.InifussTree {
						return []step.Step{
							step.InteractObject(o.Name, func(d data.Data) bool {
								for _, o := range d.Objects {
									if o.Name == object.InifussTree {
										return !o.Selectable
									}
								}

								return true
							}),
						}
					}
				}

				return []step.Step{}
			}),

			action.BuildStatic(func(d data.Data) []step.Step {
				for _, i := range d.Items.Ground {
					if i.Name == scrollOfInifuss {
						return []step.Step{step.PickupItem(a.logger, i)}
					}
				}
				return nil
			}, action.IgnoreErrors()),
			a.builder.ReturnTown(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Akara),
					step.SyncStepWithCheck(func(d data.Data) error {
						hid.PressKey("esc")
						helper.Sleep(1000)
						return nil
					}, func(data.Data) step.Status {
						if d.OpenMenus.NPCInteract {
							return step.StatusInProgress
						}
						return step.StatusCompleted
					}),
				}
			}),
		}

		// Reuse Tristram Run actions
		actions = append(actions, Tristram{baseRun: a.baseRun}.BuildActions()...)

		return actions
	})
}

func (a Leveling) andariel() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		if !a.isCainInTown(d) || d.PlayerUnit.Stats[stat.Level] < 15 {
			a.logger.Info("Skipping Andariel, Cain is not in town or still not level 15")
			return []action.Action{}
		}

		a.logger.Info("Starting Andariel run")
		return []action.Action{
			a.builder.WayPoint(area.CatacombsLevel2),
			a.char.Buff(),
			a.builder.MoveToAreaAndKill(area.CatacombsLevel3),
			a.builder.MoveToAreaAndKill(area.CatacombsLevel4),
			a.builder.MoveAndKill(func(d data.Data) (data.Position, bool) {
				return andarielStartingPosition, true
			}),
			a.builder.ReturnTown(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Warriv),
					step.KeySequence("home", "down", "enter"),
				}
			}),
		}
	})
}

func (a Leveling) isCainInTown(d data.Data) bool {
	// TODO This fails if Cain is out of range, ideally we should move closer to his place
	_, found := d.Monsters.FindOne(npc.DeckardCain5, data.MonsterTypeNone)

	return found
}
