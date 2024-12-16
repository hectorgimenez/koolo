package run

import (
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
	// Use the waypoint to get to Arcane Sanctuary
	err := action.WayPoint(area.ArcaneSanctuary)
	if err != nil {
		return err
	}

	// Get the Summoner's position from the cached map data
	areaData := s.ctx.Data.Areas[area.ArcaneSanctuary]
	summonerNPC, found := areaData.NPCs.FindOne(npc.Summoner)
	if !found || len(summonerNPC.Positions) == 0 {
		return err
	}

	// Move to the Summoner's position using the static coordinates from map data
	if err = action.MoveToCoords(summonerNPC.Positions[0]); err != nil {
		return err
	}

	// Kill Summoner
	return s.ctx.Char.KillSummoner()
}
