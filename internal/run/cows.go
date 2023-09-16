package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

type Cows struct {
	baseRun
}

func (a Cows) Name() string {
	return "Cows"
}

func (a Cows) BuildActions() (actions []action.Action) {
	return []action.Action{
		a.getWirtsLeg(),
		a.preparePortal(),
		a.builder.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
			return d.PlayerUnit.Area == area.MooMooFarm
		}),
		a.char.Buff(),
		a.builder.ClearArea(true, data.MonsterAnyFilter()),
	}
}

func (a Cows) getWirtsLeg() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		if _, found := d.Items.Find("WirtsLeg", item.LocationStash, item.LocationInventory); found {
			a.logger.Info("WirtsLeg found, skip finding it")
			return nil
		}

		return []action.Action{
			a.builder.WayPoint(area.StonyField), // Moving to starting point (Stony Field)
			action.NewChain(func(d data.Data) []action.Action {
				for _, o := range d.Objects {
					if o.Name == object.CairnStoneAlpha {
						return []action.Action{a.builder.MoveToCoords(o.Position)}
					}
				}

				return nil
			}),
			a.builder.ClearAreaAroundPlayer(10),
			a.builder.ItemPickup(false, 15),
			a.builder.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
				return d.PlayerUnit.Area == area.Tristram
			}, step.Wait(time.Second)),
			a.builder.InteractObject(object.WirtCorpse, func(d data.Data) bool {
				_, found := d.Items.Find("WirtsLeg")

				return found
			}),
			a.builder.ItemPickup(false, 30),
			a.builder.ReturnTown(),
		}
	})
}

func (a Cows) preparePortal() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		leg, found := d.Items.Find("WirtsLeg")
		if !found {
			a.logger.Error("WirtsLeg could not be found, portal cannot be opened")
			return nil
		}

		actions := []action.Action{
			a.builder.BuyAtVendor(npc.Akara, action.VendorItemRequest{
				Item:     item.TomeOfTownPortal,
				Quantity: 1,
				Tab:      4,
			}),
			action.NewChain(func(d data.Data) []action.Action {
				tpTome, found := d.Items.Find(item.TomeOfTownPortal, item.LocationInventory)
				if !found {
					a.logger.Error("TomeOfTownPortal could not be found, portal cannot be opened")
					return nil
				}

				return []action.Action{
					a.builder.CubeAddItems(leg, tpTome),
					a.builder.CubeTransmute(),
				}
			}),
		}

		return actions
	})
}
