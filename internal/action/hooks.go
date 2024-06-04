package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

/*
Hooks file provide an easy way to execute specific actions in specific stages of the bot execution.
Be sure to do appropriate checks to avoid infinite loops or unwanted behavior, specially on hooks that are executed often.
*/

// NewGameHook is executed when a new game is created. Actions returned here will be executed when a new game is created before the main actions.
func (b *Builder) NewGameHook() []Action {
	return []Action{
		b.SwitchToLegacyMode(),
	}
}

// PreRunHook is executed before the main actions are executed. Actions returned here will be executed before the main actions for each run.
func (b *Builder) PreRunHook(firstRun bool) []Action {
	return b.PreRun(firstRun)
}

// EachLoopHook is executed every loop iteration before the main actions are executed. If you return any actions here, they will be executed before the main actions.
// This is useful to execute actions that are high priority and should be executed before the main actions, since will interrupt the main actions.
// For example, buffing, healing, going back to town to buy pots, revive merc...
func (b *Builder) EachLoopHook(d game.Data) (actions []Action) {
	// Check if we have HP & MP potions
	_, healingPotsFound := d.Inventory.Belt.GetFirstPotion(data.HealingPotion)
	_, manaPotsFound := d.Inventory.Belt.GetFirstPotion(data.ManaPotion)

	// Check if we need to go back to town (no pots or merc died)
	if (d.CharacterCfg.BackToTown.NoHpPotions && !healingPotsFound ||
		d.CharacterCfg.BackToTown.NoMpPotions && !manaPotsFound ||
		d.CharacterCfg.BackToTown.MercDied && d.Data.MercHPPercent() <= 0 && d.CharacterCfg.Character.UseMerc) &&
		!d.PlayerUnit.Area.IsTown() {
		actions = append(actions, b.InRunReturnTownRoutine()...)
	}

	if b.IsRebuffRequired(d) {
		actions = append(actions, b.BuffIfRequired(d))
	}

	return
}

// PostRunHook is executed after the main actions are executed. Actions returned here will be executed after the main actions for each run.
func (b *Builder) PostRunHook(isLastRun bool) (actions []Action) {
	// For companions, we don't need them to do anything, just follow the leader
	if config.Characters[b.Supervisor].Companion.Enabled && !config.Characters[b.Supervisor].Companion.Leader {
		return []Action{}
	}

	actions = append(actions,
		b.ClearAreaAroundPlayer(5, data.MonsterAnyFilter()),
		b.ItemPickup(true, -1),
	)

	// Don't return town on last run
	if !isLastRun {
		if config.Characters[b.Supervisor].Game.ClearTPArea {
			actions = append(actions, b.ClearAreaAroundPlayer(5, data.MonsterAnyFilter()))
			actions = append(actions, b.ItemPickup(false, -1))
		}
		actions = append(actions, b.ReturnTown())
	}

	return
}
