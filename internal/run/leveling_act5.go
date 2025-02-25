package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Leveling) act5() error {
	if a.ctx.Data.PlayerUnit.Area != area.Harrogath {
		return nil
	}

	if a.ctx.Data.Quests[quest.Act5RiteOfPassage].Completed() {
		a.ctx.Logger.Info("Starting Baal run...")
		if a.ctx.CharacterCfg.Game.Difficulty != difficulty.Normal {
			a.ctx.CharacterCfg.Game.Baal.SoulQuit = true
		}
		NewBaal(nil).Run()

		lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
		if a.ctx.Data.PlayerUnit.Area == area.TheWorldstoneChamber && len(a.ctx.Data.Monsters.Enemies()) == 0 {
			switch a.ctx.CharacterCfg.Game.Difficulty {
			case difficulty.Normal:
				if lvl.Value >= 41 {
					a.ctx.CharacterCfg.Game.Difficulty = difficulty.Nightmare
				}
			case difficulty.Nightmare:
				if lvl.Value >= 65 {
					a.ctx.CharacterCfg.Game.Difficulty = difficulty.Hell
				}
			}
		}
		return nil

	}

	wp, _ := a.ctx.Data.Objects.FindOne(object.ExpansionWaypoint)
	action.MoveToCoords(wp.Position)

	if _, found := a.ctx.Data.Monsters.FindOne(npc.Drehya, data.MonsterTypeNone); !found {
		NewQuests().rescueAnyaQuest()
	}

	err := NewQuests().killAncientsQuest()
	if err != nil {
		return err
	}

	return nil
}
