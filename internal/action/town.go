package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func PreRun(firstRun bool) error {
	ctx := context.Get()

	DropMouseItem()
	step.SetSkill(skill.Vigor)
	RecoverCorpse()
	ManageBelt()
	// Just to make sure messages like TZ change or public game spam arent on the way
	ClearMessages()

	if firstRun {
		Stash(false)
	}

	UpdateQuestLog()

	// Store items that need to be left unidentified
	if HaveItemsToStashUnidentified() {
		Stash(firstRun)
	}

	// Identify - either via Cain or Tome
	IdentifyAll(firstRun)

	// Stash before vendor
	Stash(firstRun)

	// Refill pots, sell, buy etc
	VendorRefill(false, true)

	// Gamble
	Gamble()

	// Stash again if needed
	Stash(false)

	CubeRecipes()

	// Leveling related checks
	if ctx.CharacterCfg.Game.Leveling.EnsurePointsAllocation {
		ResetStats()
		EnsureStatPoints()
		EnsureSkillPoints()
	}

	if ctx.CharacterCfg.Game.Leveling.EnsureKeyBinding {
		EnsureSkillBindings()
	}

	HealAtNPC()
	ReviveMerc()
	HireMerc()

	return Repair()
}

func InRunReturnTownRoutine() error {
	ctx := context.Get()

	ReturnTown()
	step.SetSkill(skill.Vigor)
	RecoverCorpse()
	ManageBelt()

	// Let's stash items that need to be left unidentified
	if ctx.CharacterCfg.Game.UseCainIdentify && HaveItemsToStashUnidentified() {
		Stash(false)
	}

	IdentifyAll(false)

	VendorRefill(false, true)
	Stash(false)
	Gamble()
	Stash(false)
	CubeRecipes()

	if ctx.CharacterCfg.Game.Leveling.EnsurePointsAllocation {
		EnsureStatPoints()
		EnsureSkillPoints()
	}

	if ctx.CharacterCfg.Game.Leveling.EnsureKeyBinding {
		EnsureSkillBindings()
	}

	HealAtNPC()
	ReviveMerc()
	HireMerc()
	Repair()
	
	if (ctx.CharacterCfg.Companion.Leader) {
		UsePortalInTown()
		utils.Sleep(500)
		return OpenTPIfLeader()
	}
	
	return UsePortalInTown()
}
