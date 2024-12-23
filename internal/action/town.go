package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
)

func PreRun(firstRun bool) error {
	ctx := context.Get()

	DropMouseItem()
	step.SetSkill(skill.Vigor)

	// Always recover corpse, regardless of town status
	RecoverCorpse()

	ManageBelt()

	if firstRun {
		Stash(firstRun)
	}

	UpdateQuestLog()

	if HaveItemsToStashUnidentified() {
		Stash(firstRun)
	}

	if ctx.Data.PlayerUnit.Area.IsTown() {
		// Identify - either via Cain or Tome
		IdentifyAll(firstRun)

		// Stash before vendor
		Stash(firstRun)

		// Refill pots, sell, buy etc
		VendorRefill(false, true)
		Gamble()

		// Stash again if needed
		Stash(false)

		// Perform cube recipes
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
		if RepairRequired() {
			Repair()
		}
	}

	return nil
}

func InRunReturnTownRoutine() error {
	ctx := context.Get()

	// Track if we need repair before we even start the return
	needsRepair := RepairRequired()

	// If not in town, initiate return to town logic
	if !ctx.Data.PlayerUnit.Area.IsTown() {
		ReturnTown()
		return nil
	}

	// Always recover corpse, regardless of town status
	RecoverCorpse()

	// Only proceed with town actions if we actually made it to town
	if ctx.Data.PlayerUnit.Area.IsTown() {
		step.SetSkill(skill.Vigor)
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

		// if we came to town for repairs
		if needsRepair || RepairRequired() {
			if err := Repair(); err != nil {
				ctx.Logger.Warn("Failed to repair items", "error", err)
				return nil // Don't try to use portal if repair failed
			}
		}

		// If repairs are still needed, stay in town
		if needsRepair && RepairRequired() {
			return nil // Stay in town if items still need repair
		}

		// If everything is fine, return via portal
		return UsePortalInTown()
	}

	// Return nothing if none of the above checks pass
	return nil
}
