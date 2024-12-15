package run

import (
    "errors"

    "github.com/hectorgimenez/d2go/pkg/data"
    "github.com/hectorgimenez/d2go/pkg/data/area"
    "github.com/hectorgimenez/d2go/pkg/data/npc"
    "github.com/hectorgimenez/d2go/pkg/data/object"
    "github.com/hectorgimenez/koolo/internal/action"
    "github.com/hectorgimenez/koolo/internal/config"
    "github.com/hectorgimenez/koolo/internal/context"
)

var fixedPlaceNearRedPortal = data.Position{
    X: 5130,
    Y: 5120,
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
    // First return to town if we're not already there
    if !p.ctx.Data.PlayerUnit.Area.IsTown() {
        if err := action.ReturnTown(); err != nil {
            return err
        }
    }

    // Get to Harrogath
    err := action.WayPoint(area.Harrogath)
    if err != nil {
        return err
    }

    // Move near the red portal
    _ = action.MoveToCoords(fixedPlaceNearRedPortal)

    // Find and use the red portal
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

    // Get NPCs from cached map data
    for _, npcData := range p.ctx.Data.Areas[area.NihlathaksTemple].NPCs {
        if npcData.ID == npc.DefiledWarrior {
            // Let the character implementation handle the actual attack
            return p.ctx.Char.KillPindle()
        }
    }

    return errors.New("pindleskin not found")
}
