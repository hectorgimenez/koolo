package run

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

type A3Temples struct {
	baseRun
}

func (a A3Temples) Name() string {
	return string(config.A3TemplesRun)
}

func (a A3Temples) BuildActions() []action.Action {
	var actions []action.Action

	// Kurast Bazaar temples
	if a.CharacterCfg.Game.A3Temples.DoRuinedTemple || a.CharacterCfg.Game.A3Temples.DoDisusedFane {
		actions = append(actions, a.KurastBazaarTemplesActions())
	}

	// Upper Kurast temples
	if a.CharacterCfg.Game.A3Temples.DoForgottenTemple || a.CharacterCfg.Game.A3Temples.DoForgottenReliquary {
		actions = append(actions, a.UpperKurastTemplesActions())
	}

	// Kurast Causeway temples
	if a.CharacterCfg.Game.A3Temples.DoDisusedReliquary || a.CharacterCfg.Game.A3Temples.DoRuinedFane {
		actions = append(actions, a.KurastCausewayActions())
	}

	// Final return to town if any actions were performed
	if len(actions) > 0 {
		actions = append(actions, a.builder.ReturnTown())
	}

	return actions
}

func (a A3Temples) moveToKurastCauseway() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		wp, found := d.Objects.FindOne(object.Act3TownWaypoint)
		if !found {
			a.logger.Error("Act 3 Town Waypoint not found")
			return nil
		}

		actions := []action.Action{
			a.createMoveToAction(wp.Position, 0),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				return data.Position{X: wp.Position.X + 95, Y: wp.Position.Y + 31}, true
			}, step.StopAtDistance(0), step.WithTimeout(time.Second*30)),
			a.moveVerticallyBy(103),
			a.moveVerticallyBy(15),
			//additional vertical movement to ensure we reach Kurast Causeway
			a.moveVerticallyBy(10),
		}

		return actions
	})
}

func (a A3Temples) moveVerticallyBy(deltaY int) action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		currentPos := d.PlayerUnit.Position
		newY := currentPos.Y + deltaY

		actions := []action.Action{
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				return data.Position{X: currentPos.X, Y: newY}, true
			}, step.StopAtDistance(0), step.WithTimeout(time.Second*30)),
		}

		return actions
	})
}

func (a A3Temples) createMoveToAction(destination data.Position, stopDistance int) action.Action {
	return a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
		return destination, true
	}, step.StopAtDistance(stopDistance), step.WithTimeout(time.Second*30))
}

func (a A3Temples) moveWithRetry(targetArea area.ID) action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{
					step.SyncStep(func(d game.Data) error {
						a.logger.Info(fmt.Sprintf("Attempting to move to %s", targetArea.Area().Name))
						return nil
					}),
				}
			}),
			a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				for _, l := range d.AdjacentLevels {
					if l.Area == targetArea {
						return l.Position, true
					}
				}
				return data.Position{}, false
			}, step.WithTimeout(time.Second*30)),
			a.builder.MoveToArea(targetArea),
		}
	}, action.Resettable(), action.CanBeSkipped())
}

func (a A3Temples) KurastBazaarTemplesActions() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action

		actions = append(actions, a.builder.WayPoint(area.KurastBazaar))

		if a.CharacterCfg.Game.A3Temples.DoRuinedTemple {
			actions = append(actions, a.templeActionWithErrorHandling(area.RuinedTemple)...)
		}
		if a.CharacterCfg.Game.A3Temples.DoDisusedFane {
			actions = append(actions, a.templeActionWithErrorHandling(area.DisusedFane)...)
		}

		return actions
	})
}

func (a A3Temples) UpperKurastTemplesActions() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action

		actions = append(actions, a.builder.WayPoint(area.UpperKurast))

		if a.CharacterCfg.Game.A3Temples.DoForgottenTemple {
			actions = append(actions, a.templeActionWithErrorHandling(area.ForgottenTemple)...)
		}
		if a.CharacterCfg.Game.A3Temples.DoForgottenReliquary {
			actions = append(actions, a.templeActionWithErrorHandling(area.ForgottenReliquary)...)
		}

		return actions
	})
}

func (a A3Temples) KurastCausewayActions() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action

		actions = append(actions, a.builder.WayPoint(area.Travincal))
		actions = append(actions, a.moveToKurastCauseway())

		if a.CharacterCfg.Game.A3Temples.DoDisusedReliquary {
			actions = append(actions, a.templeActionWithErrorHandling(area.DisusedReliquary)...)
		}
		if a.CharacterCfg.Game.A3Temples.DoRuinedFane {
			actions = append(actions, a.templeActionWithErrorHandling(area.RuinedFane)...)
		}

		return actions
	})
}

func (a A3Temples) templeActionWithErrorHandling(templeArea area.ID) []action.Action {
	return []action.Action{
		a.moveWithRetry(templeArea),
		action.NewChain(func(d game.Data) []action.Action {
			if d.PlayerUnit.Area != templeArea {
				a.logger.Warn(fmt.Sprintf("Failed to reach %s, returning to town", templeArea.Area().Name))
				return []action.Action{
					a.builder.ReturnTown(),
				}
			}
			actions := a.templeActions(templeArea)
			actions = append(actions, a.moveAfterTemple(templeArea))
			return actions
		}),
	}
}

func (a A3Temples) moveAfterTemple(templeArea area.ID) action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		switch templeArea {
		case area.RuinedTemple, area.DisusedFane:
			return []action.Action{a.builder.MoveToArea(area.KurastBazaar)}
		case area.ForgottenTemple, area.ForgottenReliquary:
			return []action.Action{a.builder.MoveToArea(area.UpperKurast)}
		case area.DisusedReliquary:
			return []action.Action{a.builder.MoveToArea(area.KurastCauseway)}
		case area.RuinedFane:
			return []action.Action{a.builder.ReturnTown()}
		default:
			a.logger.Warn(fmt.Sprintf("Unknown temple area %s, returning to town", templeArea.Area().Name))
			return []action.Action{a.builder.ReturnTown()}
		}
	})
}

func (a A3Temples) templeActions(templeArea area.ID) []action.Action {
	actions := []action.Action{
		a.clearArea(),
		a.builder.ItemPickup(false, 10),
	}

	if templeArea == area.RuinedTemple && a.CharacterCfg.Game.A3Temples.OnlyKillBattleMaidSarina {
		actions = []action.Action{
			a.MoveToLamEsensTome(),
			a.KillBattleMaidSarina(),
			a.builder.ItemPickup(false, 10),
		}
	}

	return actions
}

func (a A3Temples) clearArea() action.Action {
	filter := data.MonsterAnyFilter()
	if a.CharacterCfg.Game.A3Temples.FocusOnElitePacks {
		filter = data.MonsterEliteFilter()
	}

	targetObjects := []object.Name{
		object.WeaponRackRight,
		object.WeaponRackLeft,
		object.ArmorStandRight,
		object.ArmorStandLeft,
	}

	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			a.efficientClearAndInteract(filter, targetObjects),
		}
	})
}

func (a A3Temples) efficientClearAndInteract(filter func(data.Monsters) []data.Monster, targetObjects []object.Name) action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action

		// Get walkable positions in the current area
		walkablePositions := a.PathFinder.GetWalkableRooms(d)

		for _, pos := range walkablePositions {
			actions = append(actions, a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
				return pos, true
			}, step.StopAtDistance(5)))

			// Clear monsters
			actions = append(actions, a.builder.ClearAreaAroundPlayer(20, filter))

			// Open chests and interact with target objects
			actions = append(actions, a.interactWithNearbyObjects(targetObjects))

			// Pickup items
			actions = append(actions, a.builder.ItemPickup(false, 20))
		}

		return actions
	})
}

func (a A3Temples) interactWithNearbyObjects(targetObjects []object.Name) action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action
		interactionAttempts := make(map[data.UnitID]int)

		for _, obj := range d.Objects {
			if (a.CharacterCfg.Game.A3Temples.OpenChests && obj.IsChest()) || (contains(targetObjects, obj.Name) && pather.DistanceFromMe(d, obj.Position) <= 20) {
				objID := obj.ID
				actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
					attempts := interactionAttempts[objID]
					if attempts >= 3 {
						a.logger.Warn(fmt.Sprintf("Skipping object %d after 3 failed interaction attempts", objID))
						return nil
					}

					return []action.Action{
						// Move to the object first
						a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
							o, found := d.Objects.FindByID(objID)
							if !found {
								return data.Position{}, false
							}
							return o.Position, true
						}, step.StopAtDistance(3)), // Adjust distance as needed

						// Then interact with the object
						a.builder.InteractObjectByID(objID, func(d game.Data) bool {
							o, found := d.Objects.FindByID(objID)
							success := !found || !o.Selectable
							if !success {
								interactionAttempts[objID]++
							}
							return success
						}),
					}
				}))
			}
		}

		return actions
	})
}

func contains(slice []object.Name, item object.Name) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func (a A3Temples) KillBattleMaidSarina() action.Action {
	return a.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(npc.FleshHunter, data.MonsterTypeSuperUnique); found {
			return m.UnitID, true
		}
		return 0, false
	}, nil)
}

func (a A3Temples) MoveToLamEsensTome() action.Action {
	return action.NewStepChain(func(d game.Data) []step.Step {
		for _, o := range d.Objects {
			if o.Name == object.LamEsensTome {
				return []step.Step{step.MoveTo(o.Position, step.StopAtDistance(10))}
			}
		}
		return nil
	})
}
