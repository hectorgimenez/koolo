package run

import (
	"errors"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

var fixedPlaceNearRedPortal = data.Position{
	X: 5131, // Adjusted coords from X: 5130 to prevent bot stuck
	Y: 5123, // Adjusted coords from Y: 5120 to prevent bot stuck
}

var pindleSafePosition = data.Position{
	X: 10058,
	Y: 13236,
}

type Pindleskin struct {
	ctx *context.Status
}

func NewPindleskin() *Pindleskin {
	return &Pindleskin{
		ctx: context.Get(),
	}
}

func (p Pindleskin) Name() string {
	return string(config.PindleskinRun)
}

func (p Pindleskin) Run() error {
	err := action.WayPoint(area.Harrogath)
	if err != nil {
		return err
	}

	_ = action.MoveToCoords(fixedPlaceNearRedPortal)

	redPortal, found := p.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("red portal not found")
	}

	err = action.InteractObject(redPortal, func() bool {
		return p.ctx.Data.AreaData.Area == area.NihlathaksTemple && p.ctx.Data.AreaData.IsInside(p.ctx.Data.PlayerUnit.Position)
	})
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	action.Buff()
	_ = action.MoveToCoords(pindleSafePosition)

	_ = p.ctx.Char.KillPindle()

	if p.ctx.CharacterCfg.Game.Pindleskin.KillNihlathak {
		_ = action.MoveToArea(area.HallsOfAnguish)
		_ = action.MoveToArea(area.HallsOfPain)
		_ = action.MoveToArea(area.HallsOfVaught)

		o, found := p.ctx.Data.Objects.FindOne(object.NihlathakWildernessStartPositionName)
		if !found {
			return errors.New("failed to find Nihlathak's Start Position")
		}

		// Move to Nihlathak
		action.Buff()

		if err = action.MoveToCoords(o.Position); err != nil {
			return err
		}

		// Disable item pickup before the fight
		p.ctx.DisableItemPickup()

		// Kill Nihlathak
		if err = p.ctx.Char.KillNihlathak(); err != nil {
			// Re-enable item pickup even if kill fails
			p.ctx.EnableItemPickup()
			return err
		}

		// Re-enable item pickup after kill
		p.ctx.EnableItemPickup()
	}

	return nil
}
