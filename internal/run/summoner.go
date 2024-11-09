package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Summoner struct {
	ctx *context.Status
}

func NewSummoner() *Summoner {
	return &Summoner{
		ctx: context.Get(),
	}
}

func (s Summoner) Name() string {
	return string(config.SummonerRun)
}

func (s Summoner) Run() error {

	// Use the waypoint
	err := action.WayPoint(area.ArcaneSanctuary)
	if err != nil {
		return err
	}

	// Move to boss position
	if err = action.MoveTo(func() (data.Position, bool) {
		m, found := s.ctx.Data.NPCs.FindOne(npc.Summoner)
		return m.Positions[0], found
	}); err != nil {
		return err
	}

	// Kill Summoner
	return s.ctx.Char.KillSummoner()
}
