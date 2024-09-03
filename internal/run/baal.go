package run

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	context2 "github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

var baalThronePosition = data.Position{
	X: 15095,
	Y: 5042,
}

type Baal struct {
	ctx *context2.Status
}

func NewBaal() *Baal {
	return &Baal{
		ctx: context2.Get(),
	}
}

func (s Baal) Name() string {
	return string(config.BaalRun)
}

func (s Baal) Run() error {
	// Set filter
	filter := data.MonsterAnyFilter()
	if s.ctx.CharacterCfg.Game.Baal.OnlyElites {
		filter = data.MonsterEliteFilter()
	}

	err := action2.WayPoint(area.TheWorldStoneKeepLevel2)
	if err != nil {
		return err
	}

	if s.ctx.CharacterCfg.Game.Baal.ClearFloors {
		action2.ClearCurrentLevel(false, filter)
	}

	err = action2.MoveToArea(area.TheWorldStoneKeepLevel3)
	if err != nil {
		return err
	}

	if s.ctx.CharacterCfg.Game.Baal.ClearFloors {
		action2.ClearCurrentLevel(false, filter)
	}

	err = action2.MoveToArea(area.ThroneOfDestruction)
	if err != nil {
		return err
	}
	err = action2.MoveToCoords(baalThronePosition)
	if err != nil {
		return err
	}
	if s.checkForSoulsOrDolls() {
		return errors.New("souls or dolls detected, skipping")
	}

	// Let's move to a safe area and open the portal in companion mode
	if s.ctx.CharacterCfg.Companion.Leader {
		action2.MoveToCoords(data.Position{
			X: 15116,
			Y: 5071,
		})
		action2.OpenTPIfLeader()
	}

	err = action2.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	// Force rebuff before waves
	action2.Buff()

	// Come back to previous position
	err = action2.MoveToCoords(baalThronePosition)
	if err != nil {
		return err
	}

	lastWave := false
	for !lastWave {
		if _, found := s.ctx.Data.Monsters.FindOne(npc.BaalsMinion, data.MonsterTypeMinion); found {
			lastWave = true
		}

		enemies := false
		for _, e := range s.ctx.Data.Monsters.Enemies() {
			dist := pather.DistanceFromPoint(baalThronePosition, e.Position)
			if dist < 50 {
				enemies = true
			}
		}
		if enemies {
			err = action2.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())
			if err != nil {
				return err
			}
		}
		action2.MoveToCoords(baalThronePosition)
	}

	_, isLevelingChar := s.ctx.Char.(context2.LevelingCharacter)
	if s.ctx.CharacterCfg.Game.Baal.KillBaal || isLevelingChar {
		utils.Sleep(10000)
		action2.Buff()
		baalPortal, _ := s.ctx.Data.Objects.FindOne(object.BaalsPortal)
		err = action2.InteractObjectByID(baalPortal.ID, func() bool {
			return s.ctx.Data.PlayerUnit.Area == area.TheWorldstoneChamber
		})
		if err != nil {
			return err
		}
		return s.ctx.Char.KillBaal()
	}

	return nil
}

func (s Baal) checkForSoulsOrDolls() bool {
	var npcIds []npc.ID

	if s.ctx.CharacterCfg.Game.Baal.DollQuit {
		npcIds = append(npcIds, npc.UndeadStygianDoll2, npc.UndeadSoulKiller2)
	}
	if s.ctx.CharacterCfg.Game.Baal.SoulQuit {
		npcIds = append(npcIds, npc.BlackSoul2, npc.BurningSoul2)
	}

	for _, id := range npcIds {
		if _, found := s.ctx.Data.Monsters.FindOne(id, data.MonsterTypeNone); found {
			return true
		}
	}

	return false
}
