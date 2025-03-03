package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/lxn/win"
)

func (a Leveling) act1() error {
	running := false
	if running || a.ctx.Data.PlayerUnit.Area != area.RogueEncampment {
		return nil
	}

	running = true

	// Clear Den of Evil til level 3 - might need to run it in each difficulty if we need more than one respec
	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 3 {
		a.ctx.Logger.Debug("Current lvl %s under 3 - Leveling in Den of Evil")
		return NewQuests().clearDenQuest()
	}
	// Do Cold Plains til level 6
	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 6 {
		return a.coldPlains()
	}

	// Do Stony Field until 9
	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 9 {
		return a.stonyField()
	}

	// Do Countess Runs until level 14 - skipping in Hell because cold/fire immune - there's no point getting stuck if our merc isn't geared enough yet

	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 14 || a.ctx.Data.CharacterCfg.Game.Difficulty == difficulty.Nightmare {
		if !a.ctx.Data.CanTeleport() {
			a.ctx.CharacterCfg.Game.Countess.ClearFloors = true
		}
		return NewCountess().Run()
	}

	if a.ctx.Data.Quests[quest.Act1SistersToTheSlaughter].Completed() {
		action.ReturnTown()
		// Do Den of Evil if not complete before moving acts
		if !a.ctx.Data.Quests[quest.Act1DenOfEvil].Completed() {
			NewQuests().clearDenQuest()
		}
		if !a.isCainInTown() && !a.ctx.Data.Quests[quest.Act1TheSearchForCain].Completed() {
			NewQuests().rescueCainQuest()
		}

		action.InteractNPC(npc.Warriv)
		a.ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)

		return nil
	} else {
		return NewAndariel().Run()
	}
}

func (a Leveling) coldPlains() error {
	err := action.WayPoint(area.ColdPlains)
	if err != nil {
		return err
	}

	return action.ClearCurrentLevel(false, data.MonsterAnyFilter())
}

func (a Leveling) stonyField() error {
	err := action.WayPoint(area.StonyField)
	if err != nil {
		return err
	}

	return action.ClearCurrentLevel(false, data.MonsterAnyFilter())
}

func (a Leveling) isCainInTown() bool {
	_, found := a.ctx.Data.Monsters.FindOne(npc.DeckardCain5, data.MonsterTypeNone)

	return found
}
