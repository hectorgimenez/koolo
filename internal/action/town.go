package action

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
)

func PreRun(firstRun bool) error {
	ctx := context.Get()

	DropMouseItem()
	step.SetSkill(skill.Vigor)
	RecoverCorpse()
	ManageBelt()

	if err := IdentifyAll(false); err != nil {
		return err
	}

	UpdateQuestLog()
	VendorRefill(false, true)
	Stash(firstRun)
	Gamble(time.Now())
	Stash(false)
	CubeRecipes()

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

	/*
		This will be added when option for cain Identify is added

		// Let's stash items that need to be left unidentified
		if HaveItemsToStashUnidentified() {
			Stash(false)
		}
	*/

	IdentifyAll(false)

	VendorRefill(false, true)
	Stash(false)
	Gamble(time.Now())
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

	return UsePortalInTown()
}
