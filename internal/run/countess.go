package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Countess struct {
	ctx *context.Status
}

func NewCountess() *Countess {
	return &Countess{
		ctx: context.Get(),
	}
}

func (c Countess) Name() string {
	return string(config.CountessRun)
}

func (c Countess) Run() error {
	// Travel to boss level
	err := action2.WayPoint(area.BlackMarsh)
	if err != nil {
		return err
	}

	areas := []area.ID{
		area.ForgottenTower,
		area.TowerCellarLevel1,
		area.TowerCellarLevel2,
		area.TowerCellarLevel3,
		area.TowerCellarLevel4,
		area.TowerCellarLevel5,
	}

	for _, a := range areas {
		err = action2.MoveToArea(a)
		if err != nil {
			return err
		}
	}

	// Try to move around Countess area
	action2.MoveTo(func() (data.Position, bool) {
		for _, o := range c.ctx.Data.Objects {
			if o.Name == object.GoodChest {
				return o.Position, true
			}
		}

		// Try to teleport over Countess in case we are not able to find the chest position, a bit more risky
		if countess, found := c.ctx.Data.Monsters.FindOne(npc.DarkStalker, data.MonsterTypeSuperUnique); found {
			return countess.Position, true
		}

		return data.Position{}, false
	})

	// Kill Countess
	return c.ctx.Char.KillCountess()
}
