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

func (a Leveling) act1() action.Action {
	running := false
	return action.NewFactory(func(d data.Data) action.Action {
		if running || d.PlayerUnit.Area != area.RogueEncampment {
			return nil
		}

		running = true
		if d.PlayerUnit.Stats[stat.Level] <= 5 {
			return a.denOfEvil()
		}

		if d.PlayerUnit.Stats[stat.Level] > 5 && d.PlayerUnit.Stats[stat.Level] < 15 {
			return a.countess()
		}

		if !a.isCainInTown(d) {
			return a.deckardCain()
		}

		return a.andariel()
	})
}

func (a Leveling) denOfEvil() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		a.logger.Info("Starting Den of Evil run")
		return []action.Action{
			a.builder.MoveToArea(area.BloodMoor),
			a.char.Buff(),
			a.builder.MoveToArea(area.DenOfEvil),
			a.char.Buff(),
			a.builder.ClearArea(false, data.MonsterAnyFilter()),
		}
	})

}

func (a Leveling) bloodRaven() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		a.logger.Info("Starting Blood Raven quest")
		return []action.Action{
			a.builder.WayPoint(area.ColdPlains),
			a.builder.MoveToArea(area.BurialGrounds),
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
		}
	})
}

func (a Leveling) countess() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		// Moving to starting point (Black Marsh)
		a.logger.Info("Starting Countess run")

		return Countess{baseRun: a.baseRun}.BuildActions()
	})
}

func (a Leveling) deckardCain() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		a.logger.Info("Rescuing Cain")
		actions := []action.Action{
			a.builder.WayPoint(area.DarkWood),
			a.char.Buff(),
			a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
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

		// Heal and refill pots
		actions = append(actions, a.builder.PreRun(false)...)

		// Reuse Tristram Run actions
		actions = append(actions, Tristram{baseRun: a.baseRun}.BuildActions()...)

		return actions
	})
}

func (a Leveling) andariel() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		a.logger.Info("Starting Andariel run")
		actions := []action.Action{
			a.builder.WayPoint(area.CatacombsLevel2),
			a.char.Buff(),
			a.builder.MoveToArea(area.CatacombsLevel3),
			a.builder.MoveToArea(area.CatacombsLevel4),
		}
		actions = append(actions, a.builder.ReturnTown()) // Return town to pickup pots and heal, just in case...
		actions = append(actions, a.builder.PreRun(false)...)
		actions = append(actions,
			a.builder.UsePortalInTown(),
			a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
				return andarielStartingPosition, true
			}),
			a.char.KillAndariel(),
			a.builder.ReturnTown(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Warriv),
					step.KeySequence("esc", "home", "down", "enter"),
				}
			}))

		return actions
	})
}

func (a Leveling) isCainInTown(d data.Data) bool {
	_, found := d.Monsters.FindOne(npc.DeckardCain5, data.MonsterTypeNone)

	return found
}
