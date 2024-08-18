package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/v2/action/step"
	"github.com/hectorgimenez/koolo/internal/v2/context"
)

func OpenTPIfLeader() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "OpenTPIfLeader"

	isLeader := ctx.CharacterCfg.Companion.Enabled && ctx.CharacterCfg.Companion.Leader

	if isLeader {
		return step.OpenPortal()
	}

	return nil
}

func isMonsterSealElite(monster data.Monster) bool {
	if monster.Type == data.MonsterTypeSuperUnique && (monster.Name == npc.OblivionKnight || monster.Name == npc.VenomLord || monster.Name == npc.StormCaster) {
		return true
	}

	return false
}

func PostRun(isLastRun bool) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "PostRun"

	// For companions, we don't need them to do anything, just follow the leader
	if ctx.CharacterCfg.Companion.Enabled && !ctx.CharacterCfg.Companion.Leader {
		return nil
	}

	ClearAreaAroundPlayer(5, data.MonsterAnyFilter())
	ItemPickup(-1)

	// Don't return town on last run
	if !isLastRun {
		return ReturnTown()
	}

	return nil
}