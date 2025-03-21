package run

import (
	"errors"
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
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
	action.OpenTPIfLeader()
	// Move to Halls Of Vaught
	if err = action.MoveToArea(area.HallsOfVaught); err != nil {
		return err
	}
	action.OpenTPIfLeader()
	var nihlaObject data.Object

	o, found := n.ctx.Data.Objects.FindOne(object.NihlathakWildernessStartPositionName)
	if !found {
		return errors.New("failed to find Nihlathak's Start Position")
	}

	// Move to Nihlathak
	action.Buff()
	action.MoveToCoords(o.Position)
	action.OpenTPIfLeader()

	// Try to position in the safest corner
	action.MoveToCoords(n.findBestCorner(o.Position))

	// Disable item pickup before the fight
	n.ctx.DisableItemPickup()

	// Kill Nihlathak
	if err = n.ctx.Char.KillNihlathak(); err != nil {
		// Re-enable item pickup even if kill fails
		n.ctx.EnableItemPickup()
		return err
	}

	// Re-enable item pickup after kill
	n.ctx.EnableItemPickup()

	// Clear monsters around the area, sometimes it makes difficult to pickup items if there are many monsters around the area
	if n.ctx.CharacterCfg.Game.Nihlathak.ClearArea {
		n.ctx.Logger.Debug("Clearing monsters around Nihlathak position")

		n.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies() {
				if d := pather.DistanceFromPoint(nihlaObject.Position, m.Position); d < 15 {
					return m.UnitID, true
				}
			}

			return 0, false
		}, nil)
	}

	return nil
}

func (n Nihlathak) findBestCorner(nihlathakPosition data.Position) data.Position {
	corners := [4]data.Position{
		{
			X: nihlathakPosition.X + 20,
			Y: nihlathakPosition.Y + 20,
		},
		{
			X: nihlathakPosition.X - 20,
			Y: nihlathakPosition.Y + 20,
		},
		{
			X: nihlathakPosition.X - 20,
			Y: nihlathakPosition.Y - 20,
		},
		{
			X: nihlathakPosition.X + 20,
			Y: nihlathakPosition.Y - 20,
		},
	}

	bestCorner := 0
	bestCornerDistance := 0
	for i, c := range corners {
		if n.ctx.Data.AreaData.IsWalkable(c) {
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

	return corners[bestCorner]
}
