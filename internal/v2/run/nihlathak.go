package run

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/action"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/pather"
)

type Nihlathak struct {
	ctx *context.Status
}

func NewNihlathak() *Nihlathak {
	return &Nihlathak{
		ctx: context.Get(),
	}
}

func (n Nihlathak) Name() string {
	return string(config.NihlathakRun)
}

func (n Nihlathak) Run() error {

	// Use the waypoint to HallsOfPain
	err := action.WayPoint(area.HallsOfPain)
	if err != nil {
		return err
	}

	// Move to Halls Of Vaught
	if err = action.MoveToArea(area.HallsOfVaught); err != nil {
		return err
	}

	var nihlaObject data.Object

	// Move to Nihlathak
	action.MoveTo(func() (data.Position, bool) {

		for _, o := range n.ctx.Data.Objects {
			if o.Name == object.NihlathakWildernessStartPositionName {
				nihlaObject = o
				return o.Position, true
			}
		}

		return data.Position{}, false
	})

	// Try to find the safest place to move
	action.MoveTo(func() (data.Position, bool) {
		corners := [4]data.Position{
			{
				X: nihlaObject.Position.X + 20,
				Y: nihlaObject.Position.Y + 20,
			},
			{
				X: nihlaObject.Position.X - 20,
				Y: nihlaObject.Position.Y + 20,
			},
			{
				X: nihlaObject.Position.X - 20,
				Y: nihlaObject.Position.Y - 20,
			},
			{
				X: nihlaObject.Position.X + 20,
				Y: nihlaObject.Position.Y - 20,
			},
		}

		bestCorner := 0
		bestCornerDistance := 0
		for i, c := range corners {
			if pather.IsWalkable(c, n.ctx.Data.PlayerUnit.Position, n.ctx.Data.CollisionGrid) {
				averageDistance := 0
				for _, m := range n.ctx.Data.Monsters.Enemies() {
					averageDistance += pather.DistanceFromPoint(c, m.Position)
				}
				if averageDistance > bestCornerDistance {
					bestCorner = i
					bestCornerDistance = averageDistance
				}
				n.ctx.Logger.Debug("Corner", slog.Int("corner", i), slog.Int("monsters", len(n.ctx.Data.Monsters.Enemies())), slog.Int("distance", averageDistance))
			}
		}

		return corners[bestCorner], true
	})

	// Kill Nihlathak
	if err = n.ctx.Char.KillNihlathak(); err != nil {
		return err
	}

	// Clear monsters around the area, sometimes it makes difficult to pickup items if there are many monsters around the area
	if n.ctx.CharacterCfg.Game.Nihlathak.ClearArea {
		n.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies() {
				if d := pather.DistanceFromPoint(nihlaObject.Position, m.Position); d < 15 {
					n.ctx.Logger.Debug("Clearing monsters around Nihlathak position", slog.Any("monster", m))
					return m.UnitID, true
				}
			}

			return 0, false
		}, nil)
	}

	return nil
}
