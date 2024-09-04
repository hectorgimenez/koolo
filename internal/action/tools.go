package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func OpenTPIfLeader() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "OpenTPIfLeader"

	isLeader := ctx.CharacterCfg.Companion.Leader

	if isLeader {
		return step.OpenPortal()
	}

	return nil
}

func IsMonsterSealElite(monster data.Monster) bool {
	if monster.Type == data.MonsterTypeSuperUnique && (monster.Name == npc.OblivionKnight || monster.Name == npc.VenomLord || monster.Name == npc.StormCaster) {
		return true
	}

	return false
}

func PostRun(isLastRun bool) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "PostRun"

	// Allow some time for items drop to the ground, otherwise we might miss some
	utils.Sleep(200)
	ClearAreaAroundPlayer(5, data.MonsterAnyFilter())
	ItemPickup(-1)

	// Don't return town on last run
	if !isLastRun {
		return ReturnTown()
	}

	return nil
}
