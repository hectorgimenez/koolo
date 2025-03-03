package run

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxOrificeAttempts = 10
	orificeCheckDelay  = 200
)

var talTombs = []area.ID{area.TalRashasTomb1, area.TalRashasTomb2, area.TalRashasTomb3, area.TalRashasTomb4, area.TalRashasTomb5, area.TalRashasTomb6, area.TalRashasTomb7}

type Duriel struct {
	ctx *context.Status
}

func NewDuriel() *Duriel {
	return &Duriel{
		ctx: context.Get(),
	}
}

func (d Duriel) Name() string {
	return string(config.DurielRun)
}

func (d Duriel) Run() error {
	err := action.WayPoint(area.CanyonOfTheMagi)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	// Find and move to the real Tal Rasha tomb
	realTalRashaTomb, err := d.findRealTomb()
	if err != nil {
		return err
	}

	err = action.MoveToArea(realTalRashaTomb)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	// Wait for area to fully load and get synchronized
	utils.Sleep(500)
	d.ctx.RefreshGameData()

	// Find orifice with retry logic
	var orifice data.Object
	var found bool

	for attempts := 0; attempts < maxOrificeAttempts; attempts++ {
		orifice, found = d.ctx.Data.Objects.FindOne(object.HoradricOrifice)
		if found && orifice.Mode == mode.ObjectModeOpened {
			break
		}
		utils.Sleep(orificeCheckDelay)
		d.ctx.RefreshGameData()
	}

	if !found {
		return errors.New("failed to find Duriel's Lair entrance after multiple attempts")
	}

	// Move to orifice and clear the area
	err = action.MoveToCoords(orifice.Position)
	if err != nil {
		return err
	}

	action.OpenTPIfLeader()
	err = action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	// Pre-fight buff
	action.Buff()

	// Find portal and enter Duriel's Lair
	var portal data.Object
	for attempts := 0; attempts < maxOrificeAttempts; attempts++ {
		portal, found = d.ctx.Data.Objects.FindOne(object.DurielsLairPortal)
		if found && portal.Mode == mode.ObjectModeOpened {
			break
		}
		utils.Sleep(orificeCheckDelay)
		d.ctx.RefreshGameData()
	}

	if !found {
		return errors.New("failed to find Duriel's portal after multiple attempts")
	}

	//Exception: Duriel Lair portal has no destination in memory
	err = action.InteractObject(portal, func() bool {
		return d.ctx.Data.PlayerUnit.Area == area.DurielsLair
	})
	if err != nil {
		return err
	}

	// Final refresh before fight
	d.ctx.RefreshGameData()

	utils.Sleep(700)
	action.OpenTPIfLeader()
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
