package run

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Eldritch struct {
	ctx *context.Status
}

func NewEldritch() *Eldritch {
	return &Eldritch{
		ctx: context.Get(),
	}
}

func (e Eldritch) Name() string {
	return string(config.EldritchRun)
}

func (e Eldritch) Run() error {
	// First return to town if we're not already there
	if !e.ctx.Data.PlayerUnit.Area.IsTown() {
		if err := action.ReturnTown(); err != nil {
			return err
		}
	}

	// Travel to FrigidHighlands using waypoint
	err := action.WayPoint(area.FrigidHighlands)
	if err != nil {
		return err
	}

	// Kill Eldritch
	e.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(npc.MinionExp, data.MonsterTypeSuperUnique); found {
			return m.UnitID, true
		}

		return 0, false
	}, nil)

	// Move to Shenk and kill him, if enabled
	if e.ctx.CharacterCfg.Game.Eldritch.KillShenk {
		// Move into position
		if err = action.MoveToCoords(data.Position{X: 3876, Y: 5130}); err != nil {
			return errors.New("failed to move to shenk")
		}

		// Kill Shenk
		err := e.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			if m, found := d.Monsters.FindOne(npc.OverSeer, data.MonsterTypeSuperUnique); found {
				if m.Stats[stat.Life] > 0 {
					return m.UnitID, true
				}
				return 0, false
			}

			return 0, false
		}, nil)

		// We don't want to return to town if this is the last run
		if !isLastRun(e.ctx) {
			if err := action.ReturnTown(); err != nil {
				return err
			}
		}

		return err
	}

	// Return to town after Eldritch (if we didn't do Shenk) and it's not the last run
	if !isLastRun(e.ctx) {
		return action.ReturnTown()
	}

	return nil
}

// Helper function to check if this is the last run
func isLastRun(ctx *context.Status) bool {
	runs := ctx.CharacterCfg.Game.Runs
	return len(runs) > 0 && runs[len(runs)-1] == config.EldritchRun
}
