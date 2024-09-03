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
	err := action2.WayPoint(area.Harrogath)
	if err != nil {
		return err
	}

	_ = action2.MoveToCoords(fixedPlaceNearRedPortal)

	redPortal, found := p.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("red portal not found")
	}

	err = action2.InteractObject(redPortal, func() bool {
		return p.ctx.Data.PlayerUnit.Area == area.NihlathaksTemple
	})
	if err != nil {
		return err
	}

	_ = action2.MoveToCoords(pindleSafePosition)

	return p.ctx.Char.KillPindle()
}
