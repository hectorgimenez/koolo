package action

import "github.com/hectorgimenez/koolo/internal/config"

func (b Builder) PreRun(firstRun bool) []Action {
	if config.Config.Companion.Enabled && !config.Config.Companion.Leader {
		return []Action{
			b.RecoverCorpse(),
			b.Heal(),
		}
	}

	return []Action{
		b.EnsureStatPoints(),
		b.EnsureSkillPoints(),
		b.RecoverCorpse(),
		b.IdentifyAll(firstRun),
		b.Stash(firstRun),
		b.VendorRefill(),
		b.Heal(),
		b.ReviveMerc(),
		b.Repair(),
	}
}
