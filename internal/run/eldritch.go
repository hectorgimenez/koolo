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
	ctx := context.Get()
	// Travel to FrigidHighlands
	err := action.WayPoint(area.FrigidHighlands)
	if err != nil {
		return err
	}

	// Kill Eldritch
	e.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(npc.MinionExp, data.MonsterTypeSuperUnique); found {
			monsterIsImmune := false
			for _, resist := range ctx.Data.CharacterCfg.Runtime.ImmunityFilter {
				if m.IsImmune(resist) {
					monsterIsImmune = true
					break
				}
			}

			if monsterIsImmune {
				return 0, false
			}

			return m.UnitID, true
		}

		return 0, false
	})

	// Move to Shenk and kill him, if enabled
	if e.ctx.CharacterCfg.Game.Eldritch.KillShenk {
		// Move into position
		if err = action.MoveToCoords(data.Position{X: 3876, Y: 5130}); err != nil {
			return errors.New("failed to move to shenk")
		}

		// Kill Shenk
		return e.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			if m, found := d.Monsters.FindOne(npc.OverSeer, data.MonsterTypeSuperUnique); found {
				if m.Stats[stat.Life] > 0 {
					return m.UnitID, true
				}
				return 0, false
			}

			return 0, false
		})
	}

	return nil
}
