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
	X: 5130,
	Y: 5120,
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
		return p.ctx.Data.PlayerUnit.Area == area.NihlathaksTemple
	})
	if err != nil {
		return err
	}

	_ = action.MoveToCoords(pindleSafePosition)

	return p.ctx.Char.KillPindle()
}
