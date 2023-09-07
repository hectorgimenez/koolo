package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
)

type Countess struct {
	baseRun
}

func (c Countess) Name() string {
	return "Countess"
}

func (c Countess) BuildActions() (actions []action.Action) {
	// Travel to boss level
	actions = append(actions,
		c.builder.WayPoint(area.BlackMarsh), // Moving to starting point (Black Marsh)
		c.builder.MoveToArea(area.ForgottenTower),
		c.builder.MoveToArea(area.TowerCellarLevel1),
		c.builder.MoveToArea(area.TowerCellarLevel2),
		c.builder.MoveToArea(area.TowerCellarLevel3),
		c.builder.MoveToArea(area.TowerCellarLevel4),
		c.builder.MoveToArea(area.TowerCellarLevel5),
	)

	// Try to move around Countess area
	actions = append(actions, c.builder.MoveTo(func(d data.Data) (data.Position, bool) {
		for _, o := range d.Objects {
			if o.Name == object.GoodChest {
				return o.Position, true
			}
		}

		// Try to teleport over Countess in case we are not able to find the chest position, a bit more risky
		if countess, found := d.Monsters.FindOne(npc.DarkStalker, data.MonsterTypeSuperUnique); found {
			return countess.Position, true
		}

		return data.Position{}, false
	}))

	// Kill Countess
	return append(actions, c.char.KillCountess())
}
