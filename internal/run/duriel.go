package run

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

var talTombs = []area.ID{area.TalRashasTomb1, area.TalRashasTomb2, area.TalRashasTomb3, area.TalRashasTomb4, area.TalRashasTomb5, area.TalRashasTomb6, area.TalRashasTomb7}

type Duriel struct {
	ctx *context.Status
}

func NewDuriel() *DrifterCavern {
	return &DrifterCavern{
		ctx: context.Get(),
	}
}

func (d Duriel) Name() string {
	return string(config.DurielRun)
}

func (d Duriel) Run() error {

	// Use the waypoint
	err := action2.WayPoint(area.CanyonOfTheMagi)
	if err != nil {
		return err
	}

	// Move to the real Tal Rasha tomb
	if realTalRashaTomb, err := d.findRealTomb(); err != nil {
		return err
	} else {
		action2.MoveToArea(realTalRashaTomb)
	}

	// Get Orifice position and move to it, clear surrounding area
	if orifice, found := d.ctx.Data.Objects.FindOne(object.HoradricOrifice); found {
		action2.MoveToCoords(orifice.Position)
		action2.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
	} else {
		return errors.New("failed to find Duriel's Lair entrance")
	}

	// Buff before we enter :)
	action2.Buff()

	// Find Duriel's entrance and enter
	if portal, found := d.ctx.Data.Objects.FindOne(object.DurielsLairPortal); found {
		action2.InteractObject(portal, func() bool {
			return d.ctx.Data.PlayerUnit.Area == area.DurielsLair
		})
	}

	return d.ctx.Char.KillDuriel()
}

func (d Duriel) findRealTomb() (area.ID, error) {
	var realTomb area.ID

	for _, tomb := range talTombs {
		for _, obj := range d.ctx.Data.Areas[tomb].Objects {
			if obj.Name == object.HoradricOrifice {
				realTomb = tomb
				break
			}
		}
	}

	if realTomb == 0 {
		return 0, errors.New("failed to find the real Tal Rasha tomb")
	}

	return realTomb, nil
}
