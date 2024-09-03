package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type TalRashaTombs struct {
	ctx *context.Status
}

func NewTalRashaTombs() *TalRashaTombs {
	return &TalRashaTombs{
		ctx: context.Get(),
	}
}

func (a TalRashaTombs) Name() string {
	return string(config.TalRashaTombsRun)
}

var talRashaTombs = []area.ID{area.TalRashasTomb1, area.TalRashasTomb2, area.TalRashasTomb3, area.TalRashasTomb4, area.TalRashasTomb5, area.TalRashasTomb6, area.TalRashasTomb7}

func (a TalRashaTombs) Run() error {

	// Iterate over all Tal Rasha Tombs
	for _, tomb := range talRashaTombs {

		// Use the waypoint to travel to Canyon Of The Magi
		err := action2.WayPoint(area.CanyonOfTheMagi)
		if err != nil {
			return err
		}

		// Move to the next Tomb
		if err = action2.MoveToArea(tomb); err != nil {
			return err
		}

		// Open a TP if we're the leader
		action2.OpenTPIfLeader()

		// Buff before we start
		action2.Buff()

		// Clear the Tomb
		if err = action2.ClearCurrentLevel(true, data.MonsterAnyFilter()); err != nil {
			return err
		}

		// Return to town
		if err = action2.ReturnTown(); err != nil {
			return err
		}
	}

	return nil
}
