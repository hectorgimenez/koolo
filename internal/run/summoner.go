package run

import (
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
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
	if err = s.ctx.Char.KillSummoner(); err != nil {
		return err
	}

	if s.ctx.CharacterCfg.Game.Summoner.ExitToA4 {

		// Move to tome for portal to canyon
		s.ctx.Logger.Debug("Moving to tome")
		tome, _ := s.ctx.Data.Objects.FindOne(object.YetAnotherTome)
		action.MoveToCoords(tome.Position)

		// Clear around tome and where red portal will spawn, monster can block tome or red portal interaction
		action.ClearAreaAroundPlayer(10, s.ctx.Data.MonsterFilterAnyReachable())

		// interact with tome to open red portal
		action.InteractObject(tome, func() bool {
			if _, found := s.ctx.Data.Objects.FindOne(object.PermanentTownPortal); found {
				s.ctx.Logger.Debug("Found red portal")
				return true
			}
			return false
		})

		// interact with red portal
		s.ctx.Logger.Debug("Moving to red portal")
		portal, _ := s.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
		action.MoveToCoords(portal.Position)

		action.InteractObject(portal, func() bool {
			return s.ctx.Data.PlayerUnit.Area == area.CanyonOfTheMagi
		})

		// Move to A4 if possible to shorten the run time
		err = action.WayPoint(area.ThePandemoniumFortress)
		if err != nil {
			return err
		}
	}

	return nil

}
