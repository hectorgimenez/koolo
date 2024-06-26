package run

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Rushing) rushAct4() action.Action {
	running := false
	return action.NewChain(func(d game.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.ThePandemoniumFortress {
			return nil
		}

		running = true

		if a.CharacterCfg.Game.Rushing.GiveWPs {
			return []action.Action{
				a.builder.VendorRefill(true, false),
				a.GiveAct4WPs(),
				a.killIzualQuest(),
				a.killDiabloQuest(),
			}
		}

		return []action.Action{
			a.builder.VendorRefill(true, false),
			a.killIzualQuest(),
			a.killDiabloQuest(),
		}
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
			// a.waitForParty(d),
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
			// a.waitForParty(d),
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

func (a Rushing) generateClearActions(positions []data.Position, filter data.MonsterFilter) []action.Action {
	var actions []action.Action
	var maxPosDiff = 20

	for _, pos := range positions {
		actions = append(actions,
			action.NewChain(func(d game.Data) []action.Action {
				multiplier := 1

				if pather.IsWalkable(pos, d.AreaOrigin, d.CollisionGrid) {
					return []action.Action{a.builder.MoveToCoordsWithMinDistance(pos, 30)}
				}

				for _ = range 2 {
					for i := 1; i < maxPosDiff; i++ {

						newPos := data.Position{X: pos.X + (i * multiplier), Y: pos.Y + (i * multiplier)}

						if pather.IsWalkable(newPos, d.AreaOrigin, d.CollisionGrid) {
							return []action.Action{a.builder.MoveToCoordsWithMinDistance(newPos, 30)}
						}

					}

					multiplier *= -1
				}

				return []action.Action{a.builder.MoveToCoordsWithMinDistance(pos, 30)}
			}),
			a.builder.ClearAreaAroundPlayer(35, data.MonsterAnyFilter()),
			a.builder.ClearAreaAroundPlayer(35, func(m data.Monsters) []data.Monster {
				var monsters []data.Monster

				monsters = filter(m)
				monsters = skipStormCasterFilter(monsters)

				return monsters
			}),
			a.builder.ItemPickup(false, 35),
		)
	}
	return actions
}
