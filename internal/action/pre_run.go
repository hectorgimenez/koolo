package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/skill"
)

func (b *Builder) PreRun(firstRun bool) []Action {
	actions := []Action{
		b.DropMouseItem(),
		b.UseSkillIfBind(skill.Vigor),
		b.RecoverCorpse(),
	}

	if firstRun {
		actions = append(actions, b.Stash(firstRun))
	}

	actions = append(actions,
		b.UpdateQuestLog(),
		b.IdentifyAll(firstRun),
		b.VendorRefill(false, true),
		b.Stash(firstRun),
		b.Gamble(),
		b.Stash(false),
		b.CubeRecipes(),
	)

	if b.CharacterCfg.Game.Leveling.EnsurePointsAllocation {
		actions = append(actions,
			b.ResetStats(),
			b.EnsureStatPoints(),
			b.EnsureSkillPoints(),
		)
	}

	if b.CharacterCfg.Game.Leveling.EnsureKeyBinding {
		actions = append(actions,
			b.EnsureSkillBindings(),
		)
	}

	actions = append(actions,
		b.Heal(),
		b.ReviveMerc(),
		b.HireMerc(),
		b.Repair(),
	)

	return actions
}

func (b *Builder) InRunReturnTownRoutine() []Action {
	actions := []Action{
		b.UseSkillIfBind(skill.Vigor),
		b.ReturnTown(),
		b.RecoverCorpse(),
		b.IdentifyAll(false),
		b.VendorRefill(false, true),
		b.Stash(false),
		b.Gamble(),
		b.Stash(false),
		b.CubeRecipes(),
	}

	if b.CharacterCfg.Game.Leveling.EnsurePointsAllocation {
		actions = append(actions,
			b.EnsureStatPoints(),
			b.EnsureSkillPoints(),
		)
	}

	if b.CharacterCfg.Game.Leveling.EnsureKeyBinding {
		actions = append(actions,
			b.EnsureSkillBindings(),
		)
	}

	actions = append(actions,
		b.Heal(),
		b.ReviveMerc(),
		b.HireMerc(),
		b.Repair(),
		b.UsePortalInTown(),
	)

	return actions
}
