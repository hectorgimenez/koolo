package action

import "github.com/hectorgimenez/koolo/internal/config"

func (b *Builder) PreRun(firstRun bool) []Action {
	if config.Config.Companion.Enabled && !config.Config.Companion.Leader {
		return []Action{
			b.RecoverCorpse(),
			b.Heal(),
		}
	}

	return []Action{
		b.ResetStats(),
		b.EnsureStatPoints(),
		b.EnsureSkillPoints(),
		b.RecoverCorpse(),
		b.IdentifyAll(firstRun),
		b.Stash(firstRun),
		b.VendorRefill(false, true),
		b.Gamble(),
		b.Stash(false),
		b.EnsureSkillBindings(),
		b.Heal(),
		b.ReviveMerc(),
		b.HireMerc(),
		b.Repair(),
	}
}

func (b *Builder) InRunReturnTownRoutine() []Action {
	return []Action{
		b.ReturnTown(),
		b.EnsureStatPoints(),
		b.EnsureSkillPoints(),
		b.RecoverCorpse(),
		b.IdentifyAll(false),
		b.Stash(false),
		b.VendorRefill(false, true),
		b.Gamble(),
		b.Stash(false),
		b.EnsureSkillBindings(),
		b.Heal(),
		b.ReviveMerc(),
		b.HireMerc(),
		b.Repair(),
		b.UsePortalInTown(),
	}
}
