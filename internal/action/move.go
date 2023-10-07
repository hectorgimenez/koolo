package action

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/reader"
	"go.uber.org/zap"
)

func (b *Builder) MoveToArea(dst area.Area, opts ...step.MoveToStepOption) *Chain {
	// Exception for Arcane Sanctuary, we need to find the portal first
	if dst == area.ArcaneSanctuary {
		return NewChain(func(d data.Data) []Action {
			b.logger.Debug("Arcane Sanctuary detected, finding the Portal")
			portal, _ := d.Objects.FindOne(object.ArcaneSanctuaryPortal)
			return []Action{
				b.MoveToCoords(portal.Position),
				NewStepChain(func(d data.Data) []step.Step {
					return []step.Step{
						step.InteractObject(object.ArcaneSanctuaryPortal, func(d data.Data) bool {
							return d.PlayerUnit.Area == area.ArcaneSanctuary
						}),
					}
				}),
			}
		})
	}

	toFun := func(d data.Data) (data.Position, bool) {
		if d.PlayerUnit.Area == dst {
			b.logger.Debug("Already in area", zap.Any("area", dst))
			return data.Position{}, false
		}

		switch dst {
		case area.MonasteryGate:
			b.logger.Debug("Monastery Gate detected, moving to static coords")
			return data.Position{X: 15139, Y: 5056}, true
		}

		for _, a := range d.AdjacentLevels {
			if a.Area == dst {
				// To correctly detect the two possible exits from Lut Gholein
				if dst == area.RockyWaste && d.PlayerUnit.Area == area.LutGholein {
					if _, _, found := pather.GetPath(d, data.Position{X: 5004, Y: 5065}); found {
						return data.Position{X: 4989, Y: 5063}, true
					} else {
						return data.Position{X: 5096, Y: 4997}, true
					}
				}

				// This means it's a cave, we don't want to load the map, just find the entrance and interact
				if a.IsEntrance {
					return a.Position, true
				}

				lvl, _ := reader.CachedMapData.GetLevelData(a.Area)
				_, _, objects, _ := reader.CachedMapData.NPCsExitsAndObjects(lvl.Offset, a.Area)

				// Let's try to find any random object to use as a destination point, once we enter the level we will exit this flow
				for _, obj := range objects {
					_, _, found := pather.GetPath(d, obj.Position)
					if found {
						return obj.Position, true
					}
				}

				return a.Position, true
			}
		}

		b.logger.Debug("Destination area not found", zap.Any("area", dst))

		return data.Position{}, false
	}

	return NewChain(func(d data.Data) []Action {
		return []Action{
			b.MoveTo(toFun, opts...),
			NewStepChain(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractEntrance(dst),
				}
			}),
		}
	})
}

func (b *Builder) MoveToCoords(to data.Position, opts ...step.MoveToStepOption) *Chain {
	return b.MoveTo(func(d data.Data) (data.Position, bool) {
		return to, true
	}, opts...)
}

func (b *Builder) MoveTo(toFunc func(d data.Data) (data.Position, bool), opts ...step.MoveToStepOption) *Chain {
	pickupBeforeMoving := false
	openedDoors := make(map[object.Name]data.Position)
	previousIterationPosition := data.Position{}
	previousIterationTo := data.Position{}
	var currentStep step.Step

	return NewChain(func(d data.Data) []Action {
		to, found := toFunc(d)
		if !found {
			return nil
		}

		if previousIterationTo != to && currentStep != nil {
			currentStep = nil
		}

		// To stop the movement, not very accurate
		if pather.DistanceFromMe(d, to) < 7 {
			return nil
		}

		// Let's go pickup more pots if we have less than 2 (only during leveling)
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if isLevelingChar && !d.PlayerUnit.Area.IsTown() {
			_, healingPotsFound := d.Items.Belt.GetFirstPotion(data.HealingPotion)
			_, manaPotsFound := d.Items.Belt.GetFirstPotion(data.ManaPotion)
			if (!healingPotsFound || !manaPotsFound) && d.PlayerUnit.TotalGold() > 1000 {
				return []Action{NewChain(func(d data.Data) []Action {
					return b.InRunReturnTownRoutine()
				})}
			}
		}

		if helper.CanTeleport(d) {
			// If we can teleport, and we're not on leveling sequence, just return the normal MoveTo step and stop here
			if !isLevelingChar {
				return []Action{NewStepChain(func(d data.Data) []step.Step {
					return []step.Step{step.MoveTo(to, opts...)}
				})}
			}
			// But if we are leveling and have enough money (to buy mana pots), let's teleport. We add the timeout
			// to re-trigger this action, so we can get back to town to buy pots in case of empty belt
			if d.PlayerUnit.TotalGold() > 30000 {
				return []Action{NewStepChain(func(d data.Data) []step.Step {
					newOpts := append(opts, step.WithTimeout(5*time.Second))
					return []step.Step{step.MoveTo(to, newOpts...)}
				})}
			}
		}

		// Check if there is a door blocking our path
		for _, o := range d.Objects {
			if o.IsDoor() && pather.DistanceFromMe(d, o.Position) < 10 && openedDoors[o.Name] != o.Position {
				if o.Selectable {
					return []Action{NewStepChain(func(d data.Data) []step.Step {
						b.logger.Info("Door detected and teleport is not available, trying to open it...")
						openedDoors[o.Name] = o.Position
						return []step.Step{step.InteractObject(o.Name, func(d data.Data) bool {
							for _, obj := range d.Objects {
								if obj.Name == o.Name && obj.Position == o.Position && !obj.Selectable {
									return true
								}
							}
							return false
						})}
					}, CanBeSkipped())}
				}
			}
		}

		// Detect if there are monsters close to the player
		closestMonster := data.Monster{}
		closestMonsterDistance := 9999999
		targetedNormalEnemies := make([]data.Monster, 0)
		targetedElites := make([]data.Monster, 0)
		minDistance := 6
		minDistanceForElites := 20                                       // This will make the character to kill elites even if they are far away, ONLY during leveling
		stuck := pather.DistanceFromMe(d, previousIterationPosition) < 5 // Detect if character was not able to move from last iteration
		for _, m := range d.Monsters.Enemies() {
			// Skip if monster is already dead
			if m.Stats[stat.Life] <= 0 {
				continue
			}

			dist := pather.DistanceFromMe(d, m.Position)
			appended := false
			if m.IsElite() && dist <= minDistanceForElites {
				targetedElites = append(targetedElites, m)
				appended = true
			}

			if dist <= minDistance {
				targetedNormalEnemies = append(targetedNormalEnemies, m)
				appended = true
			}

			if appended {
				if dist < closestMonsterDistance {
					closestMonsterDistance = dist
					closestMonster = m
				}
			}
		}

		if len(targetedNormalEnemies) > 5 || len(targetedElites) > 0 || (stuck && (len(targetedNormalEnemies) > 0 || len(targetedElites) > 0)) || (pather.IsNarrowMap(d.PlayerUnit.Area) && (len(targetedNormalEnemies) > 0 || len(targetedElites) > 0)) {
			if stuck {
				b.logger.Info("Character stuck and monsters detected, trying to kill monsters around")
			} else {
				b.logger.Info(fmt.Sprintf("At least %d monsters detected close to the character, targeting closest one: %d", len(targetedNormalEnemies)+len(targetedElites), closestMonster.Name))
			}

			path, _, mPathFound := pather.GetPath(d, closestMonster.Position)
			if mPathFound {
				doorIsBlocking := false
				for _, o := range d.Objects {
					if o.IsDoor() && o.Selectable && path.Intersects(d, o.Position, 4) {
						b.logger.Debug("Door is blocking the path to the monster, skipping attack sequence")
						doorIsBlocking = true
					}
				}

				if !doorIsBlocking {
					pickupBeforeMoving = true
					return []Action{b.ch.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
						return closestMonster.UnitID, true
					}, nil)}
				}
			}
		}

		if pickupBeforeMoving {
			pickupBeforeMoving = false
			return []Action{b.ItemPickup(false, 30)}
		}

		// Continue moving
		return []Action{NewStepChain(func(d data.Data) []step.Step {
			newOpts := append(opts, step.ClosestWalkable(), step.WithTimeout(time.Millisecond*1000))
			previousIterationPosition = d.PlayerUnit.Position
			if currentStep == nil {
				currentStep = step.MoveTo(
					to,
					newOpts...,
				)
			} else {
				currentStep.Reset()
			}

			return []step.Step{currentStep}
		})}
	}, RepeatUntilNoSteps())
}
