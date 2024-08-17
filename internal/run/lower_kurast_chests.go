package run

import (
	"fmt"
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

var bonfireName = object.SmallFire
var chestsIds = []object.Name{object.JungleMediumChestLeft, object.JungleChest}
var minChestDistanceFromBonfire = 25
var maxChestDistanceFromBonfire = 45

type LowerKurastChest struct {
	baseRun
}

func (a LowerKurastChest) Name() string {
	return string(config.LowerKurastChestRun)
}

func (a LowerKurastChest) BuildActions() []action.Action {
	actions := []action.Action{
		a.builder.WayPoint(area.LowerKurast),
		action.NewChain(func(d game.Data) []action.Action {
			// We can have one or two bonfires
			var bonFirePositions []data.Position

			for _, o := range d.Objects {
				if o.Name == bonfireName {
					bonFirePositions = append(bonFirePositions, o.Position)
				}
			}

			a.logger.Info(fmt.Sprintf("Found %d bonfire positions", len(bonFirePositions)))

			var bonfireActions []action.Action

			for _, bonfirePos := range bonFirePositions {
				bonfireActions = append(bonfireActions,
					a.builder.MoveToCoords(bonfirePos, step.StopAtDistance(5)),
					action.NewChain(func(d game.Data) []action.Action {
						var chests []data.Object

						for _, o := range d.Objects {
							if slices.Contains(chestsIds, o.Name) && isChestWithinBonfireRange(o, bonfirePos) {
								chests = append(chests, o)
							}
						}

						var subActions []action.Action

						idx := 0
						position := d.PlayerUnit.Position
						ct := len(chests)
						for idx < ct {
							chest := chests[0]
							closestIdx := 0
							for i, chest2 := range chests {
								if pather.DistanceFromPoint(position, chest2.Position) < pather.DistanceFromPoint(position, chest.Position) {
									chest = chest2
									closestIdx = i
								}
							}
							//put the front chest where the closest was so we visit it later, then drop front
							chests[closestIdx] = chests[0]
							chests = chests[1:]

							subActions = append(subActions,
								a.builder.InteractObjectByID(chest.ID, func(d game.Data) bool {
									for _, obj := range d.Objects {
										isSameObj := obj.Name == chest.Name && obj.Position.X == chest.Position.X && obj.Position.Y == chest.Position.Y

										if isSameObj && !obj.Selectable {
											return true
										}
									}

									return false
								}),
								a.builder.Wait(200),
								a.builder.ItemPickup(false, 15),
							)
							idx = idx + 1
							//next position will be at this chest
							position = chest.Position
						}

						return subActions
					}),
				)
			}

			return bonfireActions
		}),
		// Make a path shorter for the next run if game exited instead of running in Act 3
		a.builder.ReturnTown(),
		a.builder.WayPoint(area.ThePandemoniumFortress),
	}

	return actions
}

func isChestWithinBonfireRange(chest data.Object, bonfirePosition data.Position) bool {
	distance := pather.DistanceFromPoint(chest.Position, bonfirePosition)

	return distance >= minChestDistanceFromBonfire && distance <= maxChestDistanceFromBonfire
}
