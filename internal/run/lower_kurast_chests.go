package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/pather"
	"slices"
)

var bonfireName = object.Name(160)
var chestsIds = []object.Name{object.Name(183), object.Name(181)}
var minChestDistanceFromBonfire = 25
var maxChestDistanceFromBonfire = 45

type LowerKurastChest struct {
	baseRun
}

func (a LowerKurastChest) Name() string {
	return "LowerKurastChest"
}

func (a LowerKurastChest) BuildActions() []action.Action {
	actions := []action.Action{
		a.builder.WayPoint(area.LowerKurast),
		action.NewChain(func(d data.Data) []action.Action {
			// We can have one or two bonfires
			var bonFirePositions []data.Position

			for _, o := range d.Objects {
				if o.Name == bonfireName {
					bonFirePositions = append(bonFirePositions, o.Position)
				}
			}

			var bonfireActions []action.Action

			for _, bonfirePos := range bonFirePositions {
				bonfireActions = append(bonfireActions,
					a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
						return bonfirePos, true
					}),
					action.NewChain(func(d data.Data) []action.Action {
						var chests []data.Object

						for _, o := range d.Objects {
							if slices.Contains(chestsIds, o.Name) && isChestWithinBonfireRange(o, bonfirePos) {
								chests = append(chests, o)
							}
						}

						var subActions []action.Action

						for _, chest := range chests {
							subActions = append(subActions,
								a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
									return chest.Position, true
								}),
								a.builder.InteractObject(chest.Name, func(d data.Data) bool {
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
