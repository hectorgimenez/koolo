package action

func (b *Builder) PreRun(firstRun bool) []Action {
	actions := []Action{
		b.RecoverCorpse(),
		b.UpdateQuestLog(),
		b.IdentifyAll(firstRun),
		b.Stash(firstRun),
		b.VendorRefill(false, true),
		b.Gamble(),
		b.Stash(false),
	}

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
		b.ReturnTown(),
		b.RecoverCorpse(),
		b.IdentifyAll(false),
		b.Stash(false),
		b.VendorRefill(false, true),
		b.Gamble(),
		b.Stash(false),
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
