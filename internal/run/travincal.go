package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Council struct {
	ctx *context.Status
}

func NewTravincal() *Council {
	return &Council{
		ctx: context.Get(),
	}
}

func (s Council) Name() string {
	return string(config.TravincalRun)
}

func (s Council) Run() error {
	err := action.WayPoint(area.Travincal) // Moving to starting point (Travincal)
	if err != nil {
		return err
	}

	for _, al := range s.ctx.Data.AdjacentLevels {
		if al.Area == area.DuranceOfHateLevel1 {
			err = action.MoveToCoords(data.Position{
				X: al.Position.X - 1,
				Y: al.Position.Y + 3,
			})
			if err != nil {
				s.ctx.Logger.Warn("Error moving to council area", err)
			}
		}
	}

	return s.ctx.Char.KillCouncil()
}
