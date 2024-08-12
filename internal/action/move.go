package action

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b *Builder) MoveToArea(dst area.ID) *Chain {
	// Exception for Arcane Sanctuary, we need to find the portal first
	if dst == area.ArcaneSanctuary {
		return NewChain(func(d game.Data) []Action {
			b.Logger.Debug("Arcane Sanctuary detected, finding the Portal")
			portal, _ := d.Objects.FindOne(object.ArcaneSanctuaryPortal)
			return []Action{
				b.MoveToCoords(portal.Position),
				NewStepChain(func(d game.Data) []step.Step {
					return []step.Step{
						step.InteractObject(object.ArcaneSanctuaryPortal, func(d game.Data) bool {
							return d.PlayerUnit.Area == area.ArcaneSanctuary
						}),
					}
				}),
			}
		})
	}

	toFun := func(d game.Data) (data.Position, bool) {
		if d.PlayerUnit.Area == dst {
			b.Logger.Debug("Reached area", slog.String("area", dst.Area().Name))
			return data.Position{}, false
		}

		switch dst {
		case area.MonasteryGate:
			b.Logger.Debug("Monastery Gate detected, moving to static coords")
			return data.Position{X: 15139, Y: 5056}, true
		}

		for _, a := range d.AdjacentLevels {
			if a.Area == dst {
				// To correctly detect the two possible exits from Lut Gholein
				if dst == area.RockyWaste && d.PlayerUnit.Area == area.LutGholein {
					if _, _, found := b.PathFinder.GetPath(d, data.Position{X: 5004, Y: 5065}); found {
						return data.Position{X: 4989, Y: 5063}, true
					} else {
						return data.Position{X: 5096, Y: 4997}, true
					}
				}

				// This means it's a cave, we don't want to load the map, just find the entrance and interact
				if a.IsEntrance {
					return a.Position, true
				}

				lvl, _ := b.Reader.GetCachedMapData(false).GetLevelData(a.Area)
				_, _, objects, _ := b.Reader.GetCachedMapData(false).NPCsExitsAndObjects(lvl.Offset, a.Area)

				// Sort objects by the distance from me
				sort.Slice(objects, func(i, j int) bool {
					distanceI := pather.DistanceFromMe(d, objects[i].Position)
					distanceJ := pather.DistanceFromMe(d, objects[j].Position)

					return distanceI < distanceJ
				})

				// Let's try to find any random object to use as a destination point, once we enter the level we will exit this flow
				for _, obj := range objects {
					_, _, found := b.PathFinder.GetPath(d, obj.Position)
					if found {
						return obj.Position, true
					}
				}

				return a.Position, true
			}
		}

		b.Logger.Debug("Destination area not found", slog.String("area", dst.Area().Name))

		return data.Position{}, false
	}

	return NewChain(func(d game.Data) []Action {
		return []Action{
			b.MoveTo(toFun),
			NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{
					step.InteractEntrance(dst),
					step.SyncStep(func(d game.Data) error {
						event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(dst), event.InteractionTypeEntrance))
						return nil
					}),
				}
			}),
		}
	}, Resettable())
}

func (b *Builder) MoveToCoordsWithMinDistance(to data.Position, minDistance int, opts ...step.MoveToStepOption) *Chain {
	return b.MoveTo(func(d game.Data) (data.Position, bool) {
		distance := pather.DistanceFromMe(d, to)
		if distance <= minDistance {
			return d.PlayerUnit.Position, false
		}
		return to, true
	}, opts...)
}

func (b *Builder) MoveToCoords(to data.Position, opts ...step.MoveToStepOption) *Chain {
	return b.MoveTo(func(d game.Data) (data.Position, bool) {
		return to, true
	}, opts...)
}

func (b *Builder) MoveTo(toFunc func(d game.Data) (data.Position, bool), opts ...step.MoveToStepOption) *Chain {
	pickupBeforeMoving := false
	openedDoors := make(map[object.Name]data.Position)
	previousIterationPosition := data.Position{}

	return NewChain(func(d game.Data) []Action {
		to, found := toFunc(d)
		if !found {
			return nil
		}

		_, distance, _ := b.PathFinder.GetPath(d, to)
		mvtStep := step.MoveTo(to, opts...)
		if distance <= mvtStep.GetStopDistance() {
			return nil
		}

		// This prevents we stuck in an infinite loop when we can not get closer to the destination
		if pather.DistanceFromMe(d, previousIterationPosition) < 5 {
			return nil
		}

		if d.CanTeleport() {
			previousIterationPosition = d.PlayerUnit.Position

			return []Action{NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.MoveTo(to, opts...)}
			})}
		}

		// Check if there is a door blocking our path
		for _, o := range d.Objects {
			if o.IsDoor() && pather.DistanceFromMe(d, o.Position) < 10 && openedDoors[o.Name] != o.Position {
				if o.Selectable {
					return []Action{NewStepChain(func(d game.Data) []step.Step {
						b.Logger.Info("Door detected and teleport is not available, trying to open it...")
						openedDoors[o.Name] = o.Position
						return []step.Step{step.InteractObjectByID(o.ID, func(d game.Data) bool {
							obj, found := d.Objects.FindByID(o.ID)

							return found && !obj.Selectable
						})}
					}, CanBeSkipped())}
				}
			}
		}

		// Check if there is any object blocking our path
		for _, o := range d.Objects {
			if o.Name == object.Barrel && pather.DistanceFromMe(d, o.Position) < 3 {
				return []Action{NewStepChain(func(d game.Data) []step.Step {
					return []step.Step{step.InteractObjectByID(o.ID, func(d game.Data) bool {
						obj, found := d.Objects.FindByID(o.ID)
						//additional click on barrel to avoid getting stuck
						x, y := b.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, o.Position.X, o.Position.Y)
						b.HID.Click(game.LeftButton, x, y)
						return found && !obj.Selectable
					})}
				})}
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
				b.Logger.Info("Character stuck and monsters detected, trying to kill monsters around")
			} else {
				b.Logger.Info(fmt.Sprintf("At least %d monsters detected close to the character, targeting closest one: %d", len(targetedNormalEnemies)+len(targetedElites), closestMonster.Name))
			}

			path, _, mPathFound := b.PathFinder.GetPath(d, closestMonster.Position)
			if mPathFound {
				doorIsBlocking := false
				for _, o := range d.Objects {
					if o.IsDoor() && o.Selectable && path.Intersects(d, o.Position, 4) {
						b.Logger.Debug("Door is blocking the path to the monster, skipping attack sequence")
						doorIsBlocking = true
					}
				}

				if !doorIsBlocking {
					pickupBeforeMoving = true
					return []Action{b.ch.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
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
		return []Action{b.WaitForAllMembersWhenLeveling(), NewStepChain(func(d game.Data) []step.Step {
			newOpts := append(opts, step.WithTimeout(time.Millisecond*1000))
			previousIterationPosition = d.PlayerUnit.Position

			return []step.Step{step.MoveTo(
				to,
				newOpts...,
			)}
		})}
	}, RepeatUntilNoSteps())
}
