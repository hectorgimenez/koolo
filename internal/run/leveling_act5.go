package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Leveling) act5() action.Action {
	return action.NewFactory(func(d data.Data) action.Action {
		if d.PlayerUnit.Area != area.Harrogath {
			return nil
		}

		return nil
	})
}

func (a Leveling) anya() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.CrystallinePassage),
			a.builder.MoveToArea(area.FrozenRiver),
			// TODO: Interact Anya
			a.builder.ReturnTown(),
			a.builder.IdentifyAll(false),
			a.builder.Stash(false),
			a.builder.ReviveMerc(),
			a.builder.Repair(),
			a.builder.VendorRefill(),
			// TODO: Interact Malah
			a.builder.UsePortalInTown(),
			// TODO: Interact Anya
		}
	})
}

func (a Leveling) ancients() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.TheAncientsWay),
			a.builder.MoveToArea(area.ArreatSummit),
		}
	})
}
